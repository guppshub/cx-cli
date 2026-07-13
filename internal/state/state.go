package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// State represents the machine-managed runtime state (state.json).
type State struct {
	CurrentContext    string                         `json:"current_context"`
	ActiveConnections map[string]*ConnectionMetadata `json:"active_connections"`
}

// ConnectionMetadata represents metadata for active background connections.
type ConnectionMetadata struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	LocalPort    int    `json:"local_port"`
	ConnectionID string `json:"connection_id"`
	ConnectedAt  string `json:"connected_at"`
	Pid          int    `json:"pid"`
	State        string `json:"state,omitempty"`
	Restarts     int    `json:"restarts,omitempty"`
	LastFailure  string `json:"last_failure,omitempty"`
	LastRestart  string `json:"last_restart,omitempty"`
	Profile      string `json:"profile,omitempty"`
	Region       string `json:"region,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
}

// Manager manages loading and saving of the runtime state file.
type Manager struct {
	path string
}

// New creates a new Manager instance targeting the specified file path.
func New(path string) *Manager {
	return &Manager{path: path}
}

// Path returns the absolute path to the runtime state file (state.json).
func Path() (string, error) {
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return filepath.Join(stateHome, "cx", "state.json"), nil
	}

	if runtime.GOOS == "windows" {
		dir, err := os.UserCacheDir() // Returns %LOCALAPPDATA%
		if err != nil {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolving home directory: %w", err)
			}
			return filepath.Join(home, "AppData", "Local", "cx", "state.json"), nil
		}
		return filepath.Join(dir, "cx", "state.json"), nil
	}

	// macOS and Linux/Unix: ~/.local/state/cx/state.json
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".local", "state", "cx", "state.json"), nil
}

// Default returns a new, valid State struct populated with default values.
func Default() *State {
	return &State{
		CurrentContext:    "",
		ActiveConnections: make(map[string]*ConnectionMetadata),
	}
}

// Load reads and parses the state file from disk.
// If the file does not exist, it returns the default state without creating it.
func (m *Manager) Load() (*State, error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	s := &State{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing state JSON: %w", err)
	}

	if s.ActiveConnections == nil {
		s.ActiveConnections = make(map[string]*ConnectionMetadata)
	}

	return s, nil
}

// Save writes the given state back to disk atomically.
func (m *Manager) Save(s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state to JSON: %w", err)
	}

	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	// Atomic write
	tmpFile, err := os.CreateTemp(dir, "state.*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary state file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("writing temporary state file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing temporary state file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temporary state file: %w", err)
	}

	if err := os.Rename(tmpPath, m.path); err != nil {
		return fmt.Errorf("moving temporary state file to destination: %w", err)
	}

	return nil
}

// Load reads and parses the default state file from its resolved path.
func Load() (*State, error) {
	sPath, err := Path()
	if err != nil {
		return nil, fmt.Errorf("resolving state path: %w", err)
	}
	return New(sPath).Load()
}

// Save writes the given state back to disk atomically at its resolved path.
func Save(s *State) error {
	sPath, err := Path()
	if err != nil {
		return fmt.Errorf("resolving state path: %w", err)
	}
	return New(sPath).Save(s)
}
