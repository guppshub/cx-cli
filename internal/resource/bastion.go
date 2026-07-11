package resource

import "github.com/guppshub/cx-cli/internal/config"

// ResolveBastion returns the active bastion instance ID, checking the resource-level
// bastion ID first, and falling back to the workspace-level configuration.
func ResolveBastion(workspace *config.Workspace, resourceBastion string) string {
	if resourceBastion != "" {
		return resourceBastion
	}
	if workspace != nil && workspace.Raw != nil {
		if wsBastion, ok := workspace.Raw["bastion_instance_id"].(string); ok {
			return wsBastion
		}
	}
	return ""
}
