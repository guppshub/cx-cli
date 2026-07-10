# Research: Configuration Foundation

## 1. Multi-Platform Directory Resolution

We researched standard platform conventions and Go library capabilities to resolve user configuration and local runtime state directories.

### Go Standard Library Capabilities
Go provides:
- `os.UserConfigDir()`: Resolves to standard user config paths (e.g. `~/.config` on Linux, `~/Library/Application Support` on macOS, `%AppData%` on Windows).
- `os.UserCacheDir()`: Resolves to cache/local app data paths (e.g. `~/.cache` on Linux, `~/Library/Caches` on macOS, `%LocalAppData%` on Windows).

Go does not provide a built-in function for local state (`$XDG_STATE_HOME` or `~/.local/state`).

### Selected Approach
To keep the codebase standard-library first and avoid bloated dependencies (like `github.com/adrg/xdg`), we implement custom path resolution logic:

1. **Configuration Directory**:
   - Resolve via `os.UserConfigDir()`.
   - Fall back to `$HOME/.config` if unavailable.
   - Join with `cx` subdirectory.
   - Config file is at `<ConfigDir>/config.yaml`.

2. **Runtime State Directory**:
   - Check `$XDG_STATE_HOME` environment variable first.
   - **Linux/Unix**: Fall back to `$HOME/.local/state/cx`.
   - **macOS**: Fall back to `$HOME/Library/Application Support/cx/state` (separating state from the root config folder).
   - **Windows**: Resolve via `%LOCALAPPDATA%` (using `os.UserCacheDir()`) to use the Local AppData folder, separating it from the Roaming AppData folder (`%APPDATA%`) used for configuration.
   - State file is at `<StateDir>/state.json`.

---

## 2. YAML Parser Dependency

### Options Evaluated
1. **gopkg.in/yaml.v3**: The industry standard for parsing YAML in Go. It supports comments preservation, struct mapping, and robust validation.
2. **Standard library json**: Convert configuration to JSON. Rejected because the PRD mandates configuration must be human-readable and manually editable in YAML.

### Decision
Use `gopkg.in/yaml.v3` as a primary third-party dependency. It provides substantial value since Go's standard library has no built-in YAML parser.
