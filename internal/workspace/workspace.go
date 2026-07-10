package workspace

import (
	"fmt"
	"sort"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/errors"
)

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

// Manager coordinates workspace management and persistence.
type Manager struct {
	store *config.Store
}

// New creates a new Manager instance.
func New(store *config.Store) *Manager {
	return &Manager{store: store}
}

// Add adds a new workspace.
// Returns ErrDuplicateWorkspace if the workspace already exists.
func (m *Manager) Add(name string, provider string, providerConfig map[string]any) error {
	if name == "" {
		return fmt.Errorf("workspace name cannot be empty")
	}
	if provider == "" {
		return fmt.Errorf("workspace provider cannot be empty")
	}

	cfg, err := m.store.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if _, exists := cfg.Workspaces[name]; exists {
		return fmt.Errorf("%w: workspace %q", errors.ErrDuplicateWorkspace, name)
	}

	cfg.Workspaces[name] = &config.Workspace{
		Provider: provider,
		Raw:      providerConfig,
	}

	if err := m.store.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

// Delete removes a workspace.
// Returns ErrWorkspaceNotFound if the workspace does not exist.
// Returns ErrWorkspaceActive if trying to delete the currently active workspace.
func (m *Manager) Delete(name string) error {
	cfg, err := m.store.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if _, exists := cfg.Workspaces[name]; !exists {
		return fmt.Errorf("%w: workspace %q", errors.ErrWorkspaceNotFound, name)
	}

	if cfg.Current == name {
		return fmt.Errorf("%w: cannot delete active workspace %q", errors.ErrWorkspaceActive, name)
	}

	delete(cfg.Workspaces, name)

	if err := m.store.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

// Rename changes a workspace name.
// Updates active workspace pointer automatically if the active workspace is renamed.
// Returns ErrWorkspaceNotFound if the old workspace does not exist.
// Returns ErrDuplicateWorkspace if the new workspace name already exists.
func (m *Manager) Rename(oldName, newName string) error {
	if newName == "" {
		return fmt.Errorf("new workspace name cannot be empty")
	}

	cfg, err := m.store.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ws, exists := cfg.Workspaces[oldName]
	if !exists {
		return fmt.Errorf("%w: workspace %q", errors.ErrWorkspaceNotFound, oldName)
	}

	if _, occupied := cfg.Workspaces[newName]; occupied {
		return fmt.Errorf("%w: workspace %q", errors.ErrDuplicateWorkspace, newName)
	}

	cfg.Workspaces[newName] = ws
	delete(cfg.Workspaces, oldName)

	if cfg.Current == oldName {
		cfg.Current = newName
	}

	if err := m.store.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

// Use sets the active workspace.
// Returns ErrWorkspaceNotFound if the workspace does not exist.
func (m *Manager) Use(name string) error {
	cfg, err := m.store.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if _, exists := cfg.Workspaces[name]; !exists {
		return fmt.Errorf("%w: workspace %q", errors.ErrWorkspaceNotFound, name)
	}

	cfg.Current = name

	if err := m.store.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

// Current returns the currently active workspace.
// Returns ErrWorkspaceNotFound if no active workspace is selected or found.
func (m *Manager) Current() (*Workspace, error) {
	cfg, err := m.store.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if cfg.Current == "" {
		return nil, fmt.Errorf("%w: no active workspace selected", errors.ErrWorkspaceNotFound)
	}

	ws, exists := cfg.Workspaces[cfg.Current]
	if !exists {
		return nil, fmt.Errorf("%w: active workspace %q not found in workspaces", errors.ErrWorkspaceNotFound, cfg.Current)
	}

	return &Workspace{
		Name:     cfg.Current,
		Provider: ws.Provider,
		Config:   ws.Raw,
	}, nil
}

// List returns a sorted list of all workspaces and their active status.
func (m *Manager) List() ([]WorkspaceSummary, error) {
	cfg, err := m.store.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	names := make([]string, 0, len(cfg.Workspaces))
	for name := range cfg.Workspaces {
		names = append(names, name)
	}
	sort.Strings(names)

	summaries := make([]WorkspaceSummary, 0, len(names))
	for _, name := range names {
		ws := cfg.Workspaces[name]
		summaries = append(summaries, WorkspaceSummary{
			Name:     name,
			Provider: ws.Provider,
			IsActive: name == cfg.Current,
		})
	}

	return summaries, nil
}

// Get retrieves a workspace by name.
// Returns ErrWorkspaceNotFound if the workspace does not exist.
func (m *Manager) Get(name string) (*Workspace, error) {
	cfg, err := m.store.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	ws, exists := cfg.Workspaces[name]
	if !exists {
		return nil, fmt.Errorf("%w: workspace %q", errors.ErrWorkspaceNotFound, name)
	}

	return &Workspace{
		Name:     name,
		Provider: ws.Provider,
		Config:   ws.Raw,
	}, nil
}
