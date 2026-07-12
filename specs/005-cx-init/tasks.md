# Tasks: cx init

**Input**: Design documents from `/specs/005-cx-init/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure checks

- [x] T001 Verify project readiness and check Go module path

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

*No blocking foundational database/API setups are required for this feature.*

---

## Phase 3: User Story 1 - Initialize configuration for the first time (Priority: P1) 🎯 MVP

**Goal**: Automatically create configuration directory and default starter YAML file on `cx init`.

**Independent Test**: Remove any existing configuration, run `./cx init`, verify file is created with default template, and run `./cx use dev` to verify workspace loading.

### Implementation for User Story 1

- [x] T002 [P] [US1] Define starter configuration template string in `internal/config/template.go`
- [x] T003 [US1] Implement Cobra command init in `cmd/init.go` to create directory and write starter configuration
- [x] T004 [US1] Register init subcommand in `cmd/root.go`

**Checkpoint**: At this point, running `cx init` should successfully create a fresh configuration and dev workspace works.

---

## Phase 4: User Story 2 - Prevent accidental overwrite (Priority: P1)

**Goal**: Protect existing configuration files from overwrite unless `--force` is specified.

**Independent Test**: Verify running `./cx init` when config exists fails with code 1, and running `./cx init --force` succeeds and overwrites it.

### Implementation for User Story 2

- [x] T005 [US2] Implement file existence checks and exit error code 1 in `cmd/init.go` when config file already exists
- [x] T006 [US2] Add `--force` (or `-f`) flag to `cmd/init.go` to override check and force-overwrite file

**Checkpoint**: User Stories 1 and 2 should both work independently.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation and codebase cleanliness.

- [x] T007 Run validation scenarios in `specs/005-cx-init/quickstart.md`
- [x] T008 Run gofmt and golangci-lint run on the entire workspace

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **User Story 1 (Phase 3)**: Depends on Setup completion.
- **User Story 2 (Phase 4)**: Depends on User Story 1 completion.
- **Polish (Phase 5)**: Depends on User Stories 1 and 2 completion.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 3: User Story 1
3. **STOP and VALIDATE**: Run Scenario A in `quickstart.md`.
4. Proceed to User Story 2.
