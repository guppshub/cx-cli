# Implementation Plan: Version & Update Commands

**Branch**: `008-version-and-update-commands` | **Date**: 2026-07-14 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/008-version-and-update-commands/spec.md`

## Summary

Implement `cx version` and `cx update` commands. 
* **Version**: Displays the compiled CLI version tag, git SHA, and build time using Go's build-time `-ldflags` variable injection.
* **Update**: Queries the public GitHub Releases API for `guppshub/cx-cli` using Go's standard library `net/http` to check for newer releases. If a newer release exists, it downloads the correct OS/architecture binary and updates itself using a platform-safe rename sequence (rename active -> active.old, write new -> active, delete active.old) to prevent Windows OS file lock issues.

## Technical Context

**Language/Version**: Go 1.24.2

**Primary Dependencies**: `github.com/spf13/cobra` (already in project)

**Storage**: None (in-memory parsing of GitHub JSON payload)

**Testing**: Standard `go test` with `net/http/httptest` to mock GitHub API responses

**Target Platform**: Linux, macOS (darwin), Windows

**Project Type**: CLI tool

**Performance Goals**: Update checks and version commands respond in <100ms on a warm connection

**Constraints**: Fall back gracefully (exit 0/1 with friendly errors) when offline or rate-limited

**Scale/Scope**: Single client utility command set

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Specification Before Implementation**: Pass. Feature specification created at `spec.md`.
- **Simplicity Over Cleverness**: Pass. No heavy dependencies or abstractions. Uses standard library utilities.
- **Go Idioms First**: Pass. Follows Cobra standards and standard Go package organization.
- **Standard Library First**: Pass. Only standard HTTP, JSON, and OS file manipulation packages are used.
- **Production-Quality Code**: Pass. Explicit error handling and thorough testing of all helper functions.
- **Security by Default**: Pass. Validate that downloaded release URLs only point to the official `github.com/guppshub/cx-cli` repository to prevent DNS spoofing or injection of arbitrary binaries.

## Project Structure

### Documentation (this feature)

```text
specs/008-version-and-update-commands/
├── plan.md              # This file
├── research.md          # Tech decisions and platform locks details
├── data-model.md        # API payloads and runtime mappings
└── quickstart.md        # Step-by-step test commands
```

### Source Code

```text
cmd/
├── root.go              # Root command configuration
├── version.go           # New file: version command
└── update.go            # New file: update command

internal/
└── update/
    ├── update.go        # New package: GitHub client, version comparison, and self-replacement logic
    └── update_test.go   # New file: Unit tests with mocked HTTP responses
```

**Structure Decision**: Single project layout matching the existing project organization. High-level commands live in `cmd/`, while core update engine logic is encapsulated in `internal/update/` to keep command handlers simple and testable.

## Complexity Tracking

*No violations.*
