package aws

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"testing"
)

func TestECS_LookPathChecks(t *testing.T) {
	p := New("test-profile", "us-east-1")
	p.lookPathFunc = func(file string) (string, error) {
		return "", exec.ErrNotFound
	}

	ctx := context.Background()

	t.Run("FetchECSClusters", func(t *testing.T) {
		_, err := p.FetchECSClusters(ctx)
		if !errors.Is(err, exec.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("FetchECSServices", func(t *testing.T) {
		_, err := p.FetchECSServices(ctx, "cluster")
		if !errors.Is(err, exec.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("FetchECSTasks", func(t *testing.T) {
		_, err := p.FetchECSTasks(ctx, "cluster", "service")
		if !errors.Is(err, exec.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestECS_DescribeTasksJSONParsing(t *testing.T) {
	// Sample JSON output from "aws ecs describe-tasks"
	sampleJSON := `{
		"tasks": [
			{
				"taskArn": "arn:aws:ecs:us-east-1:123456789012:task/dev-cluster/3df67a90-e51c-43f1-90a1-87ab24f6ef1c",
				"lastStatus": "RUNNING",
				"healthStatus": "HEALTHY",
				"createdAt": "2026-08-04T10:00:00Z",
				"startedAt": "2026-08-04T10:01:00Z",
				"stoppedAt": "0001-01-01T00:00:00Z",
				"stoppedReason": "",
				"containers": [
					{
						"name": "web-app",
						"exitCode": null,
						"reason": ""
					}
				]
			},
			{
				"taskArn": "arn:aws:ecs:us-east-1:123456789012:task/dev-cluster/fa67e912-421c-43f1-90a1-87ab24f6ef2b",
				"lastStatus": "STOPPED",
				"healthStatus": "UNKNOWN",
				"createdAt": "2026-08-04T09:00:00Z",
				"startedAt": "2026-08-04T09:01:00Z",
				"stoppedAt": "2026-08-04T09:12:00Z",
				"stoppedReason": "Essential container in task exited",
				"containers": [
					{
						"name": "web-app",
						"exitCode": 1,
						"reason": "OutOfMemoryError"
					}
				]
			}
		]
	}`

	var raw rawDescribeTasksResponse
	if err := json.Unmarshal([]byte(sampleJSON), &raw); err != nil {
		t.Fatalf("failed to parse sample describe-tasks JSON: %v", err)
	}

	if len(raw.Tasks) != 2 {
		t.Fatalf("expected 2 tasks parsed, got %d", len(raw.Tasks))
	}

	// Task 1 Checks
	t1 := raw.Tasks[0]
	if t1.LastStatus != "RUNNING" {
		t.Errorf("expected task 1 status RUNNING, got %s", t1.LastStatus)
	}
	if t1.HealthStatus != "HEALTHY" {
		t.Errorf("expected task 1 health HEALTHY, got %s", t1.HealthStatus)
	}
	if len(t1.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(t1.Containers))
	}
	if t1.Containers[0].Name != "web-app" {
		t.Errorf("expected container name 'web-app', got %q", t1.Containers[0].Name)
	}
	if t1.Containers[0].ExitCode != nil {
		t.Errorf("expected container exitCode nil, got %v", *t1.Containers[0].ExitCode)
	}

	// Task 2 Checks
	t2 := raw.Tasks[1]
	if t2.LastStatus != "STOPPED" {
		t.Errorf("expected task 2 status STOPPED, got %s", t2.LastStatus)
	}
	if t2.StoppedReason != "Essential container in task exited" {
		t.Errorf("expected stopped reason, got %q", t2.StoppedReason)
	}
	if len(t2.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(t2.Containers))
	}
	if t2.Containers[0].ExitCode == nil || *t2.Containers[0].ExitCode != 1 {
		t.Errorf("expected task 2 container exitCode 1, got %v", t2.Containers[0].ExitCode)
	}
}
