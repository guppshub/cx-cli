package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	cPath, err := Path()
	if err != nil {
		t.Fatalf("unexpected error from Path(): %v", err)
	}
	if filepath.Base(cPath) != "config.yaml" {
		t.Errorf("expected config file name to be config.yaml, got %s", filepath.Base(cPath))
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Version != Version {
		t.Errorf("expected default version to be %q, got %q", Version, cfg.Version)
	}
	if cfg.Workspaces == nil || len(cfg.Workspaces) != 0 {
		t.Errorf("expected empty workspaces map, got %v", cfg.Workspaces)
	}
	if cfg.Preferences == nil || len(cfg.Preferences) != 0 {
		t.Errorf("expected empty preferences map, got %v", cfg.Preferences)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"Valid version 1", "1", false},
		{"Valid version 1.0", "1.0", false},
		{"Invalid version empty", "", true},
		{"Invalid version 2", "2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Version: tt.version,
			}
			err := Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStoreLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	store := New(configPath)

	// 1. Load when file does not exist (should return Default and not create file)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error loading non-existent config: %v", err)
	}
	if cfg.Version != Version {
		t.Errorf("expected version to be %q, got %q", Version, cfg.Version)
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Errorf("expected config file NOT to be created on load, but os.Stat returned: %v", err)
	}

	// 2. Save a custom config
	cfg.Version = "1.0"
	cfg.Preferences["theme"] = "dark"

	// Add workspace with raw generic map
	cfg.Workspaces["staging"] = &Workspace{
		Provider: "aws",
		Raw: map[string]any{
			"profile": "staging-admin",
			"region":  "us-east-1",
		},
	}

	err = store.Save(cfg)
	if err != nil {
		t.Fatalf("unexpected error saving config: %v", err)
	}

	// Verify file is created
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file to be created on save, but got error: %v", err)
	}

	// 3. Load the saved config
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error loading saved config: %v", err)
	}
	if loaded.Version != "1.0" {
		t.Errorf("expected version to be '1.0', got %q", loaded.Version)
	}
	if loaded.Preferences["theme"] != "dark" {
		t.Errorf("expected preference theme to be 'dark', got %q", loaded.Preferences["theme"])
	}
	ws, exists := loaded.Workspaces["staging"]
	if !exists {
		t.Fatal("expected staging workspace to exist")
	}
	if ws.Provider != "aws" {
		t.Errorf("expected provider to be 'aws', got %q", ws.Provider)
	}
	if ws.Raw["profile"] != "staging-admin" {
		t.Errorf("expected profile to be 'staging-admin', got %v", ws.Raw["profile"])
	}

	// 4. Save invalid config (should fail validation)
	loaded.Version = "3.0"
	err = store.Save(loaded)
	if err == nil {
		t.Error("expected saving version 3.0 to fail, but it succeeded")
	}

	// 5. Load malformed configuration file
	err = os.WriteFile(configPath, []byte("version: 1\ninvalid_yaml: [}"), 0644)
	if err != nil {
		t.Fatalf("failed to write malformed config file: %v", err)
	}

	_, err = store.Load()
	if err == nil {
		t.Error("expected loading malformed yaml to fail, but it succeeded")
	}
}
