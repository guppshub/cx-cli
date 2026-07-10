package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader manages loading and saving of the configuration file.
type Loader struct {
	path string
}

// New creates a new Loader instance targeting the specified file path.
func New(path string) *Loader {
	return &Loader{path: path}
}

// Default returns a new, valid Config struct populated with default values.
func Default() *Config {
	return &Config{
		Version:     Version,
		Contexts:    make(map[string]*Context),
		Preferences: make(map[string]string),
	}
}

// Load reads and parses the configuration file.
// If the file does not exist, it returns the default configuration.
func (l *Loader) Load() (*Config, error) {
	data, err := os.ReadFile(l.path)
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
func (l *Loader) Save(cfg *Config) error {
	if err := Validate(cfg); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling configuration to YAML: %w", err)
	}

	dir := filepath.Dir(l.path)
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
		_ = os.Remove(tmpPath)
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

	if err := os.Rename(tmpPath, l.path); err != nil {
		return fmt.Errorf("moving temporary configuration file to destination: %w", err)
	}

	return nil
}

// Load reads and parses the default configuration file from its resolved path.
func Load() (*Config, error) {
	cPath, err := Path()
	if err != nil {
		return nil, fmt.Errorf("resolving configuration path: %w", err)
	}
	return New(cPath).Load()
}

// Save writes the given configuration back to disk atomically at its resolved path.
func Save(cfg *Config) error {
	cPath, err := Path()
	if err != nil {
		return fmt.Errorf("resolving configuration path: %w", err)
	}
	return New(cPath).Save(cfg)
}
