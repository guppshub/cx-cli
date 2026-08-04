package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/connection"
	awsprovider "github.com/guppshub/cx-cli/internal/provider/aws"
	"github.com/guppshub/cx-cli/internal/resource"
	"github.com/guppshub/cx-cli/internal/state"
	"github.com/guppshub/cx-cli/internal/tunnel"
	"github.com/guppshub/cx-cli/internal/ui/picker"
	"github.com/spf13/cobra"
)

var (
	osPortFlag       int
	osForegroundFlag bool
	osServerModeFlag bool
	osRefreshFlag    bool
)

// opensearchCmd represents the opensearch command
var opensearchCmd = &cobra.Command{
	Use:     "opensearch [domain]",
	Aliases: []string{"os"},
	Short:   "Establish a secure tunnel to an OpenSearch domain resource",
	Long:    `Establish a secure SSH tunnel to an Amazon OpenSearch domain resource configured in your active workspace through a Bastion host.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		// 1. Initialize AWS provider and verify credentials
		awsProvider, ws, err := initAWSProvider(ctx, osServerModeFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var osResource *resource.OpenSearchResource

		if len(args) > 0 {
			// Resolve specified domain
			osName := args[0]
			osResource, err = resource.ResolveOpenSearch(ws, osName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// List and select OpenSearch resource interactively
			osList, err := resource.FetchOpenSearch(ws)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(osList) == 0 {
				fmt.Println("No OpenSearch resources configured in the active workspace.")
				os.Exit(0)
			}

			if len(osList) == 1 {
				osResource = &osList[0]
				fmt.Printf("Using OpenSearch resource: %s\n", osResource.Name)
			} else {
				var rows []picker.Row
				for _, o := range osList {
					rows = append(rows, picker.Row{
						ID: o.Name,
						Fields: []string{
							o.Name,
							o.Endpoint,
							fmt.Sprint(o.Port),
							fmt.Sprint(o.LocalPort),
						},
					})
				}
				headers := []string{"OpenSearch Name", "Endpoint", "Port", "Local Port"}
				selectedID, err := picker.SingleSelect("Select OpenSearch Domain", headers, rows)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if selectedID == "" {
					fmt.Println("Selection cancelled")
					os.Exit(0)
				}

				// Find the selected resource
				for _, o := range osList {
					if o.Name == selectedID {
						osResource = &o
						break
					}
				}
			}
		}

		if osResource == nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve OpenSearch resource\n")
			os.Exit(1)
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
			os.Exit(1)
		}
		wsName := cfg.Current

		var tokenExpiration string

		// Handle MFA and STS temporary credentials if enabled
		if osResource.MFA {
			cacheKey := fmt.Sprintf("%s/%s", wsName, awsProvider.Profile())
			stsProfileName := awsProvider.Profile() + "-sts"
			if awsProvider.Profile() == "" {
				stsProfileName = "default-sts"
			}

			if osServerModeFlag {
				// Server mode: read from cache directly, no prompts
				stsCache, err := awsprovider.LoadSTSCache()
				if err == nil && stsCache != nil {
					if creds, ok := stsCache.Credentials[cacheKey]; ok {
						tokenExpiration = creds.Expiration.Format(time.RFC3339)
						// Ensure the background session uses the STS profile
						_ = awsprovider.UpdateAWSCredentialsFile(stsProfileName, &creds)
						awsProvider.SetProfile(stsProfileName)
					}
				}
			} else {
				// Client mode: load cached keys or prompt user for MFA code
				stsCache, err := awsprovider.LoadSTSCache()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error loading STS cache: %v\n", err)
					os.Exit(1)
				}

				creds, ok := stsCache.Credentials[cacheKey]

				if ok && !osRefreshFlag && time.Now().Add(5*time.Minute).Before(creds.Expiration) {
					// Use valid cached credentials
					_ = os.Setenv("AWS_ACCESS_KEY_ID", creds.AccessKeyID)
					_ = os.Setenv("AWS_SECRET_ACCESS_KEY", creds.SecretAccessKey)
					_ = os.Setenv("AWS_SESSION_TOKEN", creds.SessionToken)
					tokenExpiration = creds.Expiration.Format(time.RFC3339)

					// Write credentials to ~/.aws/credentials and swap active profile
					if err := awsprovider.UpdateAWSCredentialsFile(stsProfileName, &creds); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to update AWS credentials file: %v\n", err)
					}
					awsProvider.SetProfile(stsProfileName)
				} else {
					// Need to authenticate
					mfaSerial := osResource.MFASerial
					if mfaSerial == "" {
						fmt.Println("Auto-discovering AWS MFA device ARN...")
						serial, err := awsProvider.ListMFADevices(ctx)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error discovering MFA device: %v\n", err)
							os.Exit(1)
						}
						mfaSerial = serial
					}

					fmt.Printf("Enter AWS MFA Code (Device: %s): ", mfaSerial)
					mfaCode, err := readLineWithContext(ctx)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error reading MFA code: %v\n", err)
						os.Exit(1)
					}
					mfaCode = strings.TrimSpace(mfaCode)

					fmt.Println("Exchanging code for temporary AWS STS credentials...")
					stsCreds, err := awsProvider.GetSessionToken(ctx, mfaSerial, mfaCode, osResource.MFADuration)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: failed to get session token: %v\n", err)
						os.Exit(1)
					}

					// Cache new credentials
					cacheVal := awsprovider.STSCredentials{
						AccessKeyID:     stsCreds.AccessKeyID,
						SecretAccessKey: stsCreds.SecretAccessKey,
						SessionToken:    stsCreds.SessionToken,
						Expiration:      stsCreds.Expiration,
					}
					stsCache.Credentials[cacheKey] = cacheVal
					if err := awsprovider.SaveSTSCache(stsCache); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to save STS cache: %v\n", err)
					}

					// Inject into active environment variables
					_ = os.Setenv("AWS_ACCESS_KEY_ID", stsCreds.AccessKeyID)
					_ = os.Setenv("AWS_SECRET_ACCESS_KEY", stsCreds.SecretAccessKey)
					_ = os.Setenv("AWS_SESSION_TOKEN", stsCreds.SessionToken)
					tokenExpiration = stsCreds.Expiration.Format(time.RFC3339)

					// Write credentials to ~/.aws/credentials and swap active profile
					if err := awsprovider.UpdateAWSCredentialsFile(stsProfileName, &cacheVal); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to update AWS credentials file: %v\n", err)
					}
					awsProvider.SetProfile(stsProfileName)

					fmt.Println("✔ AWS STS credentials acquired, cached, and updated in ~/.aws/credentials successfully.")
				}
			}
		}

		sPath, err := state.Path()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve state path: %v\n", err)
			os.Exit(1)
		}
		stateStore := state.New(sPath)
		connMgr := connection.NewManager(stateStore)

		// Check if active connection already exists
		if !osServerModeFlag {
			conn, err := connMgr.GetActiveConnection(wsName, osResource.Name)
			if err == nil && conn != nil {
				// If the connection is not in a healthy or recovering state, we clean it up and restart
				if conn.State == string(connection.StateStopped) || conn.State == string(connection.StateFailed) {
					fmt.Printf("Existing tunnel for %q is in %s state. Cleaning up and restarting...\n", osResource.Name, conn.State)
					connection.TerminateProcessGroup(conn.Pid, 1000*time.Millisecond)
					_ = connMgr.DeregisterState(conn.ConnectionID)
				} else {
					stateStr := conn.State
					if stateStr == "" {
						stateStr = "Healthy"
					}
					fmt.Printf("Tunnel to OpenSearch %q is already running in background (PID: %d, State: %s).\n", conn.Name, conn.Pid, stateStr)
					fmt.Printf("OpenSearch %q is listening on local port %d.\n", conn.Name, conn.LocalPort)
					return
				}
			}
		}

		// Local port mapping
		localPort := osResource.LocalPort
		if osPortFlag > 0 {
			localPort = osPortFlag
		}

		target := &tunnel.Target{
			BastionInstanceID:  osResource.BastionInstanceID,
			RemoteHost:         osResource.Endpoint,
			RemotePort:         osResource.Port,
			PreferredLocalPort: localPort,
		}

		// 3. Handshake connectivity check (only in foreground/parent mode!)
		if !osServerModeFlag {
			// Verify bastion and SSM connectivity with a quick handshake
			fmt.Printf("Verifying connection to bastion %s...\n", target.BastionInstanceID)
			if err := connMgr.PreflightHandshake(ctx, awsProvider, target, "opensearch"); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Connection handshake successful.")

			if !osForegroundFlag {
				// Launch detached background daemon
				binPath, err := os.Executable()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to resolve binary path: %v\n", err)
					os.Exit(1)
				}

				logDir := filepath.Join(filepath.Dir(sPath), "logs")
				_ = os.MkdirAll(logDir, 0755)

				fmt.Println("Launching background OpenSearch tunnel daemon...")
				daemon, err := connection.SpawnDaemon(binPath, "opensearch", osResource.Name, localPort, logDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				// Wait up to 5 seconds to verify that the daemon registers itself successfully in state.json
				finalLocalPort, err := daemon.VerifyRegistration(stateStore, 5*time.Second)
				if err != nil {
					// Read stderr log output to display to the user
					logContent, _ := os.ReadFile(daemon.ErrorLogPath())
					fmt.Fprintf(os.Stderr, "Error: background daemon failed to start. Logs:\n%s\n", string(logContent))
					os.Exit(1)
				}

				fmt.Printf("✔ OpenSearch tunnel successfully started in background!\n")
				fmt.Printf("Local Bind: localhost:%d\n", finalLocalPort)
				fmt.Printf("View logs at: tail -f %s\n", daemon.LogPath())
				return
			}
		}

		// 3. Set up the dialer
		dialer := awsprovider.NewTunnelDialer(awsProvider, target)

		// 4. Server mode: use supervisor with auto-reconnection
		if osServerModeFlag {
			connection.IgnoreUserSignals()
			connID := fmt.Sprintf("cx-conn-%s-%d", osResource.Name, target.PreferredLocalPort)
			logger := log.New(os.Stderr, "", log.LstdFlags)

			sv := connection.NewSupervisor(connection.SupervisorConfig{
				Name:   osResource.Name,
				Type:   "opensearch",
				Dialer: dialer,
				Policy: connection.NewFixedBackoff(5*time.Second, 50),
				Logger: logger,
				OnStateChange: func(meta connection.Metadata) {
					profileStr, _ := ws.Raw["profile"].(string)
					regionStr, _ := ws.Raw["region"].(string)
					_ = connMgr.UpdateState(connID, &state.ConnectionMetadata{
						Type:         meta.Type,
						Name:         meta.Name,
						Workspace:    wsName,
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
						Expiration:   tokenExpiration,
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

		fmt.Printf("Tunneling OpenSearch %s through local port %d...\n", osResource.Name, target.PreferredLocalPort)
		fmt.Println("Press Ctrl+C to terminate connection.")

		<-ctx.Done()
		fmt.Println("Terminating tunnel connection...")
	},
}

func init() {
	opensearchCmd.Flags().IntVarP(&osPortFlag, "port", "p", 0, "Local port override")
	opensearchCmd.Flags().BoolVarP(&osForegroundFlag, "foreground", "f", false, "Run tunnel in the foreground")
	opensearchCmd.Flags().BoolVar(&osServerModeFlag, "server", false, "Internal use only: start background tunnel server")
	opensearchCmd.Flags().BoolVarP(&osRefreshFlag, "refresh", "r", false, "Force MFA token code prompt and refresh session token cache")
	_ = opensearchCmd.Flags().MarkHidden("server")
	rootCmd.AddCommand(opensearchCmd)
}
