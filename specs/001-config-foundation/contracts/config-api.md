# API Contract: Configuration Foundation

This contract defines the public API exposed by the configuration subsystem.

## Go Interface/Function Signatures

```go
package config

// Path returns the absolute path to the configuration file (config.yaml).
// It resolves the path based on the operating system convention.
func Path() (string, error)

// StatePath returns the absolute path to the runtime state file (state.json).
// It resolves the path based on the operating system convention.
func StatePath() (string, error)

// Load reads and parses the configuration file from disk.
// If the file does not exist, it returns the default configuration.
// It does NOT create the configuration file on disk.
func Load() (*Config, error)

// Save writes the given configuration back to disk atomically.
// It creates any necessary parent directories.
func Save(cfg *Config) error

// Default returns a new, valid configuration struct populated with default values.
func Default() *Config

// Validate checks the configuration for correctness (e.g. version checks, struct validity).
func Validate(cfg *Config) error
```

## Behavior Specifications

### 1. Versioning
- Only `"1"` and `"1.0"` are accepted as valid version fields.
- Any other value will cause `Validate()` and `Load()` to fail with an error.

### 2. Atomic Writes
- `Save()` should write to a temporary file in the same directory (e.g., `config.yaml.tmp`) and then atomically rename the file to `config.yaml`.
- This prevents files from getting corrupted during power cuts or system terminations.

### 3. File Creation
- `Load()` must be read-only and never write or create files on disk.
