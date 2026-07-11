# Tasks: CLI Integration

**Input**: Design documents from `specs/004-cli-integration/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli-api.md

**Tests**: Unit tests are required to satisfy the spec quality and acceptance criteria. Tests must be written and run before core implementations.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create project directory structures and confirm Go module initialization.

- [X] T001 Initialize package structures for cmd/ and internal/resource/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define the resource database structures.

- [X] T002 Implement DatabaseResource struct in internal/resource/database.go

---

## Phase 3: User Story 1 - Switch Active Workspace (Priority: P1)

**Goal**: Support switching active workspaces via configuration updates and database resource resolving.

**Independent Test**: Run `go test -v ./internal/resource/...`

### Tests for User Story 1
- [X] T003 [P] [US1] Write unit tests for ResolveDatabase yaml parser in internal/resource/database_test.go

### Implementation for User Story 1
- [X] T004 [US1] Implement ResolveDatabase parsing helper in internal/resource/database.go and use command in cmd/use.go

---

## Phase 4: User Story 2 - Connect to Database via CLI (Priority: P1)

**Goal**: Connect to database resource using the db controller and handle signal interruptions cleanly.

**Independent Test**: Run manual CLI workflow verification using go run main.go

### Implementation for User Story 2
- [X] T005 [US2] Implement Cobra command db in cmd/db.go resolving database details and initiating EnsureCredentials and Start controller
- [X] T006 [US2] Integrate signal trap NotifyContext in cmd/db.go to gracefully terminate database listeners on OS interrupts
- [X] T007 [US2] Wire use and db subcommands into Cobra root Command in cmd/root.go

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end verification, linting, and cleanup.

- [X] T008 Run validation actions listed in specs/004-cli-integration/quickstart.md
- [X] T009 Run go fmt and golangci-lint run on the entire workspace

---

## Dependencies & Execution Order

### Phase Dependencies
- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Phase 1.
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2).
- **User Story 2 (Phase 4)**: Depends on User Story 1 (Phase 3).
- **Polish (Phase 5)**: Depends on all prior phases.

### Parallel Opportunities
- Test creation task `T003` can be developed concurrently with `T001`/`T002`.
- Command wiring `T007` can be set up in parallel with `T005` command definition.

---

## Implementation Strategy

### MVP First (User Story 1 & 2)
1. Complete Setup and Foundational phases.
2. Complete User Story 1 (Switch Active Workspace).
3. Complete User Story 2 (Database CLI Command).
4. Verify execution end-to-end using manual testing.
5. Proceed to Polish phase.
