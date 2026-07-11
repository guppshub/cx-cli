# Implementation Plan: CLI Integration

**Branch**: `004-cli-integration` | **Date**: 2026-07-11 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/004-cli-integration/spec.md`

## Summary

Implement the user-facing CLI command layer for `cx-cli` using Cobra. This introduces the `use` command to select workspaces and the `db` command to resolve database targets from active workspace configuration, invoke cloud dialers, bind local ports, and handle graceful exit traps.

## Technical Context

**Language/Version**: Go 1.22.5

**Primary Dependencies**: `github.com/spf13/cobra`, `gopkg.in/yaml.v3`

**Storage**: `config.yaml`, `state.json`

**Testing**: `go test`

**Target Platform**: Linux, macOS, Windows

**Project Type**: CLI

**Performance Goals**: command bootstrap under 50ms, graceful Ctrl+C cleanup in under 100ms

**Constraints**: config serialization boundaries, signal interrupt handlers, clean output messages

**Scale/Scope**: v0.1 CLI command integration layer

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **No Dependency Injection**: All packages and managers are wired manually in `cmd/` (e.g. `loader := config.New()`, `store := workspace.New(loader)`), maintaining clean, idiomatic Go construction. (Pass)
- **Headless Provider Boundary**: CLI command layers ownBubble Tea/prompt inputs and handle OS signal traps, keeping provider packages clean. (Pass)

## Project Structure

### Documentation (this feature)

```text
specs/004-cli-integration/
├── plan.md              # This file
├── research.md          # Dynamic YAML parsing and signal trap research
├── data-model.md        # DatabaseResource entity
├── quickstart.md        # Validation scenarios & tests guide
└── contracts/
    └── cli-api.md       # Command syntax and resolver signature
```

### Source Code (repository root)

```text
cmd/
├── root.go              # Cobra CLI root
├── use.go               # Cobra workspace selection command
└── db.go                # Cobra database forwarder command
internal/
└── resource/
    ├── database.go      # Workspace database resolver helper
    └── database_test.go # Resolver unit tests
```

**Structure Decision**: Selected Single Project structure. Command files live in `cmd/` and resolver helper files in `internal/resource/`.

## Complexity Tracking

*No Constitution Check violations detected.*
