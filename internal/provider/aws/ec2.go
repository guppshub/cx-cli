package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// EC2Instance represents an EC2 virtual machine's basic information.
type EC2Instance struct {
	Name             string `json:"Name"`
	InstanceID       string `json:"InstanceId"`
	State            string `json:"State"`
	PrivateIPAddress string `json:"PrivateIpAddress"`
}

// FetchEC2Instances runs the aws ec2 describe-instances CLI command and parses the output.
func (p *Provider) FetchEC2Instances(ctx context.Context) ([]EC2Instance, error) {
	// Verify dependencies
	if _, err := p.lookPathFunc("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found in PATH: %w", err)
	}

	args := []string{
		"ec2",
		"describe-instances",
		"--query", "Reservations[*].Instances[*].{Name:Tags[?Key=='Name'].Value|[0],InstanceId:InstanceId,State:State.Name,PrivateIpAddress:PrivateIpAddress}",
		"--output", "json",
	}

	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("aws ec2 describe-instances failed: %w (stderr: %q)", err, stderr.String())
	}

	var raw [][]EC2Instance
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse AWS EC2 JSON output: %w", err)
	}

	// Flatten results
	var instances []EC2Instance
	for _, reservation := range raw {
		for _, inst := range reservation {
			if inst.InstanceID == "" {
				continue
			}
			if inst.Name == "" {
				inst.Name = "Unnamed"
			}
			if inst.PrivateIPAddress == "" {
				inst.PrivateIPAddress = "N/A"
			}
			instances = append(instances, inst)
		}
	}

	return instances, nil
}
