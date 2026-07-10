# Implementation Plan: Configuration Foundation

**Branch**: `001-config-foundation` | **Date**: 2026-07-11 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-config-foundation/spec.md`

## Summary

Implement a configuration subsystem that resolves `config.yaml` and `state.json` on Linux, macOS, and Windows. The technical approach leverages Go's standard library (`os`, `filepath`, `runtime`) for path resolution and atomic file writes, alongside `gopkg.in/yaml.v3` for parsing the YAML configuration format. The subsystem maintains strict separation between configuration and runtime state directories.

## Technical Context

**Language/Version**: Go 1.22.5

**Primary Dependencies**: `gopkg.in/yaml.v3`

**Storage**: files (`config.yaml` for configuration, `state.json` for state)

**Testing**: `go test`

**Target Platform**: Linux, macOS, Windows

**Project Type**: library/cli subsystem

**Performance Goals**: resolving and loading configuration in under 50ms

**Constraints**: configuration and state directories must remain completely separated; configuration is read-only during execution

**Scale/Scope**: v0.1 core configuration foundation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Go standard library first**: Resolves directory structures using built-in Go OS commands without third-party path dependencies. (Pass)
- **Simplicity over cleverness**: Keeps the architecture simple; no complex database files or abstractions, only simple files mapped to structures. (Pass)
- **Explicit error handling**: Uses custom error wrapping with contextual details. (Pass)
- **Testability**: Unit tests will mock temp directories using Go's `t.TempDir()`. (Pass)

## Project Structure

### Documentation (this feature)

```text
specs/001-config-foundation/
├── plan.md              # This file
├── research.md          # Multi-platform path and parser research
├── data-model.md        # Go config & state structures
├── quickstart.md        # Validation scenarios & tests guide
├── scratch/
│   └── verify.go        # Verification tool
└── contracts/
    └── config-api.md    # Public API signatures and behaviors
```

### Source Code (repository root)

```text
internal/
└── config/
    ├── config.go        # Path resolution, Load, Save, Default, and Validate APIs
    └── config_test.go   # Unit tests covering path resolution, loader, saver, and validators
```

**Structure Decision**: Selected Single Project structure. All source files live under `internal/config/` at the repository root.

## Complexity Tracking

*No Constitution Check violations detected.*
