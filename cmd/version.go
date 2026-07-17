package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the compiled release tag (injected at build time)
	Version = "v0.0.0-dev"
	// CommitSHA is the compiled git commit SHA (injected at build time)
	CommitSHA = "unknown"
	// BuildTime is the compiled RFC3339 build timestamp (injected at build time)
	BuildTime = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version metadata for the cx CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(FormatVersionString())
	},
}

// FormatVersionString returns a formatted version string, dropping metadata if it is unknown.
func FormatVersionString() string {
	if CommitSHA == "unknown" && BuildTime == "unknown" {
		return fmt.Sprintf("cx %s", Version)
	}
	return fmt.Sprintf("cx %s (commit: %s, built: %s)", Version, CommitSHA, BuildTime)
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
