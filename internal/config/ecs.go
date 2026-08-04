package config

import "fmt"

// ECSConfig represents the ecs settings block under a workspace in config.yaml.
type ECSConfig struct {
	Cache          bool   `yaml:"cache"`
	DefaultCluster string `yaml:"default_cluster"`
}

// GetECSConfig extracts and parses the ECS settings from a workspace.
func GetECSConfig(ws *Workspace) (*ECSConfig, error) {
	if ws == nil {
		return &ECSConfig{}, nil
	}

	rawECS, exists := ws.Raw["ecs"]
	if !exists {
		return &ECSConfig{}, nil
	}

	ecsMap, ok := rawECS.(map[string]any)
	if !ok {
		// Handle map[any]any unmarshaling fallback
		if m, ok := rawECS.(map[any]any); ok {
			ecsMap = make(map[string]any)
			for k, v := range m {
				ecsMap[fmt.Sprint(k)] = v
			}
		} else {
			return &ECSConfig{}, nil
		}
	}

	var cfg ECSConfig
	if cacheVal, ok := ecsMap["cache"].(bool); ok {
		cfg.Cache = cacheVal
	}
	if clusterVal, ok := ecsMap["default_cluster"].(string); ok {
		cfg.DefaultCluster = clusterVal
	}

	return &cfg, nil
}

// SetECSConfig writes the ECS settings back into the workspace map.
func SetECSConfig(ws *Workspace, cfg *ECSConfig) {
	if ws == nil {
		return
	}
	if ws.Raw == nil {
		ws.Raw = make(map[string]any)
	}

	ecsMap := make(map[string]any)
	ecsMap["cache"] = cfg.Cache
	if cfg.DefaultCluster != "" {
		ecsMap["default_cluster"] = cfg.DefaultCluster
	}

	ws.Raw["ecs"] = ecsMap
}
