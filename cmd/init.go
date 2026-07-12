package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/spf13/cobra"
)

var initForceFlag bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize starter configuration",
	Long:  `Initialize the default configuration directory and write a starter config.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := config.Path()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve configuration path: %v\n", err)
			os.Exit(1)
		}

		if err := runInit(path, initForceFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runInit(path string, force bool) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		if !force {
			return fmt.Errorf("configuration file already exists at %s, use --force to overwrite", path)
		}
	}

	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create configuration directory: %w", err)
	}

	// Write starter template
	err := os.WriteFile(path, []byte(config.DefaultConfigTemplate), 0644)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	if force {
		fmt.Printf("Success! Overwrote starter configuration at %s\n", path)
	} else {
		fmt.Printf("Success! Created starter configuration at %s\n", path)
	}
	return nil
}

func init() {
	initCmd.Flags().BoolVarP(&initForceFlag, "force", "f", false, "Force overwrite existing configuration file")
	rootCmd.AddCommand(initCmd)
}
