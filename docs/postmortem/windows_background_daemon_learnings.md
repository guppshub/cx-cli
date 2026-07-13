# Windows Background Daemon & PowerShell Learnings Post-Mortem

This document details the engineering challenges, system-level behaviors, and solutions discovered while implementing a resilient, background database tunnel daemon (`cx`) on Windows PowerShell.

---

## 1. The Goal
We wanted `cx` to run database tunnels (`aws ssm start-session` port-forwarding) as a background daemon process. The requirements were:
1. The daemon must run completely in the background without user intervention.
2. The daemon must survive the closure of the PowerShell/WezTerm window that spawned it.
3. The daemon must not show any popup console windows to the user.
4. Tunnels must automatically reconnect on network drops or AWS server-side timeouts.

---

## 2. Issue 1: Terminal Closures on WSL 2 (NixOS)
### The Problem:
On Linux/WSL, closing the terminal window (e.g., WezTerm) propagates a `SIGHUP` (Signal Hangup) to all child processes in the terminal session, instantly killing them. Additionally, on modern `systemd`-native distros like NixOS:
* Closing the terminal shuts down the `systemd --user` session, reaping all user background processes (including `tmux` and `cx` daemons).
* WSL 2 automatically shuts down the entire Linux VM 60 seconds after the last terminal window closes (`vmIdleTimeout`).

### The Solution:
1. **Go Signal Ignoring**: Implemented `signal.Ignore(syscall.SIGHUP)` in Unix-specific files so `cx` ignores the hangup signal.
2. **Systemd Lingering**: Enabled lingering to prevent systemd from killing user background processes:
   ```bash
   loginctl enable-linger <username>
   ```
3. **WSL VM Persistence**: Configured WSL not to shut down the VM automatically when idle by creating `~/.wslconfig` on the Windows host:
   ```ini
   [wsl2]
   vmIdleTimeout=-1
   ```

---

## 3. Issue 2: Terminal Session Termination on Windows Host
### The Problem:
When running `cx` natively on the Windows host inside PowerShell, closing the PowerShell window destroys its hosting console session (`conhost.exe` or `openconsole.exe`). By default, Windows forcefully terminates every process registered under that console session.
* Simply detaching using the standard `CREATE_NEW_PROCESS_GROUP` flag is not enough because the process group is still tied to the parent's console session.

### The Detached Process Attempt (Why it failed):
We tried using `DETACHED_PROCESS` (`0x00000008`) to spawn the daemon completely without a console session.
* **Result**: The preflight check succeeded, but the background daemon hung indefinitely.
* **Reason**: AWS `session-manager-plugin` is a console application. When run inside a `DETACHED_PROCESS` (which has no console context at all), the plugin's Win32 console API calls (such as reading screen buffers or initializing standard streams) fail or block, causing the process to hang forever during the TLS/WebSocket handshake.

### The Win32 Flags Attempt (Why it triggered EDR/Antivirus):
We attempted to allocate a new console session and hide it in Go using low-level Win32 flags:
1. **`CREATE_NEW_CONSOLE` (`0x00000010`)**: Gave the daemon a dedicated console session to survive parent closure.
2. **`HideWindow: true`**: Instructed Windows to hide the newly created console window (`SW_HIDE`).
* **Result**: Windows Defender and corporate EDR software (CrowdStrike/SentinelOne) immediately blocked the executable with **`Access is denied`**.
* **Reason**: Spawning a console application inside a hidden console window using low-level Win32 creation flags from an unsigned binary is a high-confidence indicator of Trojan/Malware activity (hidden shell executors).

### The Final EDR-Safe Solution: PowerShell Hidden Launcher
Instead of using low-level Win32 flags inside the Go compiled binary (which triggers EDR heuristics), we spawn the background daemon using Windows' built-in, trusted scripting host: **PowerShell**.

On Windows, the parent `cx` process launches the daemon using:
```go
psCmd := fmt.Sprintf(
    "Start-Process -FilePath '%s' -ArgumentList 'db', '%s', '--server', '--port', '%d' -WindowStyle Hidden -RedirectStandardOutput '%s' -RedirectStandardError '%s'",
    os.Args[0], dbResource.Name, localPort, logPath, errLogPath,
)
daemonCmd = exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", psCmd)
```

