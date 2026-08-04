package aws

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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
		// Escape double quotes inside the command string so it is valid inside the JSON array
		escapedCmd := strings.ReplaceAll(startupCmd, `"`, `\"`)
		paramJSON := fmt.Sprintf(`{"command":["%s"]}`, escapedCmd)

		args = append(args,
			"--document-name", "AWS-StartInteractiveCommand",
			"--parameters", paramJSON,
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
	cmd.Stderr = &stderrFilter{w: os.Stderr}

	// Ignore Ctrl+C in the host cx process so that it is handled solely by the SSM session
	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	return cmd.Run()
}

type stderrFilter struct {
	w io.Writer
}

func (f *stderrFilter) Write(p []byte) (n int, err error) {
	s := string(p)
	if strings.Contains(s, "Starting session with SessionId:") {
		// Filter out the specific line
		lines := strings.Split(s, "\n")
		var out []string
		for _, l := range lines {
			if !strings.Contains(l, "Starting session with SessionId:") {
				out = append(out, l)
			}
		}
		if len(out) > 0 {
			_, err = f.w.Write([]byte(strings.Join(out, "\n")))
		}
		return len(p), err
	}
	return f.w.Write(p)
}
