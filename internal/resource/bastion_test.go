package resource

import (
	"testing"

	"github.com/guppshub/cx-cli/internal/config"
)

func TestResolveBastion(t *testing.T) {
	wsWithBastion := &config.Workspace{
		Raw: map[string]any{
			"bastion_instance_id": "i-workspace-level",
		},
	}
	wsWithoutBastion := &config.Workspace{
		Raw: map[string]any{},
	}

	tests := []struct {
		name            string
		workspace       *config.Workspace
		resourceBastion string
		expected        string
	}{
		{
			name:            "prefer resource level",
			workspace:       wsWithBastion,
			resourceBastion: "i-resource-level",
			expected:        "i-resource-level",
		},
		{
			name:            "fallback to workspace level",
			workspace:       wsWithBastion,
			resourceBastion: "",
			expected:        "i-workspace-level",
		},
		{
			name:            "no bastion anywhere",
			workspace:       wsWithoutBastion,
			resourceBastion: "",
			expected:        "",
		},
		{
			name:            "nil workspace fallback",
			workspace:       nil,
			resourceBastion: "",
			expected:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveBastion(tt.workspace, tt.resourceBastion)
			if result != tt.expected {
				t.Errorf("ResolveBastion() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
