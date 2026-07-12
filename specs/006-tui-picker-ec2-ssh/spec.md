# Feature Specification: TUI Resource Picker & EC2 SSH

**Feature Branch**: `006-tui-picker-ec2-ssh`

**Created**: 2026-07-12

**Status**: Draft

**Input**: User description: "Design and implement basic EC2 SSM/SSH session selection using the TUI resource picker"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reusable TUI Resource Picker (Priority: P1)

As an engineer using cx-cli, I want a consistent, interactive terminal menu to search, filter, and select a resource from a list without memorizing names or copying IDs.

**Why this priority**: It is the foundation for resource discovery across all future commands (`cx db`, `cx redis`, `cx ecs`, etc.).

**Independent Test**: Verify that initializing the picker with a list of mock objects displays a structured, column-aligned list in the terminal that filters instantly as keys are typed, and returns the selected item on `Enter`.

**Acceptance Scenarios**:
1. **Given** a list of item rows (e.g. Name, ID, IP, State), **When** I run the picker, **Then** it renders a clean TUI list with the first item highlighted.
2. **Given** the picker is active, **When** I type text (e.g. `prod`), **Then** the list is filtered instantly showing only items matching `prod` in any of their column fields.
3. **Given** the list is filtered, **When** I press `Enter`, **Then** the picker returns the selected item details and closes the TUI.
4. **Given** the picker is active, **When** I press `Esc`, **Then** it exits cleanly without returning any selection.

---

### User Story 2 - EC2 Instance Discovery & SSM Session (Priority: P1)

As an engineer, I want to run `cx ec2` to see all available virtual machines in my current workspace, choose one interactively, and open an interactive secure terminal shell session.

**Why this priority**: Delivers the primary command-line operational access to compute instances using AWS SSM Session Manager.

**Independent Test**: Verify that running `cx ec2` retrieves EC2 instances from the active AWS workspace, presents them in the TUI picker showing Name tags, Instance IDs, States, and IPs, and launches an interactive shell session using `aws ssm start-session` upon selection.

**Acceptance Scenarios**:
1. **Given** no active workspace is selected, **When** I run `cx ec2`, **Then** the CLI fails with a message instructing the user to run `cx use`.
2. **Given** the current workspace has no EC2 instances, **When** I run `cx ec2`, **Then** the CLI prints a message `"No EC2 instances found in workspace"` and exits cleanly.
3. **Given** multiple EC2 instances are available, **When** I run `cx ec2` and select a running instance (e.g., `bastion-prod`), **Then** the CLI launches `aws ssm start-session --target <instance-id>` to drop me into the shell.

---

## Edge Cases

- **Target Instance is Stopped**: If a user selects a stopped instance (state `stopped`), the CLI should output a warning `"Instance <id> is stopped; start the instance before attempting connection"` and exit gracefully.
- **Fuzzy Search Terminology**: If the search query has no matches, the picker must display a centered `"No matches found"` message and allow the user to backspace to reset the filter.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a reusable picker package `internal/ui/picker`.
- **FR-002**: The picker MUST support displaying rows containing multiple fields (Name, ID, State, IP) aligned in vertical columns.
- **FR-003**: The picker MUST filter entries case-insensitively across all fields in real-time as the user types.
- **FR-004**: System MUST support the command `cx ec2` under Cobra.
- **FR-005**: `cx ec2` MUST fetch the list of EC2 instances from the current workspace using the AWS SDK (fetching the `Name` tag, `InstanceId`, `State.Name`, and `PrivateIpAddress` or `PublicIpAddress`).
- **FR-006**: On selection, `cx ec2` MUST execute the SSM session connection by running `aws ssm start-session --target <instance-id>` interactively, binding `os.Stdin`, `os.Stdout`, and `os.Stderr` to the current terminal.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `cx ec2` loads the instances list and displays the picker UI in under 1 second (excluding AWS network API response time).
- **SC-002**: Fuzzy search UI updates in under 15ms.
- **SC-003**: Selecting an instance transitions the terminal into the interactive SSM shell within 500ms.

---

## Assumptions

- The `aws` CLI and `session-manager-plugin` are installed on the user's PATH (already verified by our provider checks).
- The active AWS profile has permissions to execute `ec2:DescribeInstances` and `ssm:StartSession`.
