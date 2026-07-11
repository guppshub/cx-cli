# Technical Report: Evolution of secure port-forwarding and refactoring

This document details the engineering journey of refactoring the database and cache tunneling workflows in `cx-cli` over 10+ iterations. It documents the root causes of connection hangs, protocol corruption, port conflicts, and explains the final native architecture that resolved these issues.

---

## 1. Executive Summary & Root Causes

During development, the CLI suffered from four distinct failure modes:
1. **Connection Hangs:** Client connections (e.g. `psql` or `redis-cli`) would hang indefinitely during connection handshakes.
2. **Protocol Corruption:** Strict clients (like `redis-cli`) failed immediately with `Protocol error, got "P" as reply type byte` or `"\n"`.
3. **Failing to Bind to Target Port:** The daemon would unexpectedly bind to random ports (e.g. `58614`) instead of `6379`, causing subsequent client connections to fail with `Connection refused`.
4. **Orphaned Tunnels:** Tunnels remained active and clients could still query Redis/Postgres even after executing `cx disconnect`.

These errors were traced back to four root causes:
* **Root Cause 1 (SSM stdio mismatch):** The initial implementation ran a custom Go TCP socket listener (`tunnel.Controller`) which attempted to proxy bytes by reading and writing to the `stdin`/`stdout` of the `session-manager-plugin` process. However, the SSM port-forwarding plugin does not pipe target network bytes over stdio (standard input/output is reserved strictly for plugin logs and internal control messages). The actual tunnel operates exclusively on the local TCP socket opened by the plugin itself.
* **Root Cause 2 (Early scanner exit):** To detect startup success, the stdout scanner matched `"Starting session"`. Because this is the first line of the AWS logs, the scanner exited early, leaving the remaining logs (`Port 6379 opened...` and `Waiting for connections...`) in the stream buffer. When the client connected, it received these logs, saw `"P"` (from `Port...`), and crashed.
* **Root Cause 3 (TIME_WAIT Port Contention):** The pre-flight handshake ran the verification tunnel on the target port (e.g. `6379`). When closed, the OS kept the socket in `TIME_WAIT` for a brief period. The background daemon started immediately after and found the port busy, falling back to a random port.
* **Root Cause 4 (Orphaned Child Subprocesses):** The Go CLI spawns the `aws` CLI process, which in turn launches the `session-manager-plugin` binary. When the Go process sent `SIGINT`/`Kill()` to the parent `aws` process, the child `session-manager-plugin` process was orphaned by the OS rather than terminated, continuing to listen on the port.

---

## 2. Phase-by-Phase Code Evolution

### Phase 1: The Custom Stdio Proxy (Broken)
The initial codebase attempted to run a Go proxy listener that intercepted traffic and piped it over standard input/output.

```
[redis-cli] ---> [Go TCP Listener (6379)] ---> [Piped stdin/stdout] ---> [session-manager-plugin] (Hangs)
```
* **Why it failed:** Standard input/output did not route database/Redis bytes. It caused infinite hangs or immediately flushed local plugin logs back to the client, leading to protocol mismatches.

### Phase 2: Resolving the Hangs and Protocol Corruption
To prevent the client from crashing on SSM startup logs, we modified the stdout reading behavior:
1. **Removed the log replay:** We stopped using `io.MultiReader` to prepend the startup logs back to the client stream.
2. **Scanner Alignment:** We changed the scanner keyword check in `aws.go` to match **only** on `"Waiting for connections"`:
   ```go
   if strings.Contains(line, "Waiting for connections") {
       success = true
       errChan <- nil
       return
   }
   ```
* **Why it worked:** Waiting for `"Waiting for connections"` ensures that the scanner reads and consumes all initial log output lines from the buffer. The stream is left 100% clean when the client starts communicating.

