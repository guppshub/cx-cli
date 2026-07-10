package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Path returns the absolute path to the configuration file (config.yaml).
func Path() (string, error) {
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
