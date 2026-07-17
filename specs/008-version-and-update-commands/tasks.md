# Tasks: Version & Update Commands

**Input**: Design documents from `/specs/008-version-and-update-commands/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project setup and initial registration

- [x] T001 Define version metadata string variables (`Version`, `CommitSHA`, `BuildTime`) in `cmd/version.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core update logic components

- [x] T002 [P] Implement semantic version string parser and comparator in `internal/update/update.go`

---

## Phase 3: User Story 1 - Check Installed Version (Priority: P1) 🎯 MVP

**Goal**: Support version command and root flags to display compiled metadata

**Independent Test**: Running `cx version` or `cx -v` prints version details and exits cleanly.

### Implementation for User Story 1

- [x] T003 [P] [US1] Create the Cobra `version` command and wire it up in `cmd/version.go`
- [x] T004 [US1] Update `cmd/root.go` to handle `-v` and `--version` flags as aliases for `cx version`

**Checkpoint**: User Story 1 is fully functional and can be tested.

---

## Phase 4: User Story 2 - Check for Updates (Priority: P2)

**Goal**: Fetch and compare local version against the latest GitHub release

**Independent Test**: Running `cx update --check` shows whether an update is available.

### Implementation for User Story 2

- [x] T005 [P] [US2] Implement GitHub latest release API client using Go's `net/http` in `internal/update/update.go`
- [x] T006 [P] [US2] Create HTTP mock tests for the updater client in `internal/update/update_test.go`
- [x] T007 [US2] Implement the `cx update --check` subcommand in `cmd/update.go`

**Checkpoint**: User Story 2 can be tested independently using mock inputs.

---

## Phase 5: User Story 3 - Perform Auto-Upgrade (Priority: P2)

**Goal**: Perform platform-safe download and replacement of the binary

**Independent Test**: Running `cx update` successfully upgrades the active binary.

### Implementation for User Story 3

- [x] T008 [P] [US3] Implement target OS/Arch asset matcher matching `cx-<os>-<arch>` in `internal/update/update.go`
- [x] T009 [P] [US3] Implement download and self-replacement rename-dance (supporting Unix/Windows lock safety) in `internal/update/update.go`
- [x] T010 [US3] Implement interactive prompt and `-y` / `--yes` flag handling for `cx update` in `cmd/update.go`

**Checkpoint**: User Story 3 is complete and ready for end-to-end upgrades.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, polish, and validation

- [x] T011 [P] Implement graceful rate-limit and network timeout handling in `internal/update/update.go`
- [x] T012 Run the end-to-end validation flows defined in `quickstart.md`
- [x] T013 Update main README.md with version and update usage instructions

---

## Dependencies & Execution Order

### Phase Dependencies

* **Setup (Phase 1)**: Can start immediately.
* **Foundational (Phase 2)**: Depends on Setup (Phase 1).
* **User Stories**:
  * **User Story 1 (P1)**: Depends on Phase 1 & 2.
  * **User Story 2 (P2)**: Depends on Phase 1 & 2.
  * **User Story 3 (P2)**: Depends on User Story 2 (T005 - GitHub client must be working to locate assets).

### Parallel Opportunities

* T002, T003, and T005 can be started in parallel.
* T006 (mock tests) can be written in parallel with T005.
* T008 (asset matcher) and T009 (rename replacement) can be written in parallel with T007.
