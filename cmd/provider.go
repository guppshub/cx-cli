package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/provider/aws"
	"github.com/guppshub/cx-cli/internal/tunnel"
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
			return readLineWithContext(ctx)
		}); err != nil {
			return nil, nil, fmt.Errorf("credentials negotiation failed: %w", err)
		}
	}

	return awsProvider, ws, nil
}

// readLineWithContext reads from stdin inside a goroutine, returning early if the context is cancelled.
func readLineWithContext(ctx context.Context) (string, error) {
	type result struct {
		text string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		var input string
		_, err := fmt.Scanln(&input)
		ch <- result{text: input, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-ch:
		return res.text, res.err
	}
}

var (
	// Styling helper variables
	boldStyle    = lipgloss.NewStyle().Bold(true)
	blueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	greenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	redStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

// printSuccess formats and prints a success checkmark line.
func printSuccess(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", greenStyle.Render("✔"), fmt.Sprintf(format, a...))
}

// printInfo formats and prints an informational bullet line.
func printInfo(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", blueStyle.Render("ℹ"), fmt.Sprintf(format, a...))
}

// printWarning formats and prints a warning bullet line.
func printWarning(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", yellowStyle.Render("⚠"), fmt.Sprintf(format, a...))
}

// printError formats and prints an error cross line.
func printError(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", redStyle.Render("✖"), fmt.Sprintf(format, a...))
}

// printDetailedError prints a structured, high-fidelity error representation to Stderr.
func printDetailedError(resType, resName string, target *tunnel.Target, err error) {
	fmt.Fprintln(os.Stderr)
	printError("Connection failed!")
	fmt.Fprintf(os.Stderr, "  %-18s %s (%s)\n", boldStyle.Render("Resource:"), resName, resType)
	if target != nil {
		fmt.Fprintf(os.Stderr, "  %-18s %s\n", boldStyle.Render("Bastion Host:"), target.BastionInstanceID)
		fmt.Fprintf(os.Stderr, "  %-18s %s:%d\n", boldStyle.Render("Target Endpoint:"), target.RemoteHost, target.RemotePort)
		fmt.Fprintf(os.Stderr, "  %-18s %d\n", boldStyle.Render("Local Port:"), target.PreferredLocalPort)
	}
	fmt.Fprintf(os.Stderr, "  %-18s %v\n", boldStyle.Render("Error Details:"), err)
	fmt.Fprintln(os.Stderr)
}
