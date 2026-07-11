package resource

import (
	"fmt"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/errors"
	"gopkg.in/yaml.v3"
)

// DatabaseResource represents the configuration for a target database resource.
type DatabaseResource struct {
	Name              string `yaml:"name"`
	Engine            string `yaml:"engine"`
	Endpoint          string `yaml:"endpoint"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
}

// ResolveDatabase parses the active workspace and resolves the database resource by name.
// Returns ErrWorkspaceNotFound if the database is not found or is missing in config.
func ResolveDatabase(workspace *config.Workspace, name string) (*DatabaseResource, error) {
	if workspace == nil || workspace.Raw == nil {
		return nil, fmt.Errorf("%w: workspace configuration is empty", errors.ErrWorkspaceNotFound)
	}

	data, err := yaml.Marshal(workspace.Raw)
	if err != nil {
		return nil, fmt.Errorf("marshaling workspace configuration: %w", err)
	}

	var parsed struct {
		Resources struct {
			Databases []DatabaseResource `yaml:"databases"`
		} `yaml:"resources"`
	}

	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parsing workspace configuration: %w", err)
	}

	for _, db := range parsed.Resources.Databases {
		if db.Name == name {
			db.BastionInstanceID = ResolveBastion(workspace, db.BastionInstanceID)
			if db.BastionInstanceID == "" {
				return nil, fmt.Errorf("bastion_instance_id not configured for database %q or active workspace", name)
			}
			return &db, nil
		}
	}

	return nil, fmt.Errorf("%w: database %q not found in workspaces configuration", errors.ErrWorkspaceNotFound, name)
}
