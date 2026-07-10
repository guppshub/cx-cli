# Tasks: Workspace Management

**Input**: Design documents from `specs/002-workspace-management/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/workspace-api.md

**Tests**: Unit tests are required to satisfy the spec quality and acceptance criteria. Tests must be written and run before core implementations.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Update configuration models to support active workspace tracking.

- [X] T001 Add Current string yaml:"current" to Config struct in internal/config/config.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Initialize the workspace package and structs.

- [X] T002 Implement target structs (Workspace, WorkspaceSummary, Manager) in internal/workspace/workspace.go

---

## Phase 3: User Story 1 - Add and List Workspaces (Priority: P1)

**Goal**: Support adding new workspaces and listing them alphabetically.

**Independent Test**: Run `go test -v -run TestAddAndList ./internal/workspace/...`

### Tests for User Story 1
- [X] T003 [P] [US1] Write unit test verifying workspace creation, duplicates rejection, and alphabetical listing in internal/workspace/workspace_test.go

### Implementation for User Story 1
- [X] T004 [US1] Implement Add() function checking for duplicate context names in internal/workspace/workspace.go
- [X] T005 [US1] Implement List() function returning deterministically sorted workspace lists in internal/workspace/workspace.go

---

## Phase 4: User Story 2 - Select and Retrieve Active Workspace (Priority: P1)

**Goal**: Switch active workspace and retrieve current workspace.

**Independent Test**: Run `go test -v -run TestUseAndCurrent ./internal/workspace/...`

### Tests for User Story 2
- [X] T006 [P] [US2] Write unit test for selecting workspace and retrieving current active workspace in internal/workspace/workspace_test.go

### Implementation for User Story 2
- [X] T007 [US2] Implement Use(), Current(), and Get() functions verifying target workspace exists and updating Config.Current in internal/workspace/workspace.go

---

## Phase 5: User Story 3 - Rename and Delete Workspaces with Protection (Priority: P2)

**Goal**: Rename and delete workspaces with protection rules.

**Independent Test**: Run `go test -v -run TestRenameAndDelete ./internal/workspace/...`

### Tests for User Story 3
- [X] T008 [P] [US3] Write unit test for renaming active workspaces, deleting workspaces, and active workspace deletion rejection in internal/workspace/workspace_test.go

### Implementation for User Story 3
- [X] T009 [US3] Implement Delete() with active workspace protection checks in internal/workspace/workspace.go
- [X] T010 [US3] Implement Rename() with automatic active pointer updating logic in internal/workspace/workspace.go

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end verification, linting, and cleanup.

- [X] T011 Run verification script specs/002-workspace-management/scratch/verify.go
- [X] T012 Run go fmt and golangci-lint run on the internal/workspace/ package

---

## Dependencies & Execution Order

### Phase Dependencies
- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on T001.
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2).
- **User Story 2 (Phase 4)**: Depends on User Story 1 (Phase 3).
- **User Story 3 (Phase 5)**: Depends on User Story 1 (Phase 3).
- **Polish (Phase 6)**: Depends on all prior phases.

### Parallel Opportunities
- Test files `T003`, `T006`, and `T008` in `internal/workspace/workspace_test.go` can be developed in parallel with one another.
- Implementation of `T009` (Delete) and `T010` (Rename) can proceed in parallel once User Story 1 and 2 are complete.

---

## Implementation Strategy

### MVP First (User Story 1 & 2)
1. Complete Setup and Foundational phases.
2. Complete User Story 1 (Add and List).
3. Complete User Story 2 (Select and Retrieve).
4. Run `go test -v ./internal/workspace/...` to validate the MVP.
5. Proceed to User Story 3 (Rename and Delete with protection) and Polish phases.
