# Feature Specification: Configuration Foundation

**Feature Branch**: `001-config-foundation`

**Created**: 2026-07-11

**Status**: Draft

**Input**: User description: "Implement the configuration subsystem that serves as the foundation for all future cx-cli functionality..."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Multi-Platform Configuration Resolution & Loading (Priority: P1)

As a developer or user running `cx-cli`, the application automatically resolves and loads my configuration from the correct system-specific directories (XDG on Linux, standard library paths on macOS and Windows) or loads a valid default configuration if none exists.

**Why this priority**: This is the core foundational task; all subsequent configurations and subsystems depend on knowing where and how to locate the configuration files.

**Independent Test**: Can be verified by executing a test program that attempts to resolve config directory paths on different operating systems and loads the default configuration when files do not exist.

**Acceptance Scenarios**:

1. **Given** the configuration file `config.yaml` does not exist on disk, **When** the application attempts to load it, **Then** it returns the default configuration struct without automatically creating a file on disk.
2. **Given** the configuration file `config.yaml` exists in the system-specific configuration directory and contains a valid structure, **When** the application loads it, **Then** it successfully parses the configuration and returns it.

---

### User Story 2 - Configuration Persistence & Directory Provisioning (Priority: P1)

As a developer or user running `cx-cli`, I can save my updated configurations back to disk, and the system automatically handles creating any missing parent directories and handles atomic file operations safely.

**Why this priority**: Crucial for configuration updates (e.g. adding a context or preference).

**Independent Test**: Can be verified by calling the save API on a default config and checking that directories are created and `config.yaml` is written with correct permissions.

**Acceptance Scenarios**:

1. **Given** a new configuration file needs to be written and parent directories do not exist, **When** the application saves the configuration, **Then** it creates the parent directories and writes the YAML file successfully.

---

### User Story 3 - Configuration Schema & Version Validation (Priority: P2)

As a user, I want the CLI to validate my configuration version and structure, and fail with descriptive errors if it encounters unsupported versions or malformed files.

**Why this priority**: Prevent silent errors or unexpected behavior from corrupt configs or outdated formats.

**Independent Test**: Can be verified by attempting to load configs with unsupported versions or invalid formats and checking for expected error messages.

**Acceptance Scenarios**:

1. **Given** the configuration file has a version format of `"2"` but only `"1"` is supported, **When** the application loads the configuration, **Then** it returns a validation error prefixing `"loading configuration: unsupported configuration version \"2\""`.
2. **Given** the configuration file contains duplicate context names, **When** the application validates the configuration, **Then** it returns a validation error detailing the duplicate names.

---

### Edge Cases

- **Malformed YAML Syntax**: If `config.yaml` contains invalid YAML structures, the loader must return a clear, structured parsing error instead of panicking.
- **Simultaneous Access**: If multiple CLI processes access the configuration/state files simultaneously, the system must use atomic writes (e.g., writing to a temp file and renaming) to prevent corruption.
- **Write-Protected Directories**: If parent directories are write-protected, saving must fail cleanly with a permission error without leaving corrupt temp files.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST resolve the configuration directory path based on the host OS:
  - Linux: `$XDG_CONFIG_HOME/cx` (falls back to `~/.config/cx`).
  - macOS: Standard user configuration directory.
  - Windows: Standard AppData User configuration directory.
- **FR-002**: System MUST resolve the runtime state directory path based on the host OS:
  - Linux: `$XDG_STATE_HOME/cx` (falls back to `~/.local/state/cx`).
  - macOS: Standard user local state/data directory.
  - Windows: Standard AppData Local state directory.
- **FR-003**: System MUST store configuration in `config.yaml` and runtime state in `state.json`.
- **FR-004**: System MUST keep configuration and runtime state directories separated.
- **FR-005**: System MUST enforce version validation on loaded configuration files. If the version field is not `"1"`, it MUST reject the configuration with a validation error.
- **FR-006**: System MUST return a valid default configuration (with empty contexts and preferences maps and version `"1"`) if the file does not exist.
- **FR-007**: System MUST NOT write or create files on disk during the loading process.
- **FR-008**: System MUST perform atomic writes when saving the configuration to avoid partial file corruption.
- **FR-009**: System MUST create parent directories automatically if they do not exist when saving.
- **FR-010**: System MUST validate for duplicate context names or malformed structure and return structured, wrapped error messages.

### Key Entities

- **Config**: Root object containing `version`, `contexts`, and `preferences`.
- **Preferences**: A map of key-value pairs representing user settings.
- **Contexts**: A map representing operational environments (e.g. staging, production).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Configuration is resolved and loaded in under 50ms for valid files.
- **SC-002**: Error messages for validation or structural errors consistently include contextual prefixes (e.g., `loading configuration: ...`) and wrap the underlying root cause.
- **SC-003**: 100% of public API functions (`Path`, `StatePath`, `Load`, `Save`, `Default`, `Validate`) behave deterministically and are fully covered by unit tests without requiring external network or cloud access.

## Assumptions

- The standard configuration and state locations conform to the platform specifications (XDG Base Directory Specification on Linux/macOS, AppData on Windows).
- The user has appropriate read and write access to their home configuration and state directories.
- The `config.yaml` file is relatively small (<100KB), making atomic memory-to-disk writes fast and non-blocking.
