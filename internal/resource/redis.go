package resource

import (
	"fmt"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/errors"
	"gopkg.in/yaml.v3"
)

// RedisResource represents the configuration for a target Redis resource.
type RedisResource struct {
	Name              string `yaml:"name"`
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
}

// ResolveRedis parses the active workspace and resolves the Redis resource by name.
func ResolveRedis(workspace *config.Workspace, name string) (*RedisResource, error) {
	if workspace == nil || workspace.Raw == nil {
		return nil, fmt.Errorf("%w: workspace configuration is empty", errors.ErrWorkspaceNotFound)
	}

	data, err := yaml.Marshal(workspace.Raw)
	if err != nil {
		return nil, fmt.Errorf("marshaling workspace configuration: %w", err)
	}

	var parsed struct {
		Resources struct {
			Redis []RedisResource `yaml:"redis"`
		} `yaml:"resources"`
	}

	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parsing workspace configuration: %w", err)
	}

	for _, redis := range parsed.Resources.Redis {
		if redis.Name == name {
			redis.BastionInstanceID = ResolveBastion(workspace, redis.BastionInstanceID)
			if redis.BastionInstanceID == "" {
				return nil, fmt.Errorf("bastion_instance_id not configured for Redis %q or active workspace", name)
			}
			return &redis, nil
		}
	}

	return nil, fmt.Errorf("%w: redis %q not found in workspaces configuration", errors.ErrWorkspaceNotFound, name)
}
