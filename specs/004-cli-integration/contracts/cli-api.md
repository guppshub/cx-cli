# API Contract: CLI Integration

This contract defines the CLI command schema and the Resource Resolver helper.

## 1. Resource Resolver contract

```go
package resource

import "github.com/guppshub/cx-cli/internal/config"

// ResolveDatabase parses the active workspace and resolves the database resource by name.
// Returns an error if the database resource is not found or config is malformed.
func ResolveDatabase(workspace *config.Workspace, name string) (*DatabaseResource, error)
```

## 2. CLI Command Hierarchy

### Use Command
- **Command Syntax**: `cx use <workspace>`
- **Description**: Sets the active workspace context in `config.yaml`.
- **Preconditions**: Target workspace name must exist in configuration.

### DB Command
- **Command Syntax**: `cx db <resource>`
- **Flags**:
  - `--port`, `-p` (int): Override the local port to forward to (default: `local_port` from resource config, or `5432` fallback).
- **Description**: Resolves the database configuration, triggers authentication, and spins up the port forwarding connection listener in the foreground.
- **Preconditions**:
  - An active workspace must be set (using `cx use`).
  - Target database resource must exist in the active workspace.
