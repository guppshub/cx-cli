# Feature Specification: AWS Database Tunneling

**Feature Branch**: `003-aws-database-tunneling`

**Created**: 2026-07-11

**Status**: Draft

**Input**: User feedback: "AWS Database Tunneling vertical slice implementing native socket boundaries..."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Local Port Binding & Orchestration (Priority: P1)

As an engineer, I can start a database connection workflow so that a local port (e.g. `5432`) is bound on my local machine to forward traffic to the target remote database.

**Why this priority**: Core networking slice of the CLI.

**Independent Test**: Can be verified by running the core dialer against a mock net.Conn server and checking that data sent to `localhost:5432` is correctly forwarded.

**Acceptance Scenarios**:

1. **Given** local port `5432` is free, **When** the workflow starts, **Then** it binds to `localhost:5432` and listens for connections.
2. **Given** local port `5432` is occupied, **When** the workflow starts, **Then** it falls back to finding an available port, binds to it, and prints the chosen port.

---

### User Story 2 - AWS SSM Subprocess conn Wrapper (Priority: P1)

As an engineer, I can connect to AWS RDS instances using Session Manager without having the Go codebase directly parse the WebSocket protocol, instead using a process stream wrapper.

**Why this priority**: Unlocks connection capabilities without high engineering cost.

**Independent Test**: Verify that spawning a dummy subprocess command and wrapping it in `SubprocessConn` behaves like a standard `net.Conn` stream.

**Acceptance Scenarios**:

1. **Given** valid AWS credentials and a target database context, **When** `DialTunnel()` is called, **Then** it launches `session-manager-plugin` and returns a `net.Conn` wrapping the process's I/O pipes.
2. **Given** the `aws` CLI or `session-manager-plugin` is missing in `$PATH`, **When** the connection starts, **Then** it fails immediately with an actionable error.

---

### User Story 3 - Connection State Tracking (Priority: P2)

As an engineer, I can inspect currently running tunnels so that I know what ports are forwarded and can clean them up.

**Why this priority**: Required for state consistency and tunnel lifecycle.

**Independent Test**: Check that `state.json` has the connection metadata appended on start, and removed when the tunnel connection is closed.

**Acceptance Scenarios**:

1. **Given** a tunnel connection is successfully established, **When** the workflow saves state, **Then** the details (Local Port, Resource Name, Start Time) are added to `state.json`.
2. **Given** the tunnel connection is terminated, **When** the cleanup routine executes, **Then** the connection entry is removed from `state.json`.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Core workflow MUST define a `TunnelDialer` interface:
  ```go
  type TunnelDialer interface {
      DialTunnel(ctx context.Context, target *TunnelTarget) (net.Conn, error)
  }
  ```
- **FR-002**: System MUST bind to a local socket port on `localhost`.
- **FR-003**: System MUST execute the `aws ssm` command as a background subprocess, wrapping its stdin/stdout in a custom `net.Conn` wrapper.
- **FR-004**: System MUST verify the existence of the `aws` CLI and `session-manager-plugin` in the host's system path on startup.
- **FR-005**: System MUST record connection metadata (Active Connection ID, Local Port, Resource Name) to `state.json` on successful tunnel establishment.
- **FR-006**: System MUST clear active connections from `state.json` when the tunnel closes.

### Key Entities

- **TunnelTarget**: Struct containing `BastionInstanceID` (string), `RemoteHost` (string), and `RemotePort` (int).
- **Endpoint**: Struct containing `Host` (string) and `Port` (int).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Port forward listener successfully binds and forwards data to the remote RDS host within 500ms of CLI command start.
- **SC-002**: Active connection metadata is correctly saved in `state.json` on tunnel open and removed on tunnel close.

## Assumptions

- The target AWS bastion host is configured to allow SSM Session Manager connections.
- The `session-manager-plugin` is installed on the user's host machine.
- Temporary AWS credentials are valid or resolved prior to starting the tunnel connection.
