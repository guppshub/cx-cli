package cmd

import (
	"context"
	"fmt"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/provider/aws"
)

// initAWSProvider resolves the active workspace configuration, initializes the AWS provider, and ensures credentials are valid.
func initAWSProvider(ctx context.Context, skipEnsure bool) (*aws.Provider, *config.Workspace, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.Current == "" {
		return nil, nil, fmt.Errorf("no active workspace selected. Use \"cx use <workspace>\" first")
	}

	ws, exists := cfg.Workspaces[cfg.Current]
	if !exists {
		return nil, nil, fmt.Errorf("active workspace %q not found in workspaces", cfg.Current)
	}

	// v0.1 targets AWS provider only
	if ws.Provider != "aws" {
		return nil, nil, fmt.Errorf("unsupported cloud provider %q. v0.1 only supports aws", ws.Provider)
	}

	// Parse profile and region
	profileStr, _ := ws.Raw["profile"].(string)
	regionStr, _ := ws.Raw["region"].(string)

	awsProvider := aws.New(profileStr, regionStr)

	if !skipEnsure {
		// Ensure credentials are authenticated (with MFA prompt support)
		if err := awsProvider.EnsureCredentials(ctx, func(prompt string, secret bool) (string, error) {
			fmt.Print(prompt)
			var input string
			_, err := fmt.Scanln(&input)
			return input, err
		}); err != nil {
			return nil, nil, fmt.Errorf("credentials negotiation failed: %w", err)
		}
	}

	return awsProvider, ws, nil
}
