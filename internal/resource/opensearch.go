package resource

import (
	"fmt"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/errors"
	"gopkg.in/yaml.v3"
)

// OpenSearchResource represents the configuration for a target OpenSearch domain under "opensearch".
type OpenSearchResource struct {
	Name              string `yaml:"name"`
	Endpoint          string `yaml:"endpoint"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
	MFA               bool   `yaml:"mfa"`
	MFASerial         string `yaml:"mfa_serial"`
	MFADuration       int    `yaml:"mfa_duration"`
}

// ResolveOpenSearch parses the active workspace and resolves the OpenSearch domain resource by name.
func ResolveOpenSearch(workspace *config.Workspace, name string) (*OpenSearchResource, error) {
	if workspace == nil || workspace.Raw == nil {
		return nil, fmt.Errorf("%w: workspace configuration is empty", errors.ErrWorkspaceNotFound)
	}

	data, err := yaml.Marshal(workspace.Raw)
	if err != nil {
		return nil, fmt.Errorf("marshaling workspace configuration: %w", err)
	}

	var parsed struct {
		Resources struct {
			OpenSearch []OpenSearchResource `yaml:"opensearch"`
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

	for i := range parsed.Resources.OpenSearch {
		if parsed.Resources.OpenSearch[i].Name == name {
			parsed.Resources.OpenSearch[i].BastionInstanceID = ResolveBastion(workspace, parsed.Resources.OpenSearch[i].BastionInstanceID)
			if parsed.Resources.OpenSearch[i].BastionInstanceID == "" {
				return nil, fmt.Errorf("bastion_instance_id not configured for OpenSearch resource %q or active workspace", name)
			}
			// Defaults
			if parsed.Resources.OpenSearch[i].Port <= 0 {
				parsed.Resources.OpenSearch[i].Port = 443 // Default OpenSearch HTTPS port
			}
			if parsed.Resources.OpenSearch[i].LocalPort <= 0 {
				parsed.Resources.OpenSearch[i].LocalPort = 9200 // Default local query port
			}
			// Apply workspace fallbacks if resource-level values are blank
			if parsed.Resources.OpenSearch[i].MFASerial == "" {
				parsed.Resources.OpenSearch[i].MFASerial = wsMFA.MFASerial
			}
			if parsed.Resources.OpenSearch[i].MFADuration <= 0 {
				parsed.Resources.OpenSearch[i].MFADuration = wsMFA.MFADuration
			}
			return &parsed.Resources.OpenSearch[i], nil
		}
	}

	return nil, fmt.Errorf("%w: OpenSearch resource %q not found in workspaces configuration", errors.ErrWorkspaceNotFound, name)
}

// FetchOpenSearch returns all OpenSearch resources configured in the workspace.
func FetchOpenSearch(workspace *config.Workspace) ([]OpenSearchResource, error) {
	if workspace == nil || workspace.Raw == nil {
		return nil, nil
	}

	data, err := yaml.Marshal(workspace.Raw)
	if err != nil {
		return nil, fmt.Errorf("marshaling workspace configuration: %w", err)
	}

	var parsed struct {
		Resources struct {
			OpenSearch []OpenSearchResource `yaml:"opensearch"`
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

	for i := range parsed.Resources.OpenSearch {
		parsed.Resources.OpenSearch[i].BastionInstanceID = ResolveBastion(workspace, parsed.Resources.OpenSearch[i].BastionInstanceID)
		if parsed.Resources.OpenSearch[i].Port <= 0 {
			parsed.Resources.OpenSearch[i].Port = 443
		}
		if parsed.Resources.OpenSearch[i].LocalPort <= 0 {
			parsed.Resources.OpenSearch[i].LocalPort = 9200
		}
		if parsed.Resources.OpenSearch[i].MFASerial == "" {
			parsed.Resources.OpenSearch[i].MFASerial = wsMFA.MFASerial
		}
		if parsed.Resources.OpenSearch[i].MFADuration <= 0 {
			parsed.Resources.OpenSearch[i].MFADuration = wsMFA.MFADuration
		}
	}

	return parsed.Resources.OpenSearch, nil
}
