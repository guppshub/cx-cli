# Feature Specification: CLI Integration

**Feature Branch**: `004-cli-integration`

**Created**: 2026-07-11

**Status**: Draft

**Input**: User feedback: "Proceed to 004 spec for CLI Integration..."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Switch Active Workspace (Priority: P1)

As an engineer, I can run `cx use <workspace>` to quickly switch my active environment context, so that subsequent operations execute against that target environment.

**Why this priority**: Required for context-awareness.

**Independent Test**: Verify that running `cx use staging` updates the `current` active workspace key inside `config.yaml` to `staging` and outputs a confirmation message.

**Acceptance Scenarios**:

1. **Given** the workspace `staging` exists in the configuration, **When** I run `cx use staging`, **Then** the CLI prints `Now using workspace "staging"` and writes the update to `config.yaml`.
2. **Given** the workspace `production` does not exist, **When** I run `cx use production`, **Then** the CLI prints `workspace "production" does not exist` and exits with code 1.

---

### User Story 2 - Connect to Database via CLI (Priority: P1)

As an engineer, I can run `cx db <resource>` to open a secure tunnel to my target database resource, forwarding it to a local port.

**Why this priority**: Delivers the primary database operational workflow.

**Independent Test**: Verify that running `cx db mercury` resolves the target RDS database from config, starts the port forwarding loop, and remains in the foreground until interrupted.

**Acceptance Scenarios**:

1. **Given** no active workspace is selected, **When** I run `cx db mercury`, **Then** the CLI prints `no active workspace selected. Use "cx use <workspace>" first.` and exits with code 1.
2. **Given** the active workspace has database `mercury` configured, **When** I run `cx db mercury`, **Then** the CLI prints connection logs, binds to the local port (e.g. `5432`), and stays open in the foreground.
3. **Given** the tunnel is running, **When** I press `Ctrl+C`, **Then** the CLI terminates the tunnel process, clears state entries, and exits cleanly.

---

### Edge Cases

- **Missing CLI Dependencies**: If `aws` or `session-manager-plugin` are missing, `cx db` should print a friendly error message detailing how to install them, rather than crashing with a raw stack trace.
- **Port Already Bound**: If the default local port is in use, the CLI should print a warning and report the fallback random port it bound to (e.g. `Port 5432 is in use, falling back to port 59871`).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST integrate with Cobra CLI framework to provide the base `cx` root commands.
- **FR-002**: System MUST support command `cx use <workspace>` which updates `current` in `config.yaml`.
- **FR-003**: System MUST support command `cx db <resource>` accepting a `--port` (or `-p`) override flag.
- **FR-004**: System MUST parse resource details (Bastion ID, Remote Endpoint, Local Port) from the active workspace's raw `workspaces` map in `config.yaml`.
- **FR-005**: System MUST trap OS interrupts (`SIGINT`, `SIGTERM`) during `cx db` to close connection listeners and state records before exiting.
- **FR-006**: System MUST output clean, user-friendly errors on CLI stdout/stderr without exposing internal raw Go stack traces.

### Key Entities

- **ResourceDefinition**: Schema definition parsed from raw workspace maps:
  ```go
  type DatabaseResource struct {
      Name              string `yaml:"name"`
      Engine            string `yaml:"engine"`
      Endpoint          string `yaml:"endpoint"`
      Port              int    `yaml:"port"`
      LocalPort         int    `yaml:"local_port"`
      BastionInstanceID string `yaml:"bastion_instance_id"`
  }
  ```

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `cx use` updates the active configuration file immediately (under 50ms).
- **SC-002**: Running `cx db` successfully binds a local port and prints the connection status within 500ms.
- **SC-003**: Graceful shutdown on `Ctrl+C` cleans up listener sockets and clears `state.json` entries in under 100ms.

## Assumptions

- Cobra is the standard CLI framework used for command-line argument and flag parsing.
- The `config.yaml` file conforms to the parsed resource schema block structure under `workspaces.<name>.resources.databases`.
