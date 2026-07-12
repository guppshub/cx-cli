package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInit(t *testing.T) {
	// Create temporary directory for configuration path
	tmpDir, err := os.MkdirTemp("", "cx-init-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpPath := filepath.Join(tmpDir, "config.yaml")

	// 1. Test clean initialization (file does not exist)
	err = runInit(tmpPath, false)
	if err != nil {
		t.Fatalf("expected successful initialization on empty path, got error: %v", err)
	}

	// Verify the file exists and is not empty
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("failed to read created configuration: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected configuration file to contain template data, but it is empty")
	}

	// 2. Test duplicate initialization check (file already exists, force = false)
	err = runInit(tmpPath, false)
	if err == nil {
		t.Error("expected init to fail when configuration already exists and force is false, but got no error")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected duplicate error message, got: %q", err.Error())
	}

	// 3. Test force overwrite (file already exists, force = true)
	err = runInit(tmpPath, true)
	if err != nil {
		t.Fatalf("expected successful override, got error: %v", err)
	}
}
