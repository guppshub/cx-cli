# Tasks: AWS Database Tunneling

**Input**: Design documents from `specs/003-aws-database-tunneling/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/tunnel-api.md

**Tests**: Unit tests are required to satisfy the spec quality and acceptance criteria. Tests must be written and run before core implementations.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create project directory structures and confirm Go module initialization.

- [X] T001 Initialize package structures for internal/provider/aws/ and internal/workflow/db/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define the dialer interfaces and target data structures.

- [X] T002 Implement TunnelTarget, Endpoint structs and TunnelDialer interface in internal/workflow/dialer.go

---

## Phase 3: User Story 1 - Local Port Binding & Orchestration (Priority: P1)

**Goal**: Bind to local TCP sockets on localhost and forward traffic through the dialer connection.

**Independent Test**: Run `go test -v -run TestLocalBinding ./internal/workflow/db/...`

### Tests for User Story 1
- [X] T003 [P] [US1] Write unit tests for local TCP port listener binding, fallback port finding, and connection forwarding in internal/workflow/db/db_test.go

### Implementation for User Story 1
- [X] T004 [US1] Implement TCP port forwarding listener loop with fallback binding and io.Copy pipe copy routines in internal/workflow/db/db.go

---

## Phase 4: User Story 2 - AWS SSM Subprocess conn Wrapper (Priority: P1)

**Goal**: Execute session-manager-plugin in background and wrap process streams in a net.Conn.

**Independent Test**: Run `go test -v -run TestSubprocessConn ./internal/provider/aws/...`

### Tests for User Story 2
- [X] T005 [P] [US2] Write unit tests for SubprocessConn wrapper, custom Prompter prompts, and executable checks in internal/provider/aws/aws_test.go

### Implementation for User Story 2
- [X] T006 [US2] Implement ProcessConn net.Conn wrapper and dependency check LookPath commands in internal/provider/aws/aws.go
- [X] T007 [US2] Implement Provider DialTunnel and EnsureCredentials with Prompter callbacks in internal/provider/aws/aws.go

---

## Phase 5: User Story 3 - Connection State Tracking (Priority: P2)

**Goal**: Record active database tunnel connections to state.json.

**Independent Test**: Run `go test -v -run TestStateTracking ./internal/workflow/db/...`

### Tests for User Story 3
- [X] T008 [P] [US3] Write unit test verifying that active connection metadata is correctly saved in state.json on start and removed on close in internal/workflow/db/db_test.go

### Implementation for User Story 3
- [X] T009 [US3] Integrate state package in db.Controller to append connection records on connect and delete them on disconnect in internal/workflow/db/db.go

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end verification, linting, and cleanup.

- [X] T010 Run validation actions subprocess-mock and listen using specs/003-aws-database-tunneling/scratch/verify.go
- [X] T011 Run go fmt and golangci-lint run on the entire workspace

---

## Dependencies & Execution Order

### Phase Dependencies
- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Phase 1.
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2).
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2).
- **User Story 3 (Phase 5)**: Depends on User Story 1 (Phase 3) and User Story 2 (Phase 4).
- **Polish (Phase 6)**: Depends on all prior phases.

### Parallel Opportunities
- Test files `T003` and `T005` can be developed in parallel since they reside in separate package directories (`workflow/db` and `provider/aws`).
- `T007` (AWS Provider) and `T004` (Port Forwarder) can proceed in parallel once the foundational structures (`T002`) are written.

---

## Implementation Strategy

### MVP First (User Story 1 & 2)
1. Complete Setup and Foundational phases.
2. Complete User Story 1 (Local port forwarder loop).
3. Complete User Story 2 (Subprocess connection wrapper).
4. Run `go test -v ./internal/...` to validate the MVP.
5. Proceed to User Story 3 (Connection state tracking) and Polish phases.
