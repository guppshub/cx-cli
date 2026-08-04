package aws

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadSaveCache(t *testing.T) {
	// Set CX_CONFIG to a temp file path to isolate the test from the user's config
	tmpDir, err := os.MkdirTemp("", "cx_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := filepath.Join(tmpDir, "config.yaml")
	_ = os.Setenv("CX_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("CX_CONFIG") }()

	// 1. Load empty cache (should succeed and return empty structure)
	cache, err := LoadCache()
	if err != nil {
		t.Fatalf("failed to load empty cache: %v", err)
	}
	if len(cache.Workspaces) != 0 {
		t.Errorf("expected empty workspaces, got %d", len(cache.Workspaces))
	}

	// 2. Modify cache
	wsCache := &WorkspaceCache{
		Clusters: []ECSCluster{
			{Name: "dev-cluster", ARN: "arn:aws:ecs:us-east-1:1234:cluster/dev-cluster"},
		},
		Services: map[string][]ECSService{
			"dev-cluster": {
				{Name: "web-app", ARN: "arn:aws:ecs:us-east-1:1234:service/dev-cluster/web-app"},
			},
		},
		LastUpdated: time.Now(),
	}
	cache.Workspaces["dev"] = wsCache

	// 3. Save cache
	err = SaveCache(cache)
	if err != nil {
		t.Fatalf("failed to save cache: %v", err)
	}

	// 4. Reload cache and assert
	loaded, err := LoadCache()
	if err != nil {
		t.Fatalf("failed to load cache: %v", err)
	}

	wsLoaded, ok := loaded.Workspaces["dev"]
	if !ok {
		t.Fatal("expected workspace 'dev' to exist in loaded cache")
	}

	if len(wsLoaded.Clusters) != 1 || wsLoaded.Clusters[0].Name != "dev-cluster" {
		t.Errorf("loaded cluster mismatch: %+v", wsLoaded.Clusters)
	}

	services, ok := wsLoaded.Services["dev-cluster"]
	if !ok || len(services) != 1 || services[0].Name != "web-app" {
		t.Errorf("loaded service mismatch: %+v", services)
	}
}
