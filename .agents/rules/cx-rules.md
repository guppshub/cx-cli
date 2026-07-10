# Workspace Rules: cx

These rules and principles guide the development of **cx**, a workflow-oriented cloud operations CLI. They are automatically applied to all code edits and planning tasks within this workspace.

---

## 1. Design Principles

- **Workflow over APIs**: Commands represent operational workflows (e.g., `cx db`, `cx logs`, `cx compute`) rather than low-level cloud provider APIs (e.g., `cx rds`, `cx ssm`).
- **Discover over Configure**: Autodetect resources and present interactive prompts/pickers instead of requiring users to remember names, ports, or IDs.
- **Convention over Configuration**: Use sensible defaults to minimize user inputs.
- **Provider Independence**: Core business logic must remain provider-agnostic.
- **Long-running Operations are First-class**: Manage persistent tunnels and sessions as core features.
- **Human-first UX**: Interactive prompts (e.g., using Bubble Tea) are preferred over complex flags.
- **AI-friendly Architecture**: Maintain explicit boundaries so human and AI contributors can understand code easily.

---

## 2. Architectural Rules

1. **No Business Logic in CLI**: CLI command packages (`cmd/`) are responsible only for flag parsing, validation, and invoking application services.
2. **Application Layer Isolation**: All workflow business logic belongs in the Application Layer (`internal/app/` or `pkg/app/`).
3. **Immutable Configuration**: Configuration is read-only during command execution.
4. **Mutable State**: The Runtime State (`~/.local/state/cx/state.json`) is the only mutable storage.
5. **No CLI Access in Providers**: Provider implementations must not import or access CLI packages.
6. **No Configuration Access in Providers**: Providers must not read the configuration file directly.
7. **No Business Logic in UI**: Presentation components (e.g., prompts, tables, loaders) must not execute workflow logic.
8. **Abstract Connection Supervision**: Connection supervisor interfaces must remain agnostic of the underlying implementation (e.g., `tmux` vs background daemons).
9. **Testability**: Workflows must be testable locally without active cloud access.
10. **Provider Isolation**: Cloud-specific implementations (e.g. AWS SDK, CLI wrappers) must be completely isolated behind the Provider interface.
