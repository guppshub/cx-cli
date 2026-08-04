package config

import "fmt"

// EC2Config represents the ec2 settings block under a workspace in config.yaml.
type EC2Config struct {
	Cache bool `yaml:"cache"`
}

// GetEC2Config extracts and parses the EC2 settings from a workspace.
func GetEC2Config(ws *Workspace) (*EC2Config, error) {
	if ws == nil {
		return &EC2Config{}, nil
	}

	rawEC2, exists := ws.Raw["ec2"]
	if !exists {
		return &EC2Config{}, nil
	}

	ec2Map, ok := rawEC2.(map[string]any)
	if !ok {
		// Handle map[any]any unmarshaling fallback
		if m, ok := rawEC2.(map[any]any); ok {
			ec2Map = make(map[string]any)
			for k, v := range m {
				ec2Map[fmt.Sprint(k)] = v
			}
		} else {
			return &EC2Config{}, nil
		}
	}

	var cfg EC2Config
	if cacheVal, ok := ec2Map["cache"].(bool); ok {
		cfg.Cache = cacheVal
	}

	return &cfg, nil
}

// SetEC2Config writes the EC2 settings back into the workspace map.
func SetEC2Config(ws *Workspace, cfg *EC2Config) {
	if ws == nil {
		return
	}
	if ws.Raw == nil {
		ws.Raw = make(map[string]any)
	}

	ec2Map := make(map[string]any)
	ec2Map["cache"] = cfg.Cache

	ws.Raw["ec2"] = ec2Map
}
