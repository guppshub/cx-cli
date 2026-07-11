package cmd

import (
	"fmt"
	"os"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/workspace"
	"github.com/spf13/cobra"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use [workspace]",
	Short: "Switch the active workspace context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]

		cPath, err := config.Path()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve config path: %v\n", err)
			os.Exit(1)
		}

		store := config.New(cPath)
		mgr := workspace.New(store)

		err = mgr.Use(workspaceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Now using workspace %q\n", workspaceName)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
