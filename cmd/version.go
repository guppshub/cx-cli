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
		fmt.Printf("cx version %s (commit: %s, built: %s)\n", Version, CommitSHA, BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
