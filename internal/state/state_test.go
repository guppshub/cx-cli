package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	sPath, err := Path()
	if err != nil {
		t.Fatalf("unexpected error from Path(): %v", err)
	}
	if filepath.Base(sPath) != "state.json" {
		t.Errorf("expected state file name to be state.json, got %s", filepath.Base(sPath))
	}
}

func TestDefault(t *testing.T) {
	s := Default()
	if s.CurrentContext != "" {
		t.Errorf("expected empty current context, got %q", s.CurrentContext)
	}
	if s.ActiveConnections == nil || len(s.ActiveConnections) != 0 {
		t.Errorf("expected empty active connections map, got %v", s.ActiveConnections)
	}
}

func TestManagerLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	manager := New(statePath)

	// 1. Load when file does not exist (should return Default and not create file)
	s, err := manager.Load()
	if err != nil {
		t.Fatalf("unexpected error loading non-existent state: %v", err)
	}
	if s.CurrentContext != "" {
		t.Errorf("expected empty current context, got %q", s.CurrentContext)
	}
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Errorf("expected state file NOT to be created on load, but os.Stat returned: %v", err)
	}

	// 2. Save a custom state
	s.CurrentContext = "staging"
	s.ActiveConnections["staging/database/mercury"] = &ConnectionMetadata{
		Type:         "database",
		Name:         "mercury",
		LocalPort:    5432,
		ConnectionID: "cx-conn-staging-db-mercury",
		ConnectedAt:  "2026-07-11T00:25:00Z",
	}

	err = manager.Save(s)
	if err != nil {
		t.Fatalf("unexpected error saving state: %v", err)
	}

	// Verify file is created
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("expected state file to be created on save, but got error: %v", err)
	}

	// 3. Load the saved state
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("unexpected error loading saved state: %v", err)
	}
	if loaded.CurrentContext != "staging" {
		t.Errorf("expected current context to be 'staging', got %q", loaded.CurrentContext)
	}
	conn, exists := loaded.ActiveConnections["staging/database/mercury"]
	if !exists {
		t.Fatal("expected staging connection metadata to exist")
	}
	if conn.Name != "mercury" || conn.LocalPort != 5432 || conn.ConnectionID != "cx-conn-staging-db-mercury" {
		t.Errorf("unexpected connection metadata parsed: %+v", conn)
	}

	// 4. Load malformed JSON
	err = os.WriteFile(statePath, []byte("{invalid_json:"), 0644)
	if err != nil {
		t.Fatalf("failed to write malformed state file: %v", err)
	}

	_, err = manager.Load()
	if err == nil {
		t.Error("expected loading malformed JSON to fail, but it succeeded")
	}
}
