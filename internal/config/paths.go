package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Path returns the absolute path to the configuration file (config.yaml).
func Path() (string, error) {
	if envPath := os.Getenv("CX_CONFIG"); envPath != "" {
		return envPath, nil
	}

	var baseDir string
	if runtime.GOOS == "windows" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolving windows appdata directory: %w", err)
		}
		baseDir = dir
	} else {
		// macOS & Linux: ~/.config
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving user home directory: %w", err)
		}
		baseDir = filepath.Join(home, ".config")
	}

	return filepath.Join(baseDir, "cx", "config.yaml"), nil
}
