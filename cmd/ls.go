package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/guppshub/cx-cli/internal/config"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all workspaces defined in configuration",
	Long:  `Retrieve and list all workspaces defined in your cx configuration, highlighting the active context.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
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

		activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

		fmt.Println("Workspaces:")
		for _, name := range names {
			if name == cfg.Current {
				fmt.Printf("* %s %s\n", activeStyle.Render(name), dimStyle.Render("(active)"))
			} else {
				fmt.Printf("  %s\n", name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
