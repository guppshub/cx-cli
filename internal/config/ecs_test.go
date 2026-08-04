package config

import "testing"

func TestGetSetECSConfig(t *testing.T) {
	ws := &Workspace{
		Provider: "aws",
		Raw:      make(map[string]any),
	}

	// 1. Get empty config
	cfg, err := GetECSConfig(ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Cache {
		t.Error("expected cache to be disabled by default")
	}
	if cfg.DefaultCluster != "" {
		t.Errorf("expected default cluster to be empty, got %q", cfg.DefaultCluster)
	}

	// 2. Set config
	newCfg := &ECSConfig{
		Cache:          true,
		DefaultCluster: "dev-cluster",
	}
	SetECSConfig(ws, newCfg)

	// 3. Get updated config
	cfg, err = GetECSConfig(ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Cache {
		t.Error("expected cache to be enabled")
	}
	if cfg.DefaultCluster != "dev-cluster" {
		t.Errorf("expected default cluster to be 'dev-cluster', got %q", cfg.DefaultCluster)
	}
}
