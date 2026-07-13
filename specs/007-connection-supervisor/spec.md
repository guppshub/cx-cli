# Feature Specification: Connection Supervisor

**Feature Branch**: `007-connection-supervisor`

**Created**: 2026-07-13

**Status**: Draft

**Input**: User architecture proposal: "Design and implement the Connection Supervisor owning the complete lifecycle of every long-running connection."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Self-Healing Background Tunnels (Priority: P1)

As an engineer using `cx-cli`, I want background database/cache tunnels to automatically recover and reconnect when the network drops or my laptop sleeps, so that my local database client (e.g. DBeaver, TablePlus, redis-cli) stays connected without me running commands.

**Why this priority**: Solves the primary pain point of dropped connections causing silent tunnel failures and stale status reporting.

**Independent Test**: Verify that terminating the underlying `aws ssm` process of an active background tunnel triggers an automatic restart after a configured backoff delay, restoring the connection and updating the state metadata without restarting the `cx` daemon process.

**Acceptance Scenarios**:
1. **Given** a background database tunnel is active and healthy, **When** the underlying child process (SSM tunnel) is terminated (e.g. using `kill`), **Then** the supervisor detects the process exit event immediately (without polling), logs a reconnecting state, waits for the restart policy backoff, and establishes the tunnel again.
2. **Given** the supervisor is attempting a restart, **When** the network is down and connection fails, **Then** the supervisor state moves to `Restarting`, it increments the restart count, logs the failure, and schedules the next retry using exponential backoff.
3. **Given** the supervisor reaches the maximum retry count or encounters a permanent error (e.g. invalid AWS profile/credentials), **When** a retry fails, **Then** the supervisor state transitions to `Failed` and halts further reconnection attempts to prevent API rate limiting.

---

### User Story 2 - Accurate Status & State Persistence (Priority: P1)

As an engineer, I want `cx status` to show accurate, real-time connection status (including state, port, restart count, and start times), so that I know exactly when a tunnel is healthy, restarting, or has failed.

**Why this priority**: Essential for visibility into background daemons.

**Independent Test**: Verify that running `cx status` prints a structured table showing the name, current state (`Healthy`, `Restarting`, `Failed`), local port, restart count, and start time of every managed connection.

**Acceptance Scenarios**:
1. **Given** a tunnel is starting up, **When** I run `cx status`, **Then** the state shows `Starting`.
2. **Given** a tunnel has successfully connected, **When** I run `cx status`, **Then** the state shows `Healthy`.
3. **Given** a tunnel has dropped and is backing off, **When** I run `cx status`, **Then** the state shows `Restarting` and lists the correct restart count.

---

## Edge Cases

- **Transient vs Permanent Errors**: 
  * Errors like network timeouts, broken pipes, and timeouts are classified as *transient* and trigger restarts.
  * Errors like `AccessDenied`, expired AWS SSO sessions, missing profile, or invalid instance IDs are *permanent* and immediately transition the supervisor to `Failed` (no retries).
- **Multiple Port Conflicts**: If the preferred local port becomes occupied during a restart (e.g., another process binds to port 5432 while the tunnel was down), the supervisor must attempt to bind to the same port or fail with a port occupied error if it cannot re-bind. (Usually, the daemon should hold the port, but we must handle conflict scenarios gracefully).
- **Process Group Termination on Stop**: When `Stop()` is called, the supervisor must ensure the child process and all its subprocesses are killed cleanly using process groups, avoiding orphaned processes.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST implement the `Supervisor` interface in the `internal/connection` package.
- **FR-002**: The supervisor MUST maintain a state machine with the following states: `Stopped`, `Starting`, `Healthy`, `Restarting`, `Failed`.
- **FR-003**: The supervisor MUST detect child process exits immediately using event-based execution (`cmd.Wait()`) rather than polling loops.
- **FR-004**: System MUST support pluggable `RestartPolicy` implementations (e.g., `Fixed`, `Exponential`, `NoRetry`).
- **FR-005**: System MUST classify errors into *transient* (triggers restart) and *permanent* (triggers failure).
- **FR-006**: The supervisor MUST write and update detailed connection metadata (Name, PID, Port, State, Restarts, StartedAt, LastFailure, LastRestart) in `state.json`.
- **FR-007**: The `cx status` command MUST read the supervisor metadata from `state.json` and render a formatted table.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Child process exit detection triggers a state transition in under 50ms.
- **SC-002**: Reconnection retries follow the exact timing backoff defined by the `RestartPolicy`.
- **SC-003**: Graceful stop cleans up all child processes and clears the `state.json` entry in under 200ms.
- **SC-004**: Stale processes are prevented; no orphaned `session-manager-plugin` or `aws` processes remain after the daemon exits.
