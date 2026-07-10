# Feature Specification: Workspace Management

**Feature Branch**: `002-workspace-management`

**Created**: 2026-07-11

**Status**: Draft

**Input**: User description: "Implement workspace management for cx-cli..."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add and List Workspaces (Priority: P1)

As an engineer, I can define a new workspace (e.g. `staging`, `production`) containing the cloud provider name and config without validating the provider config details, and I can list all configured workspaces deterministically.

**Why this priority**: Foundational for saving workspaces to configuration.

**Independent Test**: Can be verified by executing a test program that adds two workspaces and lists them, verifying they appear in alphabetical order.

**Acceptance Scenarios**:

1. **Given** the workspace name is `staging` and the provider is `aws`, **When** the application adds the workspace, **Then** it is saved to the configuration.
2. **Given** multiple workspaces exist in configuration, **When** the application lists the workspaces, **Then** they are returned in alphabetical order by name.

---

### User Story 2 - Select and Retrieve Active Workspace (Priority: P1)

As an engineer, I can switch the active workspace and retrieve its provider config, so subsequent workflows execute under this target context.

**Why this priority**: Allows context-awareness (`cx use`).

**Independent Test**: Can be verified by selecting a workspace and retrieving it, ensuring the returned workspace matches.

**Acceptance Scenarios**:

1. **Given** the workspace `staging` exists in the configuration, **When** the application selects the workspace, **Then** the configuration's current active workspace pointer is updated to `staging`.
2. **Given** no active workspace is selected or a non-existent workspace is requested, **When** the application attempts to select it, **Then** it returns a descriptive error.

---

### User Story 3 - Rename and Delete Workspaces with Protection (Priority: P2)

As an engineer, I can rename a workspace (updating references if active) or delete it (rejected if active), preventing configuration errors.

**Why this priority**: Important for workspace life cycle and validation boundaries.

**Independent Test**: Can be verified by attempting to delete the active workspace (returns error) and renaming the active workspace (active pointer moves to new name).

**Acceptance Scenarios**:

1. **Given** the workspace `staging` is the currently active workspace, **When** the application attempts to delete it, **Then** the deletion is rejected with a descriptive error.
2. **Given** the workspace `staging` is the active workspace, **When** the application renames it to `staging-new`, **Then** the active workspace pointer is updated to `staging-new` automatically.

---

### Edge Cases

- **Deleting Last Workspace**: If only one workspace exists and it is the active one, deletion must be rejected.
- **Renaming to Existing Name**: Renaming a workspace to a name that is already occupied by another workspace must be rejected with a duplicate error.
- **Invalid Name Characters**: Workspace names containing spaces, slash characters, or special punctuation must be rejected during validation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support creating workspaces with unique, non-empty names.
- **FR-002**: System MUST reject duplicate workspace names with a descriptive error.
- **FR-003**: System MUST store the workspace provider name (MUST be specified) and config (treated as opaque map[string]any).
- **FR-004**: System MUST allow selecting an existing workspace, updating `current` (or active pointer) in the configuration.
- **FR-005**: System MUST reject selecting a non-existent workspace with an error.
- **FR-006**: System MUST return all configured workspaces ordered deterministically (alphabetically by name).
- **FR-007**: System MUST allow deleting a workspace, provided it is NOT the currently active workspace.
- **FR-008**: System MUST reject deleting the active workspace with a descriptive error.
- **FR-009**: System MUST allow renaming a workspace.
- **FR-010**: System MUST automatically update the active workspace pointer if the active workspace is renamed.

### Key Entities

- **Workspace**: Struct containing `Name` (string), `Provider` (string), and `Config` (map[string]any).
- **WorkspaceSummary**: Struct containing `Name` (string), `Provider` (string), and `IsActive` (bool).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Workspace creation, renaming, and deletion execute in under 100ms.
- **SC-002**: Active workspace protections (e.g. rejecting deletion of active context) fail safely and return human-readable messages in under 50ms.
- **SC-003**: 100% of public API functions (`Add`, `Delete`, `Rename`, `Use`, `Current`, `List`, `Get`) behave deterministically and are fully covered by unit tests without requiring external cloud access or provider modules.

## Assumptions

- Workspaces are persisted inside `config.yaml` under the `contexts` map, where each context maps directly to a workspace.
- The configuration file is accessible and writable.
- Workspace names contain only alphanumeric characters, hyphens, and underscores.
