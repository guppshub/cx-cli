# Data Model: Workspace Management

This document defines the data structures and structures mapping to workspaces.

## 1. Domain Entities

```go
package workspace

// Workspace represents a named operational environment.
type Workspace struct {
	Name     string         `json:"name"`
	Provider string         `json:"provider"`
	Config   map[string]any `json:"config"`
}

// WorkspaceSummary represents a summary of a workspace used for listings.
type WorkspaceSummary struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	IsActive bool   `json:"is_active"`
}
```

## 2. Configuration Integration
Workspaces map to the `contexts` block of the configuration model. To support this:
- **`current`** in `config.yaml` tracks the active workspace name.
- **`contexts`** map contains the list of workspaces.
- The `Config` struct is updated to include the `current` field:
  ```go
  type Config struct {
      Version     string             `yaml:"version"`
      Current     string             `yaml:"current"`
      Contexts    map[string]*Context `yaml:"contexts"`
      Preferences map[string]string  `yaml:"preferences"`
  }
  ```
