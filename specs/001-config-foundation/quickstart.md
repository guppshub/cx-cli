# Quickstart: Configuration Foundation Validation Guide

This document describes how to validate the Configuration Foundation subsystem.

## Prerequisites
- Go installed (`1.22.5` or later)
- Access to terminal

## Validation Scenarios

### Scenario 1: Run Unit Tests
To run all unit tests verifying the directory path resolution, loaders, savers, and validation engine:

```bash
# Execute tests in the config package
go test -v -race ./internal/config/...
```

**Expected Outcome:**
- All tests pass successfully (no failures).
- Output reports test coverage.

### Scenario 2: Validation of Default Behavior (No Config File)
To verify that the system returns default settings when no config exists:
1. Ensure `~/.config/cx/config.yaml` does not exist (or backup if it does).
2. Execute the verification tool (a simple scratch/test script):
   ```bash
   go run specs/001-config-foundation/scratch/verify.go --action=load
   ```
3. **Expected Output:**
   - Configuration is loaded with default values (version: `"1"`, contexts: empty, preferences: empty).
   - No `config.yaml` is created on disk.

### Scenario 3: Validation of Version Check
To verify that unsupported configuration versions are correctly rejected:
1. Create a `config.yaml` in the configuration path containing `version: "2"`.
2. Execute the verification tool:
   ```bash
   go run specs/001-config-foundation/scratch/verify.go --action=load
   ```
3. **Expected Output:**
   - Command fails with: `loading configuration: unsupported configuration version "2"`.
