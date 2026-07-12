# Implementation Plan: TUI Resource Picker & EC2 SSH

**Branch**: `006-tui-picker-ec2-ssh` | **Date**: 2026-07-12 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/006-tui-picker-ec2-ssh/spec.md`

## Summary

This feature implements the reusable Terminal User Interface (TUI) resource picker package (`internal/ui/picker`) and wires it into a new CLI command `cx ec2`. 
* We will use **Charm Bubble Tea** (`github.com/charmbracelet/bubbletea` and `github.com/charmbracelet/bubbles`) to build a scrollable, fuzzy-filterable menu.
* To discover EC2 instances, we will invoke the `aws ec2 describe-instances` CLI command directly, parsing its JSON output in Go to extract properties (Name tag, Instance ID, State, Private IP).
* On selection, the command will launch `aws ssm start-session --target <instance-id>` interactively.

## Technical Context

**Language/Version**: Go 1.22 / 1.26

**Primary Dependencies**:
* `github.com/charmbracelet/bubbletea` (TUI loop)
* `github.com/charmbracelet/bubbles` (textinput, list)
* `github.com/charmbracelet/lipgloss` (column layout styling and colors)

**Storage**: None (in-memory TUI state, stateless execution).

**Testing**: unit testing of the list filtering logic and AWS CLI JSON parser mapping.

**Target Platform**: macOS, Linux, Windows.

**Project Type**: CLI tool.

**Performance Goals**:
* JSON parsing and UI render in under 100ms.
* Instantaneous search filtering (under 10ms).

**Constraints**:
* Standard standard input/output streams (`os.Stdin`, `os.Stdout`, `os.Stderr`) must be correctly handed off to the interactive SSM shell session upon instance selection.
* Must handle terminals that do not support TUI rendering gracefully.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Specification Before Implementation**: Fully satisfied. Spec exists in `spec.md`.
- **Simplicity Over Cleverness**: Fully satisfied. Re-using the `aws` CLI via `exec.CommandContext` and parsing JSON is far simpler than pulling in the entire AWS SDK v2 Go dependency tree.
- **Standard Library First**: Balanced. We use Bubble Tea for TUI elements since building a raw ANSI escape-code scrollable table with keypress tracking from scratch is excessively complex.
- **Go Idioms First**: Fully satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/006-tui-picker-ec2-ssh/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code (repository root)

```text
cmd/
├── ec2.go               # New cmd: cx ec2
└── root.go              # CLI command registration
internal/ui/
└── picker/
    ├── picker.go        # Public API & Bubble Tea model definition
    └── view.go          # Custom Bubble Tea view formatting
internal/provider/aws/
    ├── ec2.go           # AWS EC2 describe instances fetcher
    └── ssm.go           # SSM session launcher helper
```

**Structure Decision**: Add `cmd/ec2.go`, `internal/ui/picker/`, and AWS helper functions to `internal/provider/aws/`.

## Complexity Tracking

*No violations identified.*
