package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/guppshub/cx-cli/internal/connection"
	awsprovider "github.com/guppshub/cx-cli/internal/provider/aws"
	"github.com/guppshub/cx-cli/internal/resource"
	"github.com/guppshub/cx-cli/internal/state"
	"github.com/guppshub/cx-cli/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	redisPortFlag       int
	redisForegroundFlag bool
	redisServerModeFlag bool
)

// redisCmd represents the redis command
var redisCmd = &cobra.Command{
	Use:   "redis [cache]",
	Short: "Establish a secure tunnel to a Redis resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		redisName := args[0]

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		// 1. Initialize AWS provider and verify credentials
		awsProvider, ws, err := initAWSProvider(ctx, redisServerModeFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// 2. Resolve Redis resource details
		redisResource, err := resource.ResolveRedis(ws, redisName)
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
		if !redisServerModeFlag {
			conn, err := connMgr.GetActiveConnection(redisResource.Name)
			if err == nil && conn != nil {
				// If the connection is not in a healthy or recovering state, we clean it up and restart
				if conn.State == string(connection.StateStopped) || conn.State == string(connection.StateFailed) {
					fmt.Printf("Existing tunnel for %q is in %s state. Cleaning up and restarting...\n", redisResource.Name, conn.State)
					connection.TerminateProcessGroup(conn.Pid, 1000*time.Millisecond)
					_ = connMgr.DeregisterState(conn.ConnectionID)
				} else {
					stateStr := conn.State
					if stateStr == "" {
						stateStr = "Healthy"
					}
					fmt.Printf("Tunnel to Redis %q is already running in background (PID: %d, State: %s).\n", conn.Name, conn.Pid, stateStr)
					fmt.Printf("Redis %q is listening on local port %d.\n", conn.Name, conn.LocalPort)
					return
				}
			}
		}

		// Local port mapping
		localPort := redisPortFlag
		if localPort <= 0 {
			localPort = redisResource.LocalPort
		}
		// Final fallback port
		if localPort <= 0 {
			localPort = 6379
		}

		target := &tunnel.Target{
			BastionInstanceID:  redisResource.BastionInstanceID,
			RemoteHost:         redisResource.Host,
			RemotePort:         redisResource.Port,
			PreferredLocalPort: localPort,
		}

		// 3. Handshake connectivity check (only in foreground/parent mode!)
		if !redisServerModeFlag {
			// Verify bastion and SSM connectivity with a quick handshake
			fmt.Printf("Verifying connection to bastion %s...\n", target.BastionInstanceID)
			if err := connMgr.PreflightHandshake(ctx, awsProvider, target, "redis"); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Connection handshake successful.")

			if !redisForegroundFlag {
				// Launch detached background daemon
				logDir := filepath.Join(filepath.Dir(sPath), "logs")
				if err := os.MkdirAll(logDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to create log directory: %v\n", err)
					os.Exit(1)
				}
				logPath := filepath.Join(logDir, redisResource.Name+".log")
				logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to open log file: %v\n", err)
					os.Exit(1)
				}
				defer func() { _ = logFile.Close() }()

				daemonArgs := []string{"redis", redisName, "--server", "--port", fmt.Sprint(localPort)}
				daemonCmd := exec.Command(os.Args[0], daemonArgs...)
				daemonCmd.Stdout = logFile
				daemonCmd.Stderr = logFile
				detachCmd(daemonCmd)

				if err := daemonCmd.Start(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to start background daemon: %v\n", err)
					os.Exit(1)
				}

				// Wait for daemon to register in state.json
				fmt.Printf("Starting background tunnel daemon for Redis %s (port %d)...\n", redisResource.Name, localPort)
				registered := false
				var finalLocalPort int

				// Poll state file every 100ms for up to 5 seconds
				for i := 0; i < 50; i++ {
					time.Sleep(100 * time.Millisecond)
					s, err := stateStore.Load()
					if err == nil {
						// Look for the active connection matching this Redis and PID
						for _, conn := range s.ActiveConnections {
							if conn.Name == redisResource.Name && conn.Pid == daemonCmd.Process.Pid {
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
				fmt.Printf("Redis %q is listening on local port %d.\n", redisResource.Name, finalLocalPort)
				fmt.Printf("Log file: %s\n", logPath)
				return
			}
		}

		// 4. Server mode: use supervisor with auto-reconnection
		if redisServerModeFlag {
			connID := fmt.Sprintf("cx-conn-%s-%d", redisResource.Name, target.PreferredLocalPort)
			logger := log.New(os.Stderr, "", log.LstdFlags)
			dialer := awsprovider.NewTunnelDialer(awsProvider, target)

			sv := connection.NewSupervisor(connection.SupervisorConfig{
				Name:   redisResource.Name,
				Type:   "redis",
				Dialer: dialer,
				Policy: connection.NewFixedBackoff(5*time.Second, 50),
				Logger: logger,
				OnStateChange: func(meta connection.Metadata) {
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
					})
				},
			})

			if err := sv.Start(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting supervisor: %v\n", err)
				os.Exit(1)
			}

			// Wait for stop signal, then gracefully shut down
			<-ctx.Done()
			sv.Stop()
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

		fmt.Printf("Tunneling Redis %s through local port %d...\n", redisResource.Name, target.PreferredLocalPort)
		fmt.Println("Press Ctrl+C to terminate connection.")

		<-ctx.Done()
		fmt.Println("Terminating tunnel connection...")
	},
}

func init() {
	redisCmd.Flags().IntVarP(&redisPortFlag, "port", "p", 0, "Local port override")
	redisCmd.Flags().BoolVarP(&redisForegroundFlag, "foreground", "f", false, "Run tunnel in the foreground")
	redisCmd.Flags().BoolVar(&redisServerModeFlag, "server", false, "Internal use only: start background tunnel server")
	_ = redisCmd.Flags().MarkHidden("server")
	rootCmd.AddCommand(redisCmd)
}
