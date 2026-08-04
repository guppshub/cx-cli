package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
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
			fmt.Println("No active tunnels running.")
			return
		}

		fmt.Println("Active Tunnels:")
		fmt.Printf("%-15s %-10s %-12s %-12s %-8s %-10s %-20s\n", "NAME", "TYPE", "STATE", "LOCAL PORT", "PID", "RESTARTS", "CONNECTED AT")
		for _, conn := range active {
			connState := conn.State
			if connState == "" {
				connState = "Healthy"
			}

			// Expiration check
			if conn.Expiration != "" {
				if expTime, err := time.Parse(time.RFC3339, conn.Expiration); err == nil {
					if time.Now().After(expTime) {
						connState = "Expired"
					}
				}
			}

			var stateCol string
			switch connState {
			case "Healthy":
				stateCol = greenStyle.Width(12).Render("Healthy")
			case "Expired":
				stateCol = redStyle.Width(12).Render("Expired")
			case "Stopped", "Failed":
				stateCol = redStyle.Width(12).Render(connState)
			default:
				stateCol = lipgloss.NewStyle().Width(12).Render(connState)
			}

			nameCol := lipgloss.NewStyle().Width(15).Render(conn.Name)
			typeCol := lipgloss.NewStyle().Width(10).Render(conn.Type)
			portCol := lipgloss.NewStyle().Width(12).Render(fmt.Sprint(conn.LocalPort))
			pidCol := lipgloss.NewStyle().Width(8).Render(fmt.Sprint(conn.Pid))
			restartsCol := lipgloss.NewStyle().Width(10).Render(fmt.Sprint(conn.Restarts))
			timeCol := lipgloss.NewStyle().Width(20).Render(conn.ConnectedAt)

			fmt.Println(nameCol + " " + typeCol + " " + stateCol + " " + portCol + " " + pidCol + " " + restartsCol + " " + timeCol)
		}
	},
}

var (
	greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	redStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

func init() {
	rootCmd.AddCommand(statusCmd)
}