### Phase 3: Resolving Port Conflicts (The Handshake Isolation)
To prevent the parent pre-flight handshake from blocking the port for the daemon, we isolated the handshake's port binding:
* **Handshake on Port `0`:** In `internal/connection/manager.go`, we forced the pre-flight check to run its trial connection on a random port (`0`):
   ```go
   func (m *Manager) PreflightHandshake(ctx context.Context, dialer tunnel.Dialer, target *tunnel.Target) error {
       handshakeTarget := *target
       handshakeTarget.PreferredLocalPort = 0 // Never conflict with target port
       ...
   }
   ```
* **Why it worked:** The pre-flight handshake tests AWS credentials, MFA, and bastion-to-destination routing on a temporary random port. The target port (e.g. `6379` or `5432`) remains untouched, allowing the daemon to bind to it instantly and natively.

### Phase 4: Resolving Orphaned Processes (Process Group Termination)
To prevent child processes from remaining alive when the parent exits:
1. **Process Group Leader:** On Unix, we configure the command to launch in a new process group:
   ```go
   cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
   ```
2. **Process Group Kill:** Instead of killing the parent PID using `cmd.Process.Kill()`, we send `SIGKILL` directly to the process group using the negative PID (`-cmd.Process.Pid`):
   ```go
   _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
   ```
   *Why it worked:* By targeting `-Pid`, the OS delivers the signal to the entire process group. This cleanly terminates `session-manager-plugin` even if the parent `aws` command has already exited. (On Windows, we run `taskkill /T /F` to achieve the same result).

---

## 3. The Final Native Architecture (Why it Works)

We discarded the custom Go socket listener (`tunnel.Controller`) entirely for standard tunneling. 

Instead of proxying bytes in Go, we let the `session-manager-plugin` bind directly to the requested local port natively. The Go CLI acts purely as a **supervisor/process manager**:

```
[redis-cli] ───────────────────────(Direct TCP: 6379)────────────────────────> [session-manager-plugin]
                                                                                        │
                                                                                    (SSM Pipe)
                                                                                        ▼
[cx-cli (Go Supervisor)] ──(Launches/Monitors process via PID)──> [bastion host]
```

### Detailed Flow of the Final Code:
1. **Pre-flight Handshake:** Parent CLI calls `PreflightHandshake` using local port `0`. If AWS STS credentials, bastion ID, or remote host DNS resolution fails, the process exits within 1.5s, throwing a clean CLI error.
2. **Daemonization:** If successful, the parent CLI exits and spawns the background daemon with the target port (e.g., `6379`).
3. **Native Port Binding:** The daemon launches the SSM plugin with `--parameters localPortNumber=6379`. The SSM plugin opens the port on the loopback interface and prints `"Waiting for connections"`.
4. **State Registry:** The daemon reads `"Waiting for connections"`, registers its PID and port in `state.json`, and runs silently.
5. **Direct Connect:** The database/Redis clients connect directly to the port opened by the SSM plugin. Traffic is forwarded natively with **zero Go overhead** and **zero protocol corruption**.
6. **Graceful Disconnection:** When `disconnect` is run, it sends a `SIGINT` to the daemon. The daemon's deferred functions run, killing the SSM process and cleanly deregistering the connection from `state.json`.

---

## 4. Key Learnings & Architecture Patterns

### Generalize Downward, Keep Resources Concrete
We avoided the common Go anti-pattern of creating a premature generic `Resource` interface. Trying to unify databases, Redis, ECS, and OpenSearch under one interface causes type fighting. 
Instead:
* We kept resource parsing concrete (e.g. `resource.ResolveDatabase`, `resource.ResolveRedis`).
* They all map to a simple, shared **`tunnel.Target`** struct.
* All downstream layers (`connection.Manager`, `aws.Provider`) consume this standard target.

### Separation of Data Plane and Control Plane
* **`internal/tunnel`:** Purely represents the network target definitions and dialer interfaces.
* **`internal/connection`:** Represents the control plane. It owns process liveness checks, state registration, pre-flight handshake timeouts, and signal handling.
