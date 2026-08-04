package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/ui/picker"
	"github.com/guppshub/cx-cli/internal/workspace"
	"github.com/spf13/cobra"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use [workspace]",
	Short: "Switch the active workspace context",
	Long:  `Switch the active workspace context in your configuration. If no workspace is specified, an interactive list will be presented.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cPath, err := config.Path()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve config path: %v\n", err)
			os.Exit(1)
		}

		store := config.New(cPath)
		mgr := workspace.New(store)

		var workspaceName string

		if len(args) > 0 {
			workspaceName = args[0]
		} else {
			// Load config to check available workspaces
			cfg, err := store.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
				os.Exit(1)
			}

			if len(cfg.Workspaces) == 0 {
				fmt.Println("No workspaces defined. Run 'cx init' to create a configuration template.")
				return
			}

			var names []string
			for name := range cfg.Workspaces {
				names = append(names, name)
			}
			sort.Strings(names)

			var rows []picker.Row
			for _, name := range names {
				activeIndicator := ""
				if name == cfg.Current {
					activeIndicator = "(active)"
				}
				rows = append(rows, picker.Row{
					ID: name,
					Fields: []string{
						name,
						activeIndicator,
					},
				})
			}

			selectedID, err := picker.SingleSelect("Select Active Workspace", []string{"Workspace", "Status"}, rows)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to select workspace: %v\n", err)
				os.Exit(1)
			}

			if selectedID == "" {
				printError("Selection cancelled")
				os.Exit(0)
			}
			workspaceName = selectedID
		}

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
