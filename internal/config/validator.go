package config

import (
	"fmt"
)

// Version defines the supported configuration schema version.
const Version = "1"

// Validate checks the configuration for schema and structure correctness.
func Validate(cfg *Config) error {
	if cfg.Version != Version && cfg.Version != "1.0" {
		return fmt.Errorf("unsupported configuration version %q", cfg.Version)
	}
	return nil
}
