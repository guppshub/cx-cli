# Feature Specification: Version & Update Commands

**Feature Branch**: `008-version-and-update-commands`

**Created**: 2026-07-14

**Status**: Draft

**Input**: User description: "1) version command 2) update command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Check Installed Version (Priority: P1)

As a developer, I want to quickly check the exact version of the `cx` CLI currently installed, along with build metadata, so that I can report bugs accurately and know if I am on the correct release.

**Why this priority**: P1 because knowing the compiled version is a fundamental debugging requirement for all developers and support flows.

**Independent Test**: Running `cx version` or `cx -v` / `cx --version` prints the compiled tag version (e.g. `v0.1.16`) and exits with code 0.

**Acceptance Scenarios**:

1. **Given** a compiled CLI binary, **When** I run `cx version`, **Then** the output prints `cx version <tag-version> (commit: <hash>, built: <timestamp>)` or a local development default (e.g., `v0.0.0-dev`) if build metadata was not injected.
2. **Given** any directory context, **When** I run `cx -v` or `cx --version`, **Then** it produces the identical output.

---

### User Story 2 - Check for Updates (Priority: P2)

As a developer, I want to check if a newer version of the `cx` CLI is available on GitHub without necessarily performing the update immediately, so that I can decide when to upgrade.

**Why this priority**: P2 because it allows developers to safely check for upgrades without modifying their environment in restricted settings.

**Independent Test**: Running `cx update --check` contacts the GitHub API and reports whether a newer release exists.

**Acceptance Scenarios**:

1. **Given** I am running the latest version, **When** I run `cx update --check`, **Then** the CLI prints "You are already running the latest version of cx (<version>)."
2. **Given** a newer version is available, **When** I run `cx update --check`, **Then** the CLI prints "A newer version of cx is available: <new-version> (current: <current-version>)."

---

### User Story 3 - Perform Auto-Upgrade (Priority: P2)

As a developer, I want the CLI to automatically download and install the latest release matching my current operating system and architecture, so that I don't have to manually download binaries or re-run installation scripts.

**Why this priority**: P2 because it provides a frictionless developer experience and keeps the team aligned on the same versions.

**Independent Test**: Running `cx update` detects, downloads, and replaces the current binary with the latest release from GitHub.

**Acceptance Scenarios**:

1. **Given** a newer version is available, **When** I run `cx update`, **Then** the CLI prompts me to confirm, downloads the latest binary, replaces the active executable, and prints a success message.
2. **Given** a newer version is available and I pass `cx update --yes` (or `-y`), **Then** it performs the upgrade automatically without prompting.
3. **Given** I am running the latest version, **When** I run `cx update`, **Then** it prints that I am already up-to-date and exits without downloading anything.

---

### Edge Cases

- **Windows Executable Lock**: On Windows, the OS locks the active `cx.exe` file while it is running. The system must perform a rename-dance (rename `cx.exe` -> `cx_old.exe`, write new file `cx.exe`, and clean up) to update itself without crashes or access violations.
- **GitHub API Rate Limits**: If the GitHub API is unreachable or rate-limited, the update command must print a clear, user-friendly warning message and exit gracefully without breaking the current installation.
- **Unverified Releases / Downgrades**: If a release on GitHub is a pre-release or draft, the updater should ignore it by default, prioritizing official releases.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST support a `version` command and `-v` / `--version` flags on the root command.
- **FR-002**: The version number MUST be injectability-compatible during build-time using `go build -ldflags "-X ..."` to bind the release tag name, commit hash, and timestamp.
- **FR-003**: The `update` command MUST fetch the latest release tag from the public repository `guppshub/cx-cli` on GitHub.
- **FR-004**: The update command MUST detect the active OS (darwin, linux, windows) and architecture (amd64, arm64) to select the correct release asset.
- **FR-005**: The self-update mechanism MUST replace the currently executing active binary file and preserve execute permissions (`0755`) on Unix-like systems.
- **FR-006**: The CLI MUST support a `--check` flag to verify updates without modifying the binary.
- **FR-007**: The CLI MUST support a `--yes` (or `-y`) flag to bypass interactive confirmation prompts during the update.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `cx version` returns the correct tag version and build metadata in under 50ms.
- **SC-002**: A developer can complete an update using `cx update -y` in under 10 seconds (on average network speeds).
- **SC-003**: The update command fails gracefully (exiting with code 0 or 1 with clear user-facing error messages) when offline or rate-limited, without corrupting the existing binary.
- **SC-004**: Self-update successfully works on Windows PowerShell/CMD, macOS, and Linux.

## Assumptions

- **A-001**: Users have internet connectivity to access the `api.github.com` endpoints during updates.
- **A-002**: The compiled version tags on GitHub follow semantic versioning (e.g. `vX.Y.Z`).
- **A-003**: The compiled binaries released on GitHub follow a consistent asset naming convention (e.g. `cx-<os>-<arch>.tar.gz` or `cx-<os>-<arch>.exe`).
- **A-004**: The directory containing `cx` is writable by the user running the `cx update` command.