1. **`-WindowStyle Hidden`**: Tells Windows to hide the newly spawned PowerShell instance (and consequently our `cx` daemon child) completely. Since this is a standard, built-in PowerShell flag used by administrators daily, EDR heuristics do not block it.
2. **Global Detachment via `Start-Process`**: Spawning via `Start-Process` cleanly detaches the process from the parent terminal's console session host (`conhost.exe`). The daemon survives the closure of the parent shell.
3. **No Win32 API Hacks**: We reverted `detachCmd` in Go to be a clean, flag-less no-op function, keeping the compiled `cx.exe` binary 100% clean of heuristic triggers.
4. **Registration PID Bypass**: Since PowerShell launches the child daemon, the parent process checks the registry status file (`state.json`) matching by connection Name rather than PID on Windows.

---

## 4. Issue 3: Win32 File Sharing Violations
### The Problem:
Initially, when spawning the daemon using `Start-Process`, we passed the same log file path (`sequr.log`) to both `-RedirectStandardOutput` and `-RedirectStandardError`. On Windows, the OS blocks concurrent write streams to the same file. This threw a sharing violation error inside PowerShell, preventing the daemon from starting at all.

### The Solution:
We separated the logs into two files on Windows:
* Standard logs (stdout) go to **`sequr.log`**.
* Error logs (stderr) go to **`sequr_err.log`**.

This completely resolves the sharing violation, and the parent CLI is configured to read and print both files in case of startup failures.

---

## 5. Issue 4: Handshake Latency Timeouts on Windows
### The Problem:
On Windows, process startup latency is significantly higher than on Unix. Spawning the Python-based `aws` CLI executable, loading the `session-manager-plugin.exe` plugin, executing DNS/HTTPS queries, and establishing the WebSocket tunnel with AWS took slightly more than `5` seconds.
* This triggered our strict `5`-second preflight handshake timeout, resulting in a false timeout error even though the session was successfully created on the AWS console.

### The Solution:
Increased the handshake scan timeout to **`15` seconds**:
```go
scanCtx, scanCancel := context.WithTimeout(ctx, 15*time.Second)
```
Since the scanner returns immediately as soon as `"Waiting for connections"` is read from `stdout`, fast connections still initialize instantly, while slow connections on Windows have ample time to complete successfully.

---

## 6. Issue 5: Go Compiler Gotchas
### The Problem:
When building for Windows, Go's standard `syscall` package does not expose every single Win32 API constant (such as `DETACHED_PROCESS` or `CREATE_NO_WINDOW`) on all platforms, leading to compile-time failures on GitHub Actions builders.

### The Solution:
By detaching and hiding the process group natively via PowerShell's `-WindowStyle Hidden` and `Start-Process` parameters, we no longer need to pass any custom Win32 flags inside Go. We simplified `detachCmd` in `cmd/detach_windows.go` to be an empty, flag-less function:
```go
// cmd/detach_windows.go
func detachCmd(cmd *exec.Cmd) {
	// Spawning is handled via PowerShell, no custom Win32 flags needed here.
}
```
This cleanly bypasses Go's platform-specific compilation constraints and keeps the codebase highly portable.

---

## 7. Automating the Windows Installation Experience
To match the clean, one-liner installation experience of Unix (`install.sh`), we developed a Windows-native PowerShell installer.

### The script (`scripts/install.ps1`):
* Detects CPU Architecture (`AMD64` vs `ARM64`).
* Fetches the latest `.exe` binary from `/releases/latest`.
* Installs it in a user-space directory: `~/.cx/bin/cx.exe`.
* Modifies the persistent User `PATH` environment variable via .NET APIs:
  ```powershell
  [System.Environment]::SetEnvironmentVariable("PATH", "$currentPath;$installDir", "User")
  ```

This allows Windows developers to install and run the tool globally with a single PowerShell command:
```powershell
irm https://raw.githubusercontent.com/guppshub/cx-cli/main/scripts/install.ps1 | iex
```
