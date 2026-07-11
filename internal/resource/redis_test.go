package resource

import (
	"testing"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/errors"
)

func TestResolveRedis(t *testing.T) {
	workspace := &config.Workspace{
		Provider: "aws",
		Raw: map[string]any{
			"profile": "staging-admin",
			"region":  "us-east-1",
			"resources": map[string]any{
				"redis": []any{
					map[string]any{
						"name":                "cache-main",
						"host":                "staging-redis.elasticache.amazonaws.com",
						"port":                6379,
						"local_port":          6379,
						"bastion_instance_id": "i-123456",
					},
				},
			},
		},
	}

	// 1. Resolve existing Redis
	redis, err := ResolveRedis(workspace, "cache-main")
	if err != nil {
		t.Fatalf("unexpected error resolving cache-main: %v", err)
	}

	if redis.Name != "cache-main" {
		t.Errorf("expected redis name to be 'cache-main', got %q", redis.Name)
	}
	if redis.BastionInstanceID != "i-123456" {
		t.Errorf("expected bastion id to be 'i-123456', got %q", redis.BastionInstanceID)
	}

	// 2. Resolve non-existent Redis
	_, err = ResolveRedis(workspace, "cache-secondary")
	if !errors.Is(err, errors.ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound for cache-secondary, got %v", err)
	}

	// 3. Resolve with empty resources list
	emptyWorkspace := &config.Workspace{
		Provider: "aws",
		Raw:      map[string]any{},
	}
	_, err = ResolveRedis(emptyWorkspace, "cache-main")
	if !errors.Is(err, errors.ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound for empty workspace, got %v", err)
	}

	// 4. Resolve workspace-level fallback success
	fallbackWorkspace := &config.Workspace{
		Provider: "aws",
		Raw: map[string]any{
			"profile":             "staging-admin",
			"region":              "us-east-1",
			"bastion_instance_id": "i-fallback",
			"resources": map[string]any{
				"redis": []any{
					map[string]any{
						"name": "cache-main",
						"host": "staging-redis.elasticache.amazonaws.com",
						"port": 6379,
					},
				},
			},
		},
	}
	redisFallback, err := ResolveRedis(fallbackWorkspace, "cache-main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if redisFallback.BastionInstanceID != "i-fallback" {
		t.Errorf("expected fallback bastion ID 'i-fallback', got %q", redisFallback.BastionInstanceID)
	}

	// 5. Resolve fails when bastion is missing everywhere
	missingBastionWorkspace := &config.Workspace{
		Provider: "aws",
		Raw: map[string]any{
			"profile": "staging-admin",
			"region":  "us-east-1",
			"resources": map[string]any{
				"redis": []any{
					map[string]any{
						"name": "cache-main",
						"host": "staging-redis.elasticache.amazonaws.com",
						"port": 6379,
					},
				},
			},
		},
	}
	_, err = ResolveRedis(missingBastionWorkspace, "cache-main")
	if err == nil {
		t.Error("expected error when bastion instance ID is missing at both levels, but it resolved successfully")
	}
}
