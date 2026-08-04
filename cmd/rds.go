package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/guppshub/cx-cli/internal/connection"
	awsprovider "github.com/guppshub/cx-cli/internal/provider/aws"
	"github.com/guppshub/cx-cli/internal/resource"
	"github.com/guppshub/cx-cli/internal/state"
	"github.com/guppshub/cx-cli/internal/tunnel"
	"github.com/guppshub/cx-cli/internal/ui/picker"
	"github.com/spf13/cobra"
)

var (
	rdsPortFlag       int
	rdsForegroundFlag bool
	rdsServerModeFlag bool
)

// rdsCmd represents the rds command
var rdsCmd = &cobra.Command{
	Use:   "rds [database]",
	Short: "Establish a secure tunnel to an RDS database resource",
	Long:  `Establish a secure SSH tunnel to an Amazon RDS database resource configured in your active workspace through a Bastion host.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		// 1. Initialize AWS provider and verify credentials
		awsProvider, ws, err := initAWSProvider(ctx, rdsServerModeFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var dbResource *resource.DatabaseResource

		if len(args) > 0 {
			// Resolve specified database
			dbName := args[0]
			dbResource, err = resource.ResolveRDS(ws, dbName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// List and select database resource interactively
			rdsList, err := resource.FetchRDS(ws)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(rdsList) == 0 {
				fmt.Println("No RDS database resources configured in the active workspace.")
				os.Exit(0)
			}

			if len(rdsList) == 1 {
				dbResource = &rdsList[0]
				fmt.Printf("Using RDS database resource: %s\n", dbResource.Name)
			} else {
				var rows []picker.Row
				for _, r := range rdsList {
					rows = append(rows, picker.Row{
						ID: r.Name,
						Fields: []string{
							r.Name,
							r.Engine,
							r.Endpoint,
							fmt.Sprint(r.Port),
						},
					})
				}
				headers := []string{"RDS Name", "Engine", "Endpoint", "Port"}
				selectedID, err := picker.SingleSelect("Select RDS Database", headers, rows)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if selectedID == "" {
					fmt.Println("Selection cancelled")
					os.Exit(0)
				}

				// Find the selected resource
				for _, r := range rdsList {
					if r.Name == selectedID {
						dbResource = &r
						break
					}
				}
			}
		}

		if dbResource == nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve RDS database resource\n")
			os.Exit(1)
		}

		sPath, err := state.Path()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve state path: %v\n", err)
			os.Exit(1)
		}
		stateStore := state.New(sPath)
		connMgr := connection.NewManager(stateStore)

		// Check if active connection already exists
		if !rdsServerModeFlag {
			conn, err := connMgr.GetActiveConnection(dbResource.Name)
			if err == nil && conn != nil {
				// If the connection is not in a healthy or recovering state, we clean it up and restart
				if conn.State == string(connection.StateStopped) || conn.State == string(connection.StateFailed) {
					fmt.Printf("Existing tunnel for %q is in %s state. Cleaning up and restarting...\n", dbResource.Name, conn.State)
					connection.TerminateProcessGroup(conn.Pid, 1000*time.Millisecond)
					_ = connMgr.DeregisterState(conn.ConnectionID)
				} else {
					stateStr := conn.State
					if stateStr == "" {
						stateStr = "Healthy"
					}
					fmt.Printf("Tunnel to RDS database %q is already running in background (PID: %d, State: %s).\n", conn.Name, conn.Pid, stateStr)
					fmt.Printf("RDS database %q is listening on local port %d.\n", conn.Name, conn.LocalPort)
					return
				}
			}
		}

		// Local port mapping
		localPort := rdsPortFlag
		if localPort <= 0 {
			localPort = dbResource.LocalPort
		}
		// Final fallback port
		if localPort <= 0 {
			localPort = 5432
		}

		target := &tunnel.Target{
			BastionInstanceID:  dbResource.BastionInstanceID,
			RemoteHost:         dbResource.Endpoint,
			RemotePort:         dbResource.Port,
			PreferredLocalPort: localPort,
		}

		// 3. Handshake connectivity check (only in foreground/parent mode!)
		if !rdsServerModeFlag {
			// Verify bastion and SSM connectivity with a quick handshake
			fmt.Printf("Verifying connection to bastion %s...\n", target.BastionInstanceID)
			if err := connMgr.PreflightHandshake(ctx, awsProvider, target, dbResource.Engine); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Connection handshake successful.")

			if !rdsForegroundFlag {
				// Launch detached background daemon
				logDir := filepath.Join(filepath.Dir(sPath), "logs")
				if err := os.MkdirAll(logDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to create log directory: %v\n", err)
					os.Exit(1)
				}

				daemon, err := connection.SpawnDaemon(os.Args[0], "rds", dbResource.Name, localPort, logDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error spawning background daemon: %v\n", err)
					os.Exit(1)
				}

				fmt.Printf("Starting background tunnel daemon for RDS database %s (port %d)...\n", dbResource.Name, localPort)
				finalLocalPort, err := daemon.VerifyRegistration(stateStore, 5*time.Second)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error: background daemon failed to initialize. Check logs:")
					logData, _ := os.ReadFile(daemon.LogPath())
					fmt.Fprintf(os.Stderr, "%s\n", string(logData))
					if daemon.ErrorLogPath() != daemon.LogPath() {
						errData, _ := os.ReadFile(daemon.ErrorLogPath())
						if len(errData) > 0 {
							fmt.Fprintf(os.Stderr, "Errors:\n%s\n", string(errData))
						}
					}
					os.Exit(1)
				}

				fmt.Printf("Success! Tunnel established in background.\n")
				fmt.Printf("RDS database %q is listening on local port %d.\n", dbResource.Name, finalLocalPort)
				fmt.Printf("Log file: %s\n", daemon.LogPath())
				return
			}
		}

		// 4. Server mode: use supervisor with auto-reconnection
		if rdsServerModeFlag {
			connection.IgnoreUserSignals()
			connID := fmt.Sprintf("cx-conn-%s-%d", dbResource.Name, target.PreferredLocalPort)
			logger := log.New(os.Stderr, "", log.LstdFlags)
			dialer := awsprovider.NewTunnelDialer(awsProvider, target)

			sv := connection.NewSupervisor(connection.SupervisorConfig{
				Name:   dbResource.Name,
				Type:   "rds",
				Dialer: dialer,
				Policy: connection.NewFixedBackoff(5*time.Second, 50),
				Logger: logger,
				OnStateChange: func(meta connection.Metadata) {
					profileStr, _ := ws.Raw["profile"].(string)
					regionStr, _ := ws.Raw["region"].(string)
					_ = connMgr.UpdateState(connID, &state.ConnectionMetadata{
						Type:         meta.Type,
						Name:         meta.Name,
						LocalPort:    meta.Port,
						ConnectionID: connID,
						ConnectedAt:  meta.StartedAt.Format(time.RFC3339),
						Pid:          os.Getpid(),
						State:        string(meta.State),
						Restarts:     meta.Restarts,
						LastFailure:  meta.LastFailure,
						LastRestart:  meta.LastRestart.Format(time.RFC3339),
						Profile:      profileStr,
						Region:       regionStr,
						SessionID:    meta.SessionID,
					})
				},
			})

			if err := sv.Start(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting supervisor: %v\n", err)
				os.Exit(1)
			}

			// Wait for stop signal or supervisor exit
			select {
			case <-ctx.Done():
				sv.Stop()
			case <-sv.Done():
			}
			_ = connMgr.DeregisterState(connID)
			return
		}

		// 5. Foreground mode: direct tunnel
		tunnelConn, err := awsProvider.DialTunnel(ctx, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting tunnel: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = tunnelConn.Close() }()

		fmt.Printf("Tunneling RDS database %s (%s) through local port %d...\n", dbResource.Name, dbResource.Engine, target.PreferredLocalPort)
		fmt.Println("Press Ctrl+C to terminate connection.")

		<-ctx.Done()
		fmt.Println("Terminating tunnel connection...")
	},
}

func init() {
	rdsCmd.Flags().IntVarP(&rdsPortFlag, "port", "p", 0, "Local port override")
	rdsCmd.Flags().BoolVarP(&rdsForegroundFlag, "foreground", "f", false, "Run tunnel in the foreground")
	rdsCmd.Flags().BoolVar(&rdsServerModeFlag, "server", false, "Internal use only: start background tunnel server")
	_ = rdsCmd.Flags().MarkHidden("server")
	rootCmd.AddCommand(rdsCmd)
}
