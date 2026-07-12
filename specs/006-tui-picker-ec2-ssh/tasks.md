# Tasks: TUI Resource Picker & EC2 SSH

**Input**: Design documents from `/specs/006-tui-picker-ec2-ssh/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add new UI library dependencies to the project config

- [x] T001 Add Bubble Tea, Bubbles, and Lip Gloss dependencies to `go.mod`

---

## Phase 2: Foundational (Blocking Prerequisites)

*No database or core configuration changes are required before beginning User Story 1.*

---

## Phase 3: User Story 1 - Reusable TUI Resource Picker (Priority: P1) 🎯 MVP

**Goal**: Build a fully reusable, column-aligned, scrollable, and fuzzy-filterable TUI picker menu.

**Independent Test**: Run `go run specs/006-tui-picker-ec2-ssh/scratch/verify_picker.go`, verify keyboard arrows navigate correctly, `/` enters search mode, typing filters the list, `Esc` cancels, and `Enter` returns the selected ID.

### Implementation for User Story 1

- [x] T002 [P] [US1] Define Picker Row model and public API functions in `internal/ui/picker/picker.go`
- [x] T003 [US1] Implement Bubble Tea model, custom column-alignment view formatting, and Lip Gloss styling in `internal/ui/picker/picker.go`
- [x] T004 [US1] Implement case-insensitive fuzzy filtering logic across all fields in `internal/ui/picker/picker.go`

**Checkpoint**: Reusable TUI picker can be verified in isolation using the scratch script.

---

## Phase 4: User Story 2 - EC2 Instance Discovery & SSM Session (Priority: P1)

**Goal**: Connect `cx ec2` to AWS instance listings and launch an interactive SSM terminal shell.

**Independent Test**: Select a workspace, run `./cx ec2`, pick a running instance, and verify it successfully transitions into the SSM shell session.

### Implementation for User Story 2

- [x] T005 [US2] Implement EC2 instance listing fetcher and AWS CLI JSON output parser in `internal/provider/aws/ec2.go`
- [x] T006 [P] [US2] Implement ConnectSSM command execution handoff in `internal/provider/aws/ssm.go`
- [x] T007 [US2] Create Cobra command ec2 in `cmd/ec2.go` that queries instances, invokes the picker, and connects
- [x] T008 [US2] Register ec2 subcommand in `cmd/root.go`

**Checkpoint**: Running `./cx ec2` completes the end-to-end cloud session connection.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation, linting, and formatting.

- [x] T009 Run manual validation scenarios in `specs/006-tui-picker-ec2-ssh/quickstart.md`
- [x] T010 Run gofmt and golangci-lint run on the entire workspace

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **User Story 1 (Phase 3)**: Depends on Setup.
- **User Story 2 (Phase 4)**: Depends on User Story 1.
- **Polish (Phase 5)**: Depends on User Stories 1 and 2.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (install packages).
2. Complete Phase 3: User Story 1 (build picker UI).
3. **VALIDATE**: Run `verify_picker.go` to test keyboard inputs.
4. Proceed to Phase 4 (AWS integration).
