package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

var (
	configPathOverride string
	statePathOverride  string
)

// Config represents the application configuration format (config.yaml).
type Config struct {
	Version     string              `yaml:"version"`
	Contexts    map[string]*Context `yaml:"contexts"`
	Preferences map[string]string   `yaml:"preferences"`
}

// Context represents an environment context configuration.
type Context struct {
	Provider  string     `yaml:"provider"`
	AWS       *AWSConfig `yaml:"aws,omitempty"`
	Resources *Resources `yaml:"resources,omitempty"`
}

// AWSConfig represents provider-specific AWS configuration.
type AWSConfig struct {
	Profile string `yaml:"profile"`
	Region  string `yaml:"region"`
}

// Resources represents available environment resources.
type Resources struct {
	Databases []DatabaseResource `yaml:"databases,omitempty"`
	Caches    []CacheResource    `yaml:"caches,omitempty"`
}

// DatabaseResource represents a database resource configuration.
type DatabaseResource struct {
	Name              string `yaml:"name"`
	Engine            string `yaml:"engine"`
	Endpoint          string `yaml:"endpoint"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
}

// CacheResource represents a cache resource configuration.
type CacheResource struct {
	Name              string `yaml:"name"`
	Endpoint          string `yaml:"endpoint"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
}

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
}

// Path returns the absolute path to the configuration file (config.yaml).
func Path() (string, error) {
	if configPathOverride != "" {
		return configPathOverride, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, ".config", "cx", "config.yaml"), nil
	}
	return filepath.Join(dir, "cx", "config.yaml"), nil
}

// StatePath returns the absolute path to the runtime state file (state.json).
func StatePath() (string, error) {
	if statePathOverride != "" {
		return statePathOverride, nil
	}
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return filepath.Join(stateHome, "cx", "state.json"), nil
	}

	switch runtime.GOOS {
	case "windows":
		dir, err := os.UserCacheDir() // Returns %LOCALAPPDATA%
		if err != nil {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolving home directory: %w", err)
			}
			return filepath.Join(home, "AppData", "Local", "cx", "state.json"), nil
		}
		return filepath.Join(dir, "cx", "state.json"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", "cx", "state", "state.json"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, ".local", "state", "cx", "state.json"), nil
	}
}

// Default returns a new, valid Config struct populated with default values.
func Default() *Config {
	return &Config{
		Version:     "1",
		Contexts:    make(map[string]*Context),
		Preferences: make(map[string]string),
	}
}

// Validate checks the configuration for correctness.
func Validate(cfg *Config) error {
	if cfg.Version != "1" && cfg.Version != "1.0" {
		return fmt.Errorf("unsupported configuration version %q", cfg.Version)
	}
	return nil
}

// Load reads and parses the configuration file from disk.
// If the file does not exist, it returns the default configuration without creating it.
func Load() (*Config, error) {
	cPath, err := Path()
	if err != nil {
		return nil, fmt.Errorf("resolving configuration path: %w", err)
	}

	data, err := os.ReadFile(cPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("reading configuration file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing configuration YAML: %w", err)
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	return cfg, nil
}

// Save writes the given configuration back to disk atomically.
func Save(cfg *Config) error {
	if err := Validate(cfg); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	cPath, err := Path()
	if err != nil {
		return fmt.Errorf("resolving configuration path: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling configuration to YAML: %w", err)
	}

	dir := filepath.Dir(cPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating configuration directory: %w", err)
	}

	// Atomic write
	tmpFile, err := os.CreateTemp(dir, "config.*.yaml.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary configuration file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath) // Cleanup tmp file if rename fails
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("writing temporary configuration file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing temporary configuration file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temporary configuration file: %w", err)
	}

	if err := os.Rename(tmpPath, cPath); err != nil {
		return fmt.Errorf("moving temporary configuration file to destination: %w", err)
	}

	return nil
}
