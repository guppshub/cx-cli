# Feature Specification: cx init

**Feature Branch**: `005-cx-init`

**Created**: 2026-07-12

**Status**: Draft

**Input**: User description: "Implement cx init command to initialize configuration file"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Initialize configuration for the first time (Priority: P1)

As an engineer using cx-cli for the first time, I want to run `cx init` so that the default configuration folder and a starter `config.yaml` are generated automatically for me.

**Why this priority**: Required for new user onboarding and getting started quickly.

**Independent Test**: Verify that in a clean environment (no `~/.config/cx/` folder), running `cx init` creates `~/.config/cx/config.yaml` with a valid default configuration template and prints a success message.

**Acceptance Scenarios**:

1. **Given** the configuration file `config.yaml` does not exist, **When** I run `cx init`, **Then** the CLI creates the directory, writes a starter `config.yaml`, and prints a success message containing the generated file path.
2. **Given** `cx init` is run successfully, **When** I inspect the created file, **Then** it contains starter workspace configurations and comments explaining the layout.

---

### User Story 2 - Prevent accidental overwrite (Priority: P1)

As an engineer with an existing configuration, I want `cx init` to protect my existing settings from being silently overwritten.

**Why this priority**: Prevents critical user configuration data loss.

**Independent Test**: Verify that running `cx init` when `~/.config/cx/config.yaml` already exists fails with an error and does not modify the file.

**Acceptance Scenarios**:

1. **Given** the configuration file `config.yaml` already exists, **When** I run `cx init`, **Then** the CLI prints a warning message `"configuration already exists at <path>"` and exits with code 1 without modifying the file.
2. **Given** the configuration file `config.yaml` already exists, **When** I run `cx init --force` (or `cx init -f`), **Then** the CLI overwrites the configuration with the default template and prints a confirmation message.

---

### Edge Cases

- **Missing Parent Directories**: If the user's base config folder (like `~/.config`) does not exist, `cx init` should create it and all intermediate directories cleanly.
- **Write Permission Denied**: If the CLI lacks permissions to write to the config path, it must return a user-friendly error explaining the write failure, rather than crashing with a raw Go stack trace.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide the CLI command `cx init`.
- **FR-002**: System MUST resolve the target config path by invoking `config.Path()`.
- **FR-003**: System MUST create the parent directories of the config path if they do not exist.
- **FR-004**: System MUST check if the configuration file already exists before writing to it.
- **FR-005**: If the file exists, the system MUST exit with code 1 and print an error unless the `--force` (or `-f`) flag is provided.
- **FR-006**: System MUST support the `--force` (or `-f`) flag to overwrite an existing configuration.
- **FR-007**: System MUST write a starter configuration template (valid YAML conforming to the `config.Config` structure) containing informative comments.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `cx init` completes file creation and validation in under 50ms.
- **SC-002**: The generated configuration file is syntactically valid YAML and can be loaded successfully by the `config.Load()` utility.

---

## Assumptions

- Cobra CLI is the standard command routing framework.
- The standard user configuration location is platform-dependent and resolved dynamically via `config.Path()`.
