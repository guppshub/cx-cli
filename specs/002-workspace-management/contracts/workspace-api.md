# API Contract: Workspace Management

This contract defines the public Go API exposed by the workspace package.

## Go Interface/Function Signatures

```go
package workspace

import "github.com/guppshub/cx-cli/internal/config"

// Manager coordinates workspace management and persistence.
type Manager struct {
	store *config.Store
}

// New creates a new Manager instance.
func New(store *config.Store) *Manager

// Add adds a new workspace.
// Returns an error if the name already exists or is invalid.
func (m *Manager) Add(name string, provider string, providerConfig map[string]any) error

// Delete removes a workspace.
// Returns an error if the workspace is active or does not exist.
func (m *Manager) Delete(name string) error

// Rename changes a workspace name.
// Updates active workspace pointer automatically if the active workspace is renamed.
func (m *Manager) Rename(oldName, newName string) error

// Use sets the active workspace.
// Returns an error if the workspace does not exist.
func (m *Manager) Use(name string) error

// Current returns the currently active workspace.
// Returns an error if no active workspace is selected or found.
func (m *Manager) Current() (*Workspace, error)

// List returns a sorted list of all workspaces and their active status.
func (m *Manager) List() ([]WorkspaceSummary, error)

// Get retrieves a workspace by name.
// Returns an error if the workspace does not exist.
func (m *Manager) Get(name string) (*Workspace, error)
```

## Behavior Specifications

### 1. Active Workspace Protection
- Deleting the currently active workspace (`m.Current()`) is rejected with a descriptive error.
- Renaming the active workspace changes the active pointer (`current` in configuration) to `newName` and saves successfully.

### 2. Opaque Configuration Storage
- `providerConfig` is mapped directly to the `Context.Raw` field of the configuration package and saved without parsing or validating the values.

### 3. Verification of Empty State
- If there are no workspaces, `List()` returns an empty slice (not `nil`), and `Current()` returns an error indicating no active workspace exists.
