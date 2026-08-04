# Feature Specification: ECS Task Monitor

**Feature Branch**: `009-ecs-task-monitor`

**Created**: 2026-07-29

**Status**: Draft

**Input**: User description: "build ecs functionality first: support for monitoring state of ecs service tasks"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cluster & Service Discovery (Priority: P1)

As an engineer using `cx`, I want to select an ECS cluster and service interactively using the TUI resource picker so that I do not have to memorize long cluster or service names.

**Why this priority**: Core discovery mechanism. Users must easily select the target service before they can monitor its tasks.

**Independent Test**: Running `cx ecs` queries the AWS workspace for ECS clusters, displays them in the TUI picker, and then displays the services of the selected cluster in a second TUI picker.

**Acceptance Scenarios**:
1. **Given** no active workspace is selected, **When** I run `cx ecs`, **Then** the CLI fails with a message instructing the user to run `cx use`.
2. **Given** the active AWS workspace has no ECS clusters, **When** I run `cx ecs`, **Then** the CLI prints `"No ECS clusters found in workspace"` and exits cleanly.
3. **Given** multiple clusters exist, **When** I run `cx ecs`, **Then** the TUI picker displays the list of clusters. Selecting a cluster automatically queries its services and displays them in a second picker.

---

### User Story 2 - ECS Task State Visualization (Priority: P1)

As an engineer, once I select an ECS service, I want to see a formatted list of all active and recently stopped tasks under that service, including their IDs, current status, health check status, uptime, and container exit codes.

**Why this priority**: Deliver the primary debug visibility. Stopped reasons and exit codes are critical for diagnosing why task definitions fail to start or crash.

**Independent Test**: Selecting an ECS service retrieves the list of tasks (both running and recently stopped), fetches details for each task, and prints a structured console table showing task states.

**Acceptance Scenarios**:
1. **Given** a service has no active or stopped tasks, **When** I select it, **Then** the CLI prints `"No tasks found for service <name>"` and exits cleanly.
2. **Given** a service has running tasks, **When** I select it, **Then** the CLI prints a table showing Task ID, Last Status (`RUNNING`/`PENDING`), Health Status (`HEALTHY`/`UNHEALTHY`/`UNKNOWN`), Started Time, and Host/IP.
3. **Given** a service has tasks that recently crashed/stopped, **When** I select it, **Then** the table displays `STOPPED` tasks, showing their `StoppedReason` (e.g., `Essential container in task exited`) and the container exit codes.

---

### User Story 3 - Continuous Status Monitoring (Priority: P2)

As an engineer deploying changes, I want to watch the service task states refresh automatically in real-time so that I can monitor the progress of rolling deployments or task restarts.

**Why this priority**: Speeds up verification during active infrastructure deployments or debugging loops without manual re-running of commands.

**Independent Test**: Running `cx ecs --watch` or pressing a refresh key in the task view updates the task table periodically with live status changes.

**Acceptance Scenarios**:
1. **Given** the watch mode is active, **When** task states change on ECS (e.g. from `PENDING` to `RUNNING`), **Then** the CLI updates the terminal view within the poll interval to show the new state.
2. **Given** the watch mode is active, **When** I press `Ctrl+C` or a termination key, **Then** the CLI terminates the watch loop and exits cleanly.

---

### Edge Cases

- **Task Definition Crash Loops**: If tasks are rapidly starting and stopping, the CLI should show both the current active tasks and the most recent stopped tasks (last 10) to help diagnose the crash loop.
- **Throttling/Rate Limiting**: ECS `DescribeTasks` API requests can be subject to AWS rate limits. The CLI must handle rate limiting errors gracefully by retrying with exponential backoff or warning the user.
- **Stopped Tasks Cleanup**: AWS ECS drops stopped task history after a period of time. If no tasks are found, the CLI should warn that only active and recently stopped tasks are visible.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide the command `cx ecs` under Cobra.
- **FR-002**: `cx ecs` MUST check for an active workspace and valid AWS credentials before invoking AWS ECS APIs.
- **FR-003**: System MUST fetch clusters using ECS `ListClusters` and services using `ListServices` / `DescribeServices` APIs.
- **FR-004**: System MUST reuse the existing TUI picker (`internal/ui/picker`) for cluster and service selection.
- **FR-005**: System MUST query task details using `ListTasks` and `DescribeTasks` APIs (fetching Task Arn, LastStatus, HealthStatus, CreatedAt, StartedAt, StoppedAt, StoppedReason, and container ExitCode).
- **FR-006**: System MUST support a `--watch` / `-w` flag to enable live polling of task states (refreshing at a configurable interval, defaulting to 5 seconds).
- **FR-007**: System MUST parse task UUIDs from Task ARNs to present a clean, short Task ID (e.g., `3df67a90...`) instead of the full ARN.

### Key Entities

- **ECSCluster**: Represents an AWS ECS Cluster (Name, ARN).
- **ECSService**: Represents an ECS Service (Name, Cluster Name, Desired Count, Running Count).
- **ECSTask**: Represents a single task instance (Task ID, ARN, Last Status, Health Status, Started At, Stopped Reason, Exit Code).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Transitioning from TUI pickers to the task listing takes less than 1 second (excluding AWS network latency).
- **SC-002**: Task list watch loop uses less than 1% CPU when idle between poll intervals.
- **SC-003**: Clean, formatted Task IDs are displayed (maximum 12 characters of the UUID) to prevent terminal wrapping.

## Assumptions

- **A-001**: The user's active AWS profile has IAM permissions for `ecs:ListClusters`, `ecs:ListServices`, `ecs:DescribeServices`, `ecs:ListTasks`, and `ecs:DescribeTasks`.
- **A-002**: The `aws` CLI/SDK session credentials resolved by the active workspace are used for authorization.
