# Tasks: Connection Supervisor

**Input**: Design documents from `/specs/007-connection-supervisor/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify project layout and prepare for the refactoring.

- [X] T001 Verify project readiness and check Go module path

---

## Phase 2: Foundational (Blocking Prerequisites)

*No database setups are required for this pure-CLI change.*

---

## Phase 3: User Story 1 - Self-Healing Background Tunnels (Priority: P1) 🎯 MVP

**Goal**: Implement the Core Supervisor, Connection, and Restart Policy abstractions, enabling automatic retries on connection drops.

**Independent Test**: Verify that simulating a process kill on the child SSM tunnel triggers an immediate restart matching the configured backoff delays.

### Implementation for User Story 1

- [X] T002 [P] [US1] Define Connection, Dialer, and RestartPolicy interfaces in `internal/connection/connection.go` and `internal/connection/restart.go`
- [X] T003 [P] [US1] Implement FixedBackoff and NoRetry policies in `internal/connection/restart.go`
- [X] T004 [US1] Implement AWS DialWrapper in `internal/provider/aws/aws.go` satisfying `connection.Dialer` and returning a process-wrapped `Connection`
- [X] T005 [US1] Implement Supervisor core logic (Start/Stop/State/Wait) in `internal/connection/supervisor.go` with event-based child monitoring and error classification

**Checkpoint**: Core supervisor state machine and reconnect loop work with mock connections.

---

## Phase 4: User Story 2 - Accurate Status & State Persistence (Priority: P1)

**Goal**: Persist supervisor metadata and expose connection details in `cx status`.

**Independent Test**: Kill the connection child process, run `cx status` immediately, verify the state is `Restarting` and restarts counter increases, and verify graceful stops cleanup `state.json`.

### Implementation for User Story 2

- [X] T006 [P] [US2] Update ConnectionMetadata struct in `internal/state/state.go` to include State, Restarts, LastFailure, and LastRestart
- [X] T007 [US2] Integrate state manager updates in `internal/connection/supervisor.go` to persist metadata on state changes
- [X] T008 [US2] Refactor `cmd/db.go` and `cmd/redis.go` to utilize the supervisor for both foreground and background daemon executions
- [X] T009 [US2] Update `cmd/status.go` to render the rich metadata columns

**Checkpoint**: Tunnels can be managed, restarted, and accurately reported globally.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Add unit tests, validate end-to-end scenarios, and ensure codebase cleanliness.

- [X] T010 Implement unit tests in `internal/connection/supervisor_test.go` verifying state machine and policy delays
- [X] T011 Run manual validation scenarios in `specs/007-connection-supervisor/quickstart.md`
- [X] T012 Run gofmt and golangci-lint run on the entire workspace

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **User Story 1 (Phase 3)**: Depends on Setup.
- **User Story 2 (Phase 4)**: Depends on User Story 1.
- **Polish (Phase 5)**: Depends on User Stories 1 and 2.
