package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ECSCluster represents an ECS cluster's basic metadata.
type ECSCluster struct {
	Name string
	ARN  string
}

// ECSService represents an ECS service's basic metadata.
type ECSService struct {
	Name string
	ARN  string
}

// ECSTask represents parsed ECS task status information.
type ECSTask struct {
	ID            string // Short 12-char Task ID (UUID)
	ARN           string // Full Task ARN
	LastStatus    string // RUNNING, PENDING, STOPPED, etc.
	HealthStatus  string // HEALTHY, UNHEALTHY, UNKNOWN
	CreatedAt     time.Time
	StartedAt     time.Time
	StoppedAt     time.Time
	StoppedReason string
	ExitCode      *int
	ContainerName string
}

// FetchECSClusters lists all cluster ARNs and parses their names.
func (p *Provider) FetchECSClusters(ctx context.Context) ([]ECSCluster, error) {
	if _, err := p.lookPathFunc("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found in PATH: %w", err)
	}

	args := []string{"ecs", "list-clusters", "--output", "json"}
	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("aws ecs list-clusters failed: %w (stderr: %q)", err, stderr.String())
	}

	var output struct {
		ClusterArns []string `json:"clusterArns"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("failed to parse AWS ECS clusters JSON: %w", err)
	}

	var clusters []ECSCluster
	for _, arn := range output.ClusterArns {
		parts := strings.Split(arn, "/")
		name := parts[len(parts)-1]
		clusters = append(clusters, ECSCluster{
			Name: name,
			ARN:  arn,
		})
	}
	return clusters, nil
}

// FetchECSServices lists all services inside the target cluster.
func (p *Provider) FetchECSServices(ctx context.Context, clusterName string) ([]ECSService, error) {
	if _, err := p.lookPathFunc("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found in PATH: %w", err)
	}

	args := []string{"ecs", "list-services", "--cluster", clusterName, "--output", "json"}
	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("aws ecs list-services failed: %w (stderr: %q)", err, stderr.String())
	}

	var output struct {
		ServiceArns []string `json:"serviceArns"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("failed to parse AWS ECS services JSON: %w", err)
	}

	var services []ECSService
	for _, arn := range output.ServiceArns {
		parts := strings.Split(arn, "/")
		name := parts[len(parts)-1]
		services = append(services, ECSService{
			Name: name,
			ARN:  arn,
		})
	}
	return services, nil
}

type rawDescribeTasksResponse struct {
	Tasks []struct {
		TaskArn       string    `json:"taskArn"`
		LastStatus    string    `json:"lastStatus"`
		HealthStatus  string    `json:"healthStatus"`
		CreatedAt     time.Time `json:"createdAt"`
		StartedAt     time.Time `json:"startedAt"`
		StoppedAt     time.Time `json:"stoppedAt"`
		StoppedReason string    `json:"stoppedReason"`
		Containers    []struct {
			Name     string `json:"name"`
			ExitCode *int   `json:"exitCode"`
			Reason   string `json:"reason"`
		} `json:"containers"`
	} `json:"tasks"`
}

// FetchECSTasks queries active and stopped tasks for the specified service and fetches full status details.
func (p *Provider) FetchECSTasks(ctx context.Context, clusterName, serviceName string) ([]ECSTask, error) {
	if _, err := p.lookPathFunc("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found in PATH: %w", err)
	}

	// 1. Query RUNNING tasks
	runningArns, err := p.listTasksByStatus(ctx, clusterName, serviceName, "RUNNING")
	if err != nil {
		return nil, err
	}

	// 2. Query STOPPED tasks (limited to 10)
	stoppedArns, err := p.listTasksByStatus(ctx, clusterName, serviceName, "STOPPED")
	if err != nil {
		return nil, err
	}
	if len(stoppedArns) > 10 {
		stoppedArns = stoppedArns[:10]
	}

	// Merge task ARNs
	taskArns := append(runningArns, stoppedArns...)
	if len(taskArns) == 0 {
		return nil, nil
	}

	// 3. Describe Tasks details
	args := []string{"ecs", "describe-tasks", "--cluster", clusterName, "--output", "json", "--tasks"}
	args = append(args, taskArns...)

	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("aws ecs describe-tasks failed: %w (stderr: %q)", err, stderr.String())
	}

	var raw rawDescribeTasksResponse
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse AWS ECS describe-tasks JSON: %w", err)
	}

	var tasks []ECSTask
	for _, t := range raw.Tasks {
		// Parse short UUID from task ARN (e.g. arn:aws:ecs:region:account:task/cluster/uuid)
		parts := strings.Split(t.TaskArn, "/")
		uuid := parts[len(parts)-1]
		shortID := uuid
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}

		containerName := "N/A"
		var exitCode *int
		if len(t.Containers) > 0 {
			containerName = t.Containers[0].Name
			exitCode = t.Containers[0].ExitCode
		}

		tasks = append(tasks, ECSTask{
			ID:            shortID,
			ARN:           t.TaskArn,
			LastStatus:    t.LastStatus,
			HealthStatus:  t.HealthStatus,
			CreatedAt:     t.CreatedAt,
			StartedAt:     t.StartedAt,
			StoppedAt:     t.StoppedAt,
			StoppedReason: t.StoppedReason,
			ExitCode:      exitCode,
			ContainerName: containerName,
		})
	}

	return tasks, nil
}

func (p *Provider) listTasksByStatus(ctx context.Context, clusterName, serviceName, status string) ([]string, error) {
	args := []string{"ecs", "list-tasks", "--cluster", clusterName, "--service-name", serviceName, "--desired-status", status, "--output", "json"}
	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If list-tasks fails, return clean err
		return nil, fmt.Errorf("aws ecs list-tasks (%s) failed: %w (stderr: %q)", status, err, stderr.String())
	}

	var output struct {
		TaskArns []string `json:"taskArns"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("failed to parse AWS ECS list-tasks JSON: %w", err)
	}

	return output.TaskArns, nil
}
