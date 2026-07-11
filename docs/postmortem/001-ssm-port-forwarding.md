# Postmortem: AWS SSM Port-Forwarding Architecture & Active Verification

* **Date:** 2026-07-12
* **Scope:** CLI Database/Cache Secure Tunneling

---

## 1. What Were the Mistakes?

During the early iterations of building the database and cache port-forwarding workflows, several architectural assumptions proved incorrect, leading to hangs, protocol corruption, and orphaned background processes:

### A. The Stdio Proxy Assumption
* **Mistake:** We assumed the AWS `session-manager-plugin` could pipe raw network target bytes over standard I/O (`stdin`/`stdout`). We built a custom Go TCP proxy socket (`tunnel.Controller`) to read and write network traffic to these pipes.
* **Reality:** The port-forwarding document of `session-manager-plugin` does not use stdio for client data forwarding (stdio is strictly used for logs and control commands). Instead, it runs as a local TCP server that binds to a port and handles traffic natively. Proxying bytes over stdio resulted in absolute byte loss and infinite hangs.

### B. Brittle Log Parsing for Connection State
* **Mistake:** We checked `stdout` for specific AWS log strings (`"Starting session"` or `"Waiting for connections"`) to detect connection success. 
* **Reality:** 
  1. This was brittle. If AWS alters the log output of the `session-manager-plugin` in a future release, the check fails.
  2. Because `"Starting session"` is printed on the very first line of the AWS logs (before it contacts the bastion or resolves the database host), the scanner exited early. This left the remaining startup logs (`"Port 6379 opened..."`) in the buffer, corrupting the subsequent protocol stream. Strict clients like `redis-cli` received these text bytes instead of protocol bytes and crashed immediately.

### C. TIME_WAIT Socket Contention
* **Mistake:** The pre-flight check in the parent process verified connectivity by starting the SSM session on the desired target port (e.g. `6379`).
* **Reality:** When the pre-flight process exited, the OS held the port in a `TIME_WAIT` socket cleanup state. When the background daemon immediately started a millisecond later, it found the port busy and fell back to a random port, causing the client to get `Connection refused` on `6379`.

### D. Orphaned Subprocesses
* **Mistake:** We assumed calling `cmd.Process.Kill()` on the `aws` process would kill all of its subprocesses.
* **Reality:** The `aws` CLI spawns `session-manager-plugin` as a child process. When `aws` was killed, `session-manager-plugin` was orphaned, remaining active in the background and locking the local port.

---

## 2. What We Improved (The Fixes)

We refactored the architecture to align with Go's standard library conventions and AWS SSM's native design:

### A. Native Port Binding & Supervisor Model
We removed the custom Go TCP socket listener (`tunnel.Controller`) and proxying logic. Instead, we let the `session-manager-plugin` bind directly to the target port natively (`localPortNumber=6379` or `5432`). The Go CLI now acts as a supervisor: it spawns the process, runs pings, writes the metadata state, and terminates it when requested.

### B. Active, Protocol-Level Verification
Instead of reading stdout logs to determine connection success, we implemented native, lightweight protocol pings in **[verifier.go](file:///Users/shubham/_Coding/github/cx-cli/internal/connection/verifier.go)**:
* **Redis:** Sends RESP `PING` (`*1\r\n$4\r\nPING\r\n`) and checks for a valid RESP response prefix.
* **PostgreSQL:** Sends a standard `SSLRequest` packet and asserts that the database replies with `S` (Supports SSL) or `N` (Does not support SSL).
* **MySQL:** Validates the immediate server handshake packet.
* **HTTP/OpenSearch:** Sends a standard HTTP request and checks for an `HTTP/1.` response prefix.

### C. Handshake Port Isolation
We configured the pre-flight connection handshake to always run its check on a random port (`0`). This validates credentials, MFA, and bastion-to-destination routing without ever touching or blocking the user's preferred port.

### D. Process Group Termination
We grouped the process and its children using Process Groups (`Setpgid: true` on Unix and `CREATE_NEW_PROCESS_GROUP` on Windows) and kill the entire group via a negative PID signal (`-cmd.Process.Pid`) on close, ensuring `session-manager-plugin` is terminated gracefully.

---

## 3. Key Learnings

1. **Active Validation is Superior to Passive Validation:** Checking log outputs (passive) is vulnerable to environmental changes. Actively attempting to connect and speak the protocol (active) verifies the entire network path, authentication, and database state.
2. **Generalize Downward, Keep Resources Concrete:** Avoid premature abstractions. Concrete configuration resolvers (databases, caches) should map themselves to a unified target representation (`tunnel.Target`). Everything downstream can then handle them generically.
3. **Understand Subprocess Ownership:** Child processes spawned by Go commands must be managed via Process Groups (Unix) or Job Objects (Windows) to prevent zombie processes and port leaks.
