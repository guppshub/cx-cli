# Implementation Plan: cx init

**Branch**: `005-cx-init` | **Date**: 2026-07-12 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/005-cx-init/spec.md`

## Summary

The `cx init` command will initialize a fresh configuration environment for `cx-cli`. It automatically resolves the configuration path using `config.Path()`, creates the parent directory structure if missing, and writes a starter `config.yaml` with pre-defined examples and documentation. If the configuration already exists, it fails with a warning message to protect user settings, unless the `--force` (or `-f`) flag is explicitly supplied.

## Technical Context

**Language/Version**: Go 1.26

**Primary Dependencies**: Cobra CLI framework (already used in the project)

**Storage**: YAML file (`~/.config/cx/config.yaml` or Windows AppData equivalent)

**Testing**: standard Go `testing` package (`go test`)

**Target Platform**: macOS, Linux, Windows

**Project Type**: CLI tool

**Performance Goals**: File creation and validation in under 50ms

**Constraints**: Must fail and exit with code 1 if the file exists and `--force` is not specified; must handle file system write permissions gracefully.

**Scale/Scope**: Small scope (adding one command `cmd/init.go` and referencing existing `config.Path()`).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Specification Before Implementation**: Fully satisfied. Spec exists in `spec.md`.
- **Simplicity Over Cleverness**: Fully satisfied. Writing a static template string directly to a file using the Go standard library is simple and robust.
- **Standard Library First**: Fully satisfied. Directory creation, file checks, and file writes will use `os` and `path/filepath`.
- **Go Idioms First**: Fully satisfied. Standard Go error handling and Cobra command registration.
- **Security by Default**: Fully satisfied. The template contains no secrets; it only uses mock placeholders.

## Project Structure

### Documentation (this feature)

```text
specs/005-cx-init/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (created by speckit-tasks)
```

### Source Code (repository root)

```text
cmd/
├── init.go              # New command file
└── root.go              # cobra root registration
internal/config/
└── paths.go             # Existing path resolver
```

**Structure Decision**: Single project layout. We will create `cmd/init.go` and wire it up to `cmd/root.go`.

## Complexity Tracking

*No violations identified.*
