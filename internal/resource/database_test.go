package resource

import (
	"testing"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/errors"
)

func TestResolveRDS(t *testing.T) {
	workspace := &config.Workspace{
		Provider: "aws",
		Raw: map[string]any{
			"profile": "staging-admin",
			"region":  "us-east-1",
			"resources": map[string]any{
				"rds": []any{
					map[string]any{
						"name":                "mercury",
						"engine":              "postgres",
						"endpoint":            "staging-db.rds.amazonaws.com",
						"port":                5432,
						"local_port":          5432,
						"bastion_instance_id": "i-123456",
					},
				},
			},
		},
	}

	// 1. Resolve existing database
	db, err := ResolveRDS(workspace, "mercury")
	if err != nil {
		t.Fatalf("unexpected error resolving mercury: %v", err)
	}

	if db.Name != "mercury" {
		t.Errorf("expected db name to be 'mercury', got %q", db.Name)
	}
	if db.BastionInstanceID != "i-123456" {
		t.Errorf("expected bastion id to be 'i-123456', got %q", db.BastionInstanceID)
	}

	// 2. Resolve non-existent database
	_, err = ResolveRDS(workspace, "venus")
	if !errors.Is(err, errors.ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound for venus, got %v", err)
	}

	// 3. Resolve with empty resources list
	emptyWorkspace := &config.Workspace{
		Provider: "aws",
		Raw:      map[string]any{},
	}
	_, err = ResolveRDS(emptyWorkspace, "mercury")
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
				"rds": []any{
					map[string]any{
						"name":     "mercury",
						"engine":   "postgres",
						"endpoint": "staging-db.rds.amazonaws.com",
						"port":     5432,
					},
				},
			},
		},
	}
	dbFallback, err := ResolveRDS(fallbackWorkspace, "mercury")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dbFallback.BastionInstanceID != "i-fallback" {
		t.Errorf("expected fallback bastion ID 'i-fallback', got %q", dbFallback.BastionInstanceID)
	}

	// 5. Resolve fails when bastion is missing everywhere
	missingBastionWorkspace := &config.Workspace{
		Provider: "aws",
		Raw: map[string]any{
			"profile": "staging-admin",
			"region":  "us-east-1",
			"resources": map[string]any{
				"rds": []any{
					map[string]any{
						"name":     "mercury",
						"engine":   "postgres",
						"endpoint": "staging-db.rds.amazonaws.com",
						"port":     5432,
					},
				},
			},
		},
	}
	_, err = ResolveRDS(missingBastionWorkspace, "mercury")
	if err == nil {
		t.Error("expected error when bastion instance ID is missing at both levels, but it resolved successfully")
	}
}

func TestFetchRDS(t *testing.T) {
	workspace := &config.Workspace{
		Provider: "aws",
		Raw: map[string]any{
			"profile": "staging-admin",
			"region":  "us-east-1",
			"resources": map[string]any{
				"rds": []any{
					map[string]any{
						"name":                "mercury",
						"engine":              "postgres",
						"endpoint":            "staging-db.rds.amazonaws.com",
						"port":                5432,
						"local_port":          5432,
						"bastion_instance_id": "i-123456",
					},
					map[string]any{
						"name":     "venus",
						"engine":   "mysql",
						"endpoint": "staging-db2.rds.amazonaws.com",
						"port":     3306,
					},
				},
			},
		},
	}

	dbs, err := FetchRDS(workspace)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dbs) != 2 {
		t.Fatalf("expected 2 RDS databases, got %d", len(dbs))
	}

	if dbs[0].Name != "mercury" || dbs[1].Name != "venus" {
		t.Errorf("unexpected database names order: %s, %s", dbs[0].Name, dbs[1].Name)
	}
}
