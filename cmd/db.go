package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/guppshub/cx-cli/internal/connection"
	"github.com/guppshub/cx-cli/internal/resource"
	"github.com/guppshub/cx-cli/internal/state"
	"github.com/guppshub/cx-cli/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	portFlag       int
	foregroundFlag bool
	serverModeFlag bool
)

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db [database]",
	Short: "Establish a secure tunnel to a database resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbName := args[0]

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		// 1. Initialize AWS provider and verify credentials
		awsProvider, ws, err := initAWSProvider(ctx, serverModeFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// 2. Resolve database resource details
		dbResource, err := resource.ResolveDatabase(ws, dbName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
		if !serverModeFlag {
			conn, err := connMgr.GetActiveConnection(dbResource.Name)
			if err == nil && conn != nil {
				fmt.Printf("Tunnel to database %q is already running in background (PID: %d).\n", conn.Name, conn.Pid)
				fmt.Printf("Database %q is listening on local port %d.\n", conn.Name, conn.LocalPort)
				return
			}
		}

		// Local port mapping
		localPort := portFlag
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
		if !serverModeFlag {
			// Verify bastion and SSM connectivity with a quick handshake
			fmt.Printf("Verifying connection to bastion %s...\n", target.BastionInstanceID)
			if err := connMgr.PreflightHandshake(ctx, awsProvider, target); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Connection handshake successful.")

			if !foregroundFlag {
				// Launch detached background daemon
				logDir := filepath.Join(filepath.Dir(sPath), "logs")
				if err := os.MkdirAll(logDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to create log directory: %v\n", err)
					os.Exit(1)
				}
				logPath := filepath.Join(logDir, dbResource.Name+".log")
				logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to open log file: %v\n", err)
					os.Exit(1)
				}
				defer func() { _ = logFile.Close() }()

				daemonArgs := []string{"db", dbName, "--server", "--port", fmt.Sprint(localPort)}
				daemonCmd := exec.Command(os.Args[0], daemonArgs...)
				daemonCmd.Stdout = logFile
				daemonCmd.Stderr = logFile
				detachCmd(daemonCmd)

				if err := daemonCmd.Start(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to start background daemon: %v\n", err)
					os.Exit(1)
				}

				// Wait for daemon to register in state.json
				fmt.Printf("Starting background tunnel daemon for database %s (port %d)...\n", dbResource.Name, localPort)
				registered := false
				var finalLocalPort int

				// Poll state file every 100ms for up to 5 seconds
				for i := 0; i < 50; i++ {
					time.Sleep(100 * time.Millisecond)
					s, err := stateStore.Load()
					if err == nil {
						// Look for the active connection matching this DB and PID
						for _, conn := range s.ActiveConnections {
							if conn.Name == dbResource.Name && conn.Pid == daemonCmd.Process.Pid {
								registered = true
								finalLocalPort = conn.LocalPort
								break
							}
						}
					}
					if registered {
						break
					}
				}

				if !registered {
					fmt.Fprintln(os.Stderr, "Error: background daemon failed to initialize. Check logs:")
					logData, _ := os.ReadFile(logPath)
					fmt.Fprintf(os.Stderr, "%s\n", string(logData))
					os.Exit(1)
				}

				fmt.Printf("Success! Tunnel established in background.\n")
				fmt.Printf("Database %q is listening on local port %d.\n", dbResource.Name, finalLocalPort)
				fmt.Printf("Log file: %s\n", logPath)
				return
			}
		}

		// 4. Run the tunnel natively
		tunnelConn, err := awsProvider.DialTunnel(ctx, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting tunnel: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = tunnelConn.Close() }()

		// Register connection state if server daemon
		if serverModeFlag {
			connID := fmt.Sprintf("cx-conn-%s-%d", dbResource.Name, target.PreferredLocalPort)
			if err := connMgr.RegisterState("database", dbResource.Name, target.PreferredLocalPort, connID); err != nil {
				fmt.Fprintf(os.Stderr, "Error registering state: %v\n", err)
				cancel()
			}
			defer func() {
				_ = connMgr.DeregisterState(connID)
			}()
		} else {
			fmt.Printf("Tunneling database %s (%s) through local port %d...\n", dbResource.Name, dbResource.Engine, target.PreferredLocalPort)
			fmt.Println("Press Ctrl+C to terminate connection.")
		}

		<-ctx.Done()
		if !serverModeFlag {
			fmt.Println("Terminating tunnel connection...")
		}
	},
}

func init() {
	dbCmd.Flags().IntVarP(&portFlag, "port", "p", 0, "Local port override")
	dbCmd.Flags().BoolVarP(&foregroundFlag, "foreground", "f", false, "Run tunnel in the foreground")
	dbCmd.Flags().BoolVar(&serverModeFlag, "server", false, "Internal use only: start background tunnel server")
	_ = dbCmd.Flags().MarkHidden("server")
	rootCmd.AddCommand(dbCmd)
}
