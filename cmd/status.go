package cmd

import (
	"fmt"
	"os"

	"github.com/guppshub/cx-cli/internal/connection"
	"github.com/guppshub/cx-cli/internal/state"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of active database connections and tunnels",
	Run: func(cmd *cobra.Command, args []string) {
		sPath, err := state.Path()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve state path: %v\n", err)
			os.Exit(1)
		}
		stateStore := state.New(sPath)

		s, err := stateStore.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load state: %v\n", err)
			os.Exit(1)
		}

		// Filter out stale connections where the background process is no longer running
		var active []*state.ConnectionMetadata
		var staleIDs []string
		for id, conn := range s.ActiveConnections {
			if connection.IsProcessAlive(conn.Pid) {
				active = append(active, conn)
			} else {
				staleIDs = append(staleIDs, id)
			}
		}

		// If there are stale connections, clean them up from the state file
		if len(staleIDs) > 0 {
			for _, id := range staleIDs {
				delete(s.ActiveConnections, id)
			}
			if err := stateStore.Save(s); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save cleaned state: %v\n", err)
			}
		}

		if len(active) == 0 {
			fmt.Println("No active database tunnels running.")
			return
		}

		fmt.Println("Active Database Tunnels:")
		fmt.Printf("%-15s %-10s %-12s %-8s %-20s\n", "DATABASE", "TYPE", "LOCAL PORT", "PID", "CONNECTED AT")
		for _, conn := range active {
			fmt.Printf("%-15s %-10s %-12d %-8d %-20s\n",
				conn.Name,
				conn.Type,
				conn.LocalPort,
				conn.Pid,
				conn.ConnectedAt,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
