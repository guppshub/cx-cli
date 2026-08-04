package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/connection"
	awsprovider "github.com/guppshub/cx-cli/internal/provider/aws"
	"github.com/guppshub/cx-cli/internal/state"
	"github.com/guppshub/cx-cli/internal/ui/picker"
	"github.com/guppshub/cx-cli/internal/ui/spinner"
	"github.com/spf13/cobra"
)

// disconnectCmd represents the disconnect command
var disconnectCmd = &cobra.Command{
	Use:   "disconnect [resource]",
	Short: "Disconnect active background tunnels",
	Long:  `Disconnect active background tunnels. You can specify a resource name, a service type (rds, redis, opensearch), "all" to disconnect everything, or leave empty to pick interactively.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sPath, err := state.Path()
		if err != nil {
			printError("failed to resolve state path: %v", err)
			os.Exit(1)
		}
		stateStore := state.New(sPath)

		s, err := stateStore.Load()
		if err != nil {
			printError("failed to load state: %v", err)
			os.Exit(1)
		}

		cfg, err := config.Load()
		wsName := ""
		if err == nil && cfg != nil {
			wsName = cfg.Current
		}

		// Filter out stale connections and clean up state
		var active []*state.ConnectionMetadata
		var staleIDs []string
		for id, conn := range s.ActiveConnections {
			if connection.IsProcessAlive(conn.Pid) {
				active = append(active, conn)
			} else {
				staleIDs = append(staleIDs, id)
			}
		}

		if len(staleIDs) > 0 {
			for _, id := range staleIDs {
				delete(s.ActiveConnections, id)
			}
			_ = stateStore.Save(s)
		}

		if len(active) == 0 {
			printInfo("No active tunnels running.")
			return
		}

		var targetsToDisconnect []*state.ConnectionMetadata

		if len(args) == 0 {
			if len(active) == 1 {
				targetsToDisconnect = append(targetsToDisconnect, active[0])
			} else {
				// Show TUI picker
				var rows []picker.Row
				for _, conn := range active {
					rows = append(rows, picker.Row{
						ID: conn.ConnectionID,
						Fields: []string{
							conn.Name,
							conn.Type,
							conn.Workspace,
							fmt.Sprint(conn.LocalPort),
							conn.State,
						},
					})
				}
				headers := []string{"Name", "Type", "Workspace", "Port", "State"}
				selectedID, err := picker.SingleSelect("Select Tunnel to Disconnect", headers, rows)
				if err != nil {
					printError("failed to run picker: %v", err)
					os.Exit(1)
				}
				if selectedID == "" {
					printError("Selection cancelled")
					os.Exit(0)
				}

				// Find selected connection
				for _, conn := range active {
					if conn.ConnectionID == selectedID {
						targetsToDisconnect = append(targetsToDisconnect, conn)
						break
					}
				}
			}
		} else {
			target := args[0]
			switch target {
			case "all":
				targetsToDisconnect = active
			case "rds", "redis", "opensearch":
				for _, conn := range active {
					if conn.Type == target {
						targetsToDisconnect = append(targetsToDisconnect, conn)
					}
				}
				if len(targetsToDisconnect) == 0 {
					printError("no active tunnels of type %q found", target)
					os.Exit(1)
				}
			default:
				// Look up by name inside current workspace first
				for _, conn := range active {
					if conn.Name == target && conn.Workspace == wsName {
						targetsToDisconnect = append(targetsToDisconnect, conn)
						break
					}
				}
				// Fallback: lookup by name generally
				if len(targetsToDisconnect) == 0 {
					for _, conn := range active {
						if conn.Name == target {
							targetsToDisconnect = append(targetsToDisconnect, conn)
							break
						}
					}
				}

				if len(targetsToDisconnect) == 0 {
					printError("no active background tunnel found for resource %q", target)
					os.Exit(1)
				}
			}
		}

		var spin *spinner.Spinner
		if len(targetsToDisconnect) == 1 {
			spin = spinner.Start(fmt.Sprintf("Disconnecting tunnel to resource %s...", boldStyle.Render(targetsToDisconnect[0].Name)))
		} else {
			spin = spinner.Start(fmt.Sprintf("Disconnecting %d active tunnels...", len(targetsToDisconnect)))
		}

		for _, conn := range targetsToDisconnect {
			if conn.SessionID != "" {
				awsProvider := awsprovider.New(conn.Profile, conn.Region)
				awsProvider.TerminateSession(conn.SessionID)
			}
			connection.TerminateProcessGroup(conn.Pid, 1500*time.Millisecond)
		}

		// Reload state and delete connection nodes
		s, err = stateStore.Load()
		if err == nil {
			for _, conn := range targetsToDisconnect {
				delete(s.ActiveConnections, conn.ConnectionID)
			}
			_ = stateStore.Save(s)
		}

		spin.Stop()
		if len(targetsToDisconnect) == 1 {
			printSuccess("Tunnel connection to %s disconnected successfully.", boldStyle.Render(targetsToDisconnect[0].Name))
		} else {
			printSuccess("Successfully disconnected %d active tunnels.", len(targetsToDisconnect))
		}
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
