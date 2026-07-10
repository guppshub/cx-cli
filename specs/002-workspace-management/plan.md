# Implementation Plan: Workspace Management

**Branch**: `002-workspace-management` | **Date**: 2026-07-11 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/002-workspace-management/spec.md`

## Summary

Implement a workspace management package that integrates with the configuration package to provide creation, selection, rename, deletion, and alphabetical listing of environments. The technical approach uses a generic mapping format to keep workspace settings provider-agnostic, and enforces validation rules including active workspace protection.

## Technical Context

**Language/Version**: Go 1.22.5

**Primary Dependencies**: `gopkg.in/yaml.v3` (via `internal/config`)

**Storage**: files (`config.yaml`)

**Testing**: `go test`

**Target Platform**: Linux, macOS, Windows

**Project Type**: library

**Performance Goals**: workspace creation/renaming/deletion in under 100ms

**Constraints**: active workspace protection must reject deletion of the currently active workspace; renaming must preserve the active pointer; provider configurations must remain opaque

**Scale/Scope**: v0.1 workspace management subsystem

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Provider Independence**: Workspace configurations are stored as generic maps (`map[string]any`) without importing or depending on provider-specific packages. (Pass)
- **Security by default**: Prevents deleting the active workspace to avoid breaking runtime states. (Pass)
- **Testability**: Tests are fully deterministic and run offline using mock configurations in temporary files. (Pass)

## Project Structure

### Documentation (this feature)

```text
specs/002-workspace-management/
├── plan.md              # This file
├── research.md          # Persistent storage map and sorting research
├── data-model.md        # Go workspace entities
├── quickstart.md        # Validation scenarios & tests guide
├── scratch/
│   └── verify.go        # Verification tool
└── contracts/
    └── workspace-api.md # Public API signatures and behaviors
```

### Source Code (repository root)

```text
internal/
└── workspace/
    ├── workspace.go      # Workspace manager implementation
    └── workspace_test.go # Unit tests covering workspace management lifecycle
```

**Structure Decision**: Selected Single Project structure. All source files live under `internal/workspace/` at the repository root.

## Complexity Tracking

*No Constitution Check violations detected.*
