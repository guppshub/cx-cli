package workspace

import (
	"path/filepath"
	"testing"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/errors"
)

func TestAddAndList(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	store := config.New(configPath)
	mgr := New(store)

	// Add workspaces out of alphabetical order
	err := mgr.Add("staging", "aws", map[string]any{"region": "us-east-1"})
	if err != nil {
		t.Fatalf("unexpected error adding staging: %v", err)
	}

	err = mgr.Add("production", "aws", map[string]any{"region": "us-west-2"})
	if err != nil {
		t.Fatalf("unexpected error adding production: %v", err)
	}

	// Add duplicate workspace name
	err = mgr.Add("staging", "aws", map[string]any{"region": "us-east-1"})
	if !errors.Is(err, errors.ErrDuplicateWorkspace) {
		t.Errorf("expected ErrDuplicateWorkspace, got %v", err)
	}

	// Verify deterministic listing order
	list, err := mgr.List()
	if err != nil {
		t.Fatalf("unexpected error listing: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(list))
	}

	// Should be alphabetically ordered: production, staging
	if list[0].Name != "production" {
		t.Errorf("expected first workspace to be 'production', got %q", list[0].Name)
	}
	if list[1].Name != "staging" {
		t.Errorf("expected second workspace to be 'staging', got %q", list[1].Name)
	}
}

func TestUseAndCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	store := config.New(configPath)
	mgr := New(store)

	// Retrieve current workspace when none exist or active is empty
	_, err := mgr.Current()
	if !errors.Is(err, errors.ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}

	// Add a workspace
	err = mgr.Add("staging", "aws", map[string]any{"region": "us-east-1"})
	if err != nil {
		t.Fatalf("unexpected error adding staging: %v", err)
	}

	// Retrieve current workspace when none is active
	_, err = mgr.Current()
	if !errors.Is(err, errors.ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}

	// Select the workspace
	err = mgr.Use("staging")
	if err != nil {
		t.Fatalf("unexpected error selecting staging: %v", err)
	}

	// Retrieve current active workspace
	curr, err := mgr.Current()
	if err != nil {
		t.Fatalf("unexpected error retrieving current active workspace: %v", err)
	}
	if curr.Name != "staging" {
		t.Errorf("expected current name to be 'staging', got %q", curr.Name)
	}
	if curr.Provider != "aws" {
		t.Errorf("expected current provider to be 'aws', got %q", curr.Provider)
	}

	// Use non-existent workspace
	err = mgr.Use("production")
	if !errors.Is(err, errors.ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}
}

func TestRenameAndDelete(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	store := config.New(configPath)
	mgr := New(store)

	err := mgr.Add("staging", "aws", map[string]any{"region": "us-east-1"})
	if err != nil {
		t.Fatalf("unexpected error adding staging: %v", err)
	}

	err = mgr.Add("production", "aws", map[string]any{"region": "us-west-2"})
	if err != nil {
		t.Fatalf("unexpected error adding production: %v", err)
	}

	// Select production
	err = mgr.Use("production")
	if err != nil {
		t.Fatalf("unexpected error selecting production: %v", err)
	}

	// Delete active workspace production (should fail)
	err = mgr.Delete("production")
	if !errors.Is(err, errors.ErrWorkspaceActive) {
		t.Errorf("expected ErrWorkspaceActive, got %v", err)
	}

	// Delete non-active workspace staging (should succeed)
	err = mgr.Delete("staging")
	if err != nil {
		t.Fatalf("unexpected error deleting staging: %v", err)
	}

	// Verify staging is removed from list
	list, err := mgr.List()
	if err != nil {
		t.Fatalf("unexpected error listing: %v", err)
	}
	if len(list) != 1 || list[0].Name != "production" {
		t.Errorf("expected only 'production' to remain, got %v", list)
	}

	// Rename active workspace production -> prod
	err = mgr.Rename("production", "prod")
	if err != nil {
		t.Fatalf("unexpected error renaming production to prod: %v", err)
	}

	// Verify active workspace is now prod
	curr, err := mgr.Current()
	if err != nil {
		t.Fatalf("unexpected error getting current active: %v", err)
	}
	if curr.Name != "prod" {
		t.Errorf("expected current to be 'prod' after rename, got %q", curr.Name)
	}

	// Verify old name production does not exist
	_, err = mgr.Get("production")
	if !errors.Is(err, errors.ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}
}
