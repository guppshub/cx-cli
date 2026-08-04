package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/guppshub/cx-cli/internal/config"
)

// ECSCache represents the root structure of ecs_cache.json.
type ECSCache struct {
	Workspaces map[string]*WorkspaceCache `json:"workspaces"`
}

// WorkspaceCache contains cached clusters and services for a workspace.
type WorkspaceCache struct {
	Clusters    []ECSCluster            `json:"clusters"`
	Services    map[string][]ECSService `json:"services"` // key is cluster name
	LastUpdated time.Time               `json:"last_updated"`
}

// CachePath resolves the absolute path to ecs_cache.json.
func CachePath() (string, error) {
	cPath, err := config.Path()
	if err != nil {
		return "", fmt.Errorf("resolving config path for cache: %w", err)
	}
	return filepath.Join(filepath.Dir(cPath), "ecs_cache.json"), nil
}

// LoadCache loads and parses the ecs_cache.json file.
func LoadCache() (*ECSCache, error) {
	path, err := CachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ECSCache{Workspaces: make(map[string]*WorkspaceCache)}, nil
		}
		return nil, fmt.Errorf("reading ecs cache: %w", err)
	}

	var cache ECSCache
	if err := json.Unmarshal(data, &cache); err != nil {
		// If cache is corrupted, return a clean empty cache instead of crashing
		return &ECSCache{Workspaces: make(map[string]*WorkspaceCache)}, nil
	}

	if cache.Workspaces == nil {
		cache.Workspaces = make(map[string]*WorkspaceCache)
	}
	return &cache, nil
}

// SaveCache writes the ECS cache to disk atomically.
func SaveCache(cache *ECSCache) error {
	path, err := CachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling ecs cache to JSON: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	// Atomic write
	tmpFile, err := os.CreateTemp(dir, "ecs_cache.*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary ecs cache file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("writing temporary ecs cache file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing temporary ecs cache file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temporary ecs cache file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("moving temporary ecs cache file to destination: %w", err)
	}

	return nil
}
