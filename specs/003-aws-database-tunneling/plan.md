# Implementation Plan: AWS Database Tunneling

**Branch**: `003-aws-database-tunneling` | **Date**: 2026-07-11 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/003-aws-database-tunneling/spec.md`

## Summary

Implement the AWS Database Tunneling vertical slice by building a concrete `aws` provider that wraps the `session-manager-plugin` CLI inside a native Go `net.Conn` stream, and a database workflow `Controller` that coordinates the local socket listener forwarding loop and registers active connection sessions in `state.json`.

## Technical Context

**Language/Version**: Go 1.22.5

**Primary Dependencies**: Go Standard Library (`os/exec`, `net`, `io`, `context`)

**Storage**: state.json files (via `internal/state`)

**Testing**: `go test`

**Target Platform**: Linux, macOS, Windows

**Project Type**: library

**Performance Goals**: port forward listener successfully binds and forwards data to remote RDS host in under 500ms

**Constraints**: active connection tracking in `state.json`, fallback binding on port conflicts, headless authentication prompter callback

**Scale/Scope**: v0.1 concrete AWS database tunneling workflow slice

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Emerging Abstractions**: We implement the concrete AWS dialer directly under a simple local interface (`TunnelDialer`), avoiding an over-engineered abstract provider framework. (Pass)
- **Network Socket Boundary**: Spawning of the `session-manager-plugin` subprocess is hidden entirely behind a clean Go `net.Conn` stream, keeping the core database forwarder loop 100% decoupled from execution details. (Pass)
- **Headless Provider**: User-interactive authentication prompts (e.g. MFA prompts) are driven via callback functions rather than standard stdout/stdin writes. (Pass)

## Project Structure

### Documentation (this feature)

```text
specs/003-aws-database-tunneling/
├── plan.md              # This file
├── research.md          # Subprocess connection wrapper and forwarding research
├── data-model.md        # Dialer entities
├── quickstart.md        # Validation scenarios & tests guide
├── scratch/
│   └── verify.go        # Verification tool
└── contracts/
    └── tunnel-api.md    # Public API signatures and behaviors
```

### Source Code (repository root)

```text
internal/
├── provider/
│   └── aws/
│       ├── aws.go       # AWS connection dialer
│       └── aws_test.go  # AWS connection unit tests
└── workflow/
    └── db/
        ├── db.go        # Port forwarder listener loop
        └── db_test.go   # Forwarder loop unit tests
```

**Structure Decision**: Selected Single Project structure. Dialer source files live under `internal/provider/aws/` and controller files under `internal/workflow/db/`.

## Complexity Tracking

*No Constitution Check violations detected.*
