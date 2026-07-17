package aws

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
)

// ConnectSSM establishes an interactive terminal SSM session to the target instance, optionally running a startup command.
func (p *Provider) ConnectSSM(instanceID string, startupCmd string) error {
	// Verify dependencies
	if _, err := p.lookPathFunc("aws"); err != nil {
		return fmt.Errorf("aws CLI not found in PATH: %w", err)
	}

	args := []string{
		"ssm",
		"start-session",
		"--target", instanceID,
	}

	if startupCmd != "" {
		args = append(args,
			"--document-name", "AWS-StartInteractiveCommand",
			"--parameters", fmt.Sprintf("command=%s", startupCmd),
		)
	}

	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.Command("aws", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Ignore Ctrl+C in the host cx process so that it is handled solely by the SSM session
	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	return cmd.Run()
}
