# Quickstart: Validating the Connection Supervisor

This document defines manual validation scenarios to verify the self-healing and lifecycle behaviors of the Connection Supervisor.

---

## 1. Scenario A: Automatic Recovery (Transient Dropout)

Verify that the supervisor automatically heals connections when the child process is terminated.

1. **Establish a database tunnel**:
   ```bash
   cx db mercury
   ```
2. **Verify status is Healthy**:
   ```bash
   cx status
   ```
   *Expected Output*: Displays state `Healthy` on port `5432` with `0` restarts.
3. **Simulate a network drop by killing the child SSM tunnel process**:
   Find the PID of the underlying `aws ssm` process:
   ```bash
   ps aux | grep "aws ssm"
   ```
   Kill that child PID (e.g. `kill -9 <child-pid>`).
4. **Instantly check status**:
   ```bash
   cx status
   ```
   *Expected Output*: State transitions to `Restarting` showing `restarts` as `1`.
5. **Verify automatic recovery**:
   Wait 5 seconds, then run:
   ```bash
   cx status
   ```
   *Expected Output*: State returns to `Healthy`, with `restarts` showing `1` and a new `Last Restart` time.

---

## 2. Scenario B: Permanent Failure Handling

Verify that permanent errors (like invalid credentials) do not cause infinite loops.

1. **Simulate credentials expiration**:
   Locally change your AWS credentials or select an invalid profile in `config.yaml`.
2. **Kill the active SSM process**:
   ```bash
   kill -9 <child-pid>
   ```
3. **Check status**:
   ```bash
   cx status
   ```
   *Expected Output*: After the retry fails with an auth error, the state transitions to `Failed`, halts retries, and displays the exact authentication error under the `Last Failure` metadata field.

---

## 3. Scenario C: Graceful Disconnect

Verify that the supervisor cleans up all state and kills the process group cleanly.

1. **Run disconnect**:
   ```bash
   cx disconnect
   ```
2. **Verify status is empty**:
   ```bash
   cx status
   ```
   *Expected Output*: Shows no active connections.
3. **Verify no orphaned processes**:
   Ensure no `aws` or `session-manager-plugin` processes remain:
   ```bash
   ps aux | grep "session-manager-plugin"
   ```
   *(Should return nothing except your grep command).*
