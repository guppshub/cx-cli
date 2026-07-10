# Tasks: Configuration Foundation

**Input**: Design documents from `specs/001-config-foundation/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/config-api.md

**Tests**: Unit tests are required to satisfy the spec quality and acceptance criteria. Tests must be written and run before core implementations.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify dependencies and project module initialization.

- [X] T001 Verify and add gopkg.in/yaml.v3 dependency in go.mod

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Initialize the config package and target structs.

- [X] T002 Implement target structs (Config, Context, AWSConfig, Resources, ConnectionMetadata, State) in internal/config/config.go

---

## Phase 3: User Story 1 - Multi-Platform Configuration Resolution & Loading (Priority: P1)

**Goal**: Automatically resolve configuration paths and load configuration or return defaults.

**Independent Test**: Run `go test -v -run TestLoadDefault ./internal/config/...`

### Tests for User Story 1
- [X] T003 [P] [US1] Write loader and resolution tests in internal/config/config_test.go

### Implementation for User Story 1
- [X] T004 [US1] Implement path resolution functions Path() and StatePath() in internal/config/config.go
- [X] T005 [US1] Implement Default() configuration builder and Load() file-reading logic in internal/config/config.go

---

## Phase 4: User Story 2 - Configuration Persistence & Directory Provisioning (Priority: P1)

**Goal**: Atomically save configuration updates and automatically create parent directories.

**Independent Test**: Run `go test -v -run TestSave ./internal/config/...`

### Tests for User Story 2
- [X] T006 [P] [US2] Write unit tests for configuration saving, parent directory provisioning, and atomic file replacement in internal/config/config_test.go

### Implementation for User Story 2
- [X] T007 [US2] Implement Save() function with atomic temporary write-and-rename and directory creation logic in internal/config/config.go

---

## Phase 5: User Story 3 - Configuration Schema & Version Validation (Priority: P2)

**Goal**: Validate configuration for version checks and structural validity.

**Independent Test**: Run `go test -v -run TestValidate ./internal/config/...`

### Tests for User Story 3
- [X] T008 [P] [US3] Write unit tests for schema validation (invalid version strings, duplicate context names) in internal/config/config_test.go

### Implementation for User Story 3
- [X] T009 [US3] Implement Validate() function with structural and version validation logic in internal/config/config.go

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end verification, linting, and cleanup.

- [X] T010 Run local verification scenarios using specs/001-config-foundation/scratch/verify.go
- [X] T011 Run formatting (go fmt) and linters (golangci-lint run) in internal/config/config.go

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
- Test files `T003`, `T006`, and `T008` in `internal/config/config_test.go` can be developed in parallel with one another.
- Implementation of `T007` (Save) and `T009` (Validate) can proceed in parallel once the core loader `T005` is complete.

---

## Implementation Strategy

### MVP First (User Story 1 & 2)
1. Complete Setup and Foundational phases.
2. Complete User Story 1 (Load & Resolve paths).
3. Complete User Story 2 (Save atomically).
4. Run `go test -v ./internal/config/...` to validate the MVP.
5. Proceed to User Story 3 (Validation schema) and Polish phases.
