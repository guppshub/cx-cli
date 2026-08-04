package resource

import (
	"fmt"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/errors"
	"gopkg.in/yaml.v3"
)

// DatabaseResource represents the configuration for a target database resource under "rds".
type DatabaseResource struct {
	Name              string `yaml:"name"`
	Engine            string `yaml:"engine"`
	Endpoint          string `yaml:"endpoint"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
	MFA               bool   `yaml:"mfa"`
	MFASerial         string `yaml:"mfa_serial"`
	MFADuration       int    `yaml:"mfa_duration"`
}

// ResolveRDS parses the active workspace and resolves the RDS database resource by name.
func ResolveRDS(workspace *config.Workspace, name string) (*DatabaseResource, error) {
	if workspace == nil || workspace.Raw == nil {
		return nil, fmt.Errorf("%w: workspace configuration is empty", errors.ErrWorkspaceNotFound)
	}

	data, err := yaml.Marshal(workspace.Raw)
	if err != nil {
		return nil, fmt.Errorf("marshaling workspace configuration: %w", err)
	}

	var parsed struct {
		Resources struct {
			RDS []DatabaseResource `yaml:"rds"`
		} `yaml:"resources"`
	}

	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parsing workspace configuration: %w", err)
	}

	// Parse workspace-level fallback MFA parameters
	var wsMFA struct {
		MFASerial   string `yaml:"mfa_serial"`
		MFADuration int    `yaml:"mfa_duration"`
	}
	_ = yaml.Unmarshal(data, &wsMFA)

	for _, db := range parsed.Resources.RDS {
		if db.Name == name {
			db.BastionInstanceID = ResolveBastion(workspace, db.BastionInstanceID)
			if db.BastionInstanceID == "" {
				return nil, fmt.Errorf("bastion_instance_id not configured for RDS database %q or active workspace", name)
			}
			// Apply workspace fallbacks if resource-level values are blank
			if db.MFASerial == "" {
				db.MFASerial = wsMFA.MFASerial
			}
			if db.MFADuration <= 0 {
				db.MFADuration = wsMFA.MFADuration
			}
			return &db, nil
		}
	}

	return nil, fmt.Errorf("%w: RDS database %q not found in workspaces configuration", errors.ErrWorkspaceNotFound, name)
}

// FetchRDS returns all RDS database resources configured in the workspace.
func FetchRDS(workspace *config.Workspace) ([]DatabaseResource, error) {
	if workspace == nil || workspace.Raw == nil {
		return nil, nil
	}

	data, err := yaml.Marshal(workspace.Raw)
	if err != nil {
		return nil, fmt.Errorf("marshaling workspace configuration: %w", err)
	}

	var parsed struct {
		Resources struct {
			RDS []DatabaseResource `yaml:"rds"`
		} `yaml:"resources"`
	}

	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parsing workspace configuration: %w", err)
	}

	// Parse workspace-level fallback MFA parameters
	var wsMFA struct {
		MFASerial   string `yaml:"mfa_serial"`
		MFADuration int    `yaml:"mfa_duration"`
	}
	_ = yaml.Unmarshal(data, &wsMFA)

	for i := range parsed.Resources.RDS {
		parsed.Resources.RDS[i].BastionInstanceID = ResolveBastion(workspace, parsed.Resources.RDS[i].BastionInstanceID)
		if parsed.Resources.RDS[i].MFASerial == "" {
			parsed.Resources.RDS[i].MFASerial = wsMFA.MFASerial
		}
		if parsed.Resources.RDS[i].MFADuration <= 0 {
			parsed.Resources.RDS[i].MFADuration = wsMFA.MFADuration
		}
	}

	return parsed.Resources.RDS, nil
}
