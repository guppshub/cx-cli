# Implementation Plan: Connection Supervisor

**Branch**: `007-connection-supervisor` | **Date**: 2026-07-13 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/007-connection-supervisor/spec.md`

## Summary

This feature refactors the CLI connection management to use a structured **Connection Supervisor** architecture. 
* We will define core interfaces: `Supervisor` (manages lifecycle), `Connection` (wraps running process), and `RestartPolicy` (governs reconnect backoffs).
* The supervisor will manage state transitions (`Stopped`, `Starting`, `Healthy`, `Restarting`, `Failed`) and persist rich metadata to `state.json`.
* We will update `cmd/db.go`, `cmd/redis.go`, and `cmd/status.go` to use the supervisor, replacing the current ad-hoc polling and blocking.

## Technical Context

**Language/Version**: Go 1.22 / 1.24

**Primary Dependencies**: None (Standard Library only).

**Storage**: Updates the schema of `state.json` (managed by `internal/state/state.go`).

**Testing**: Unit tests verifying state machine transitions, restart policy backoff delays, and mock connection dial failures.

**Target Platform**: macOS, Linux (WSL), Windows.

**Performance Goals**:
* Event-based process exit detection under 20ms.
* Zero cpu overhead (no polling loops).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Specification Before Implementation**: Fully satisfied. Spec exists in `spec.md`.
- **Simplicity Over Cleverness**: Fully satisfied. Using interfaces and clean state transitions keeps the code modular, self-contained, and testable.
- **Standard Library First**: Fully satisfied. No external process supervisors or state machines.
- **Go Idioms First**: Fully satisfied. Using channels for event-driven process synchronization.
- **Long-running operations must be supervised, not merely started**: This new constitutional principle is the core design philosophy of this feature.

## Project Structure

### Documentation

```text
specs/007-connection-supervisor/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code

```text
internal/connection/
├── supervisor.go      // Main orchestration & State Machine
├── connection.go      // Connection & Dialer interfaces
├── restart.go         // RestartPolicies (Fixed, Exponential, etc.)
├── health.go          // Layered Health Checks (Process -> Port -> Protocol)
├── metadata.go        // Metadata structures persisted to state.json
├── process_unix.go    // (Existing) Platform process killing
└── process_windows.go // (Existing) Platform process killing
internal/state/
└── state.go           // Updated state.json struct mappings
cmd/
├── db.go              // Integrates Supervisor for DB tunnels
├── redis.go           // Integrates Supervisor for Redis tunnels
└── status.go          // Renders rich metadata from state.json
```
