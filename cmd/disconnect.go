package cmd

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/guppshub/cx-cli/internal/state"
	"github.com/spf13/cobra"
)

// disconnectCmd represents the disconnect command
var disconnectCmd = &cobra.Command{
	Use:   "disconnect [resource]",
	Short: "Disconnect an active background tunnel",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := args[0]

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

		var targetConn *state.ConnectionMetadata
		var targetConnID string
		for id, conn := range s.ActiveConnections {
			if conn.Name == resourceName {
				targetConn = conn
				targetConnID = id
				break
			}
		}

		if targetConn == nil {
			fmt.Fprintf(os.Stderr, "Error: no active background tunnel found for resource %q\n", resourceName)
			os.Exit(1)
		}

		fmt.Printf("Disconnecting resource %s (PID: %d)...\n", targetConn.Name, targetConn.Pid)

		proc, err := os.FindProcess(targetConn.Pid)
		if err == nil {
			// Send SIGINT so the process exits gracefully and runs its cleanup defers
			err = proc.Signal(syscall.SIGINT)
			if err != nil {
				// Fallback to Kill if SIGINT fails or on Windows
				_ = proc.Kill()
			}

			// Wait a brief moment for the process to terminate and clean up state
			time.Sleep(200 * time.Millisecond)
		}

		// Double check if the daemon cleaned up. If not (e.g. process crashed or killed forcibly), force clean state
		s, err = stateStore.Load()
		if err == nil {
			if _, exists := s.ActiveConnections[targetConnID]; exists {
				delete(s.ActiveConnections, targetConnID)
				_ = stateStore.Save(s)
			}
		}

		fmt.Printf("Success! Tunnel to resource %q disconnected.\n", resourceName)
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
