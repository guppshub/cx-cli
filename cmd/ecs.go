package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/provider/aws"
	"github.com/guppshub/cx-cli/internal/ui/picker"
	"github.com/spf13/cobra"
)

var (
	ecsWatchFlag     bool
	ecsCacheFlag     string
	ecsWorkspaceFlag string
	ecsRefreshFlag   bool
)

var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "Monitor the state of ECS service tasks",
	Long:  `Retrieve ECS clusters and services in the active workspace and monitor their task states in real-time.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		go func() {
			<-sigChan
			cancel()
		}()

		// 1. Resolve workspace name (flag override or active config)
		wsName := ecsWorkspaceFlag
		if wsName == "" {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
				os.Exit(1)
			}
			wsName = cfg.Current
		}

		if wsName == "" {
			fmt.Fprintf(os.Stderr, "Error: no active workspace selected. Use \"cx use <workspace>\" first\n")
			os.Exit(1)
		}

		// 2. Load workspace config
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		ws, exists := cfg.Workspaces[wsName]
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: workspace %q not found in workspaces\n", wsName)
			os.Exit(1)
		}

		// 3. Handle --cache true/false configuration command
		if ecsCacheFlag != "" {
			handleCacheConfiguration(ctx, wsName, ws, cfg)
			return
		}

		// 4. Initialize AWS provider for the target workspace
		awsProvider, ws, err := initAWSProviderForWorkspace(ctx, wsName, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		ecsCfg, err := config.GetECSConfig(ws)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// 5. Query execution using cache or live AWS
		var tasks []aws.ECSTask
		clusterName, serviceName, err := resolveClusterAndService(ctx, awsProvider, wsName, ws, ecsCfg, ecsRefreshFlag)
		if err != nil {
			if err.Error() == "selection cancelled" {
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// 6. Monitor tasks
		if ecsWatchFlag {
			runWatchMode(ctx, awsProvider, clusterName, serviceName)
		} else {
			fmt.Println("Fetching task states from AWS...")
			tasks, err = awsProvider.FetchECSTasks(ctx, clusterName, serviceName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to fetch tasks: %v\n", err)
				os.Exit(1)
			}
			if len(tasks) == 0 {
				fmt.Printf("No active or stopped tasks found for service %q.\n", serviceName)
				os.Exit(0)
			}
			renderTasksTable(tasks, serviceName)
		}
	},
}

func handleCacheConfiguration(ctx context.Context, wsName string, ws *config.Workspace, cfg *config.Config) {
	cacheEnabled := false
	switch ecsCacheFlag {
	case "true":
		cacheEnabled = true
	case "false":
		cacheEnabled = false
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid value for --cache: must be 'true' or 'false'\n")
		os.Exit(1)
	}

	ecsCfg, err := config.GetECSConfig(ws)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ecsCfg.Cache = cacheEnabled

	if cacheEnabled && ecsCfg.DefaultCluster == "" {
		profileStr, _ := ws.Raw["profile"].(string)
		regionStr, _ := ws.Raw["region"].(string)
		awsProvider := aws.New(profileStr, regionStr)

		fmt.Println("Fetching ECS clusters from AWS to select default...")
		if err := awsProvider.EnsureCredentials(ctx, func(prompt string, secret bool) (string, error) {
			fmt.Print(prompt)
			var input string
			_, err := fmt.Scanln(&input)
			return input, err
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: credentials negotiation failed: %v\n", err)
			os.Exit(1)
		}

		clusters, err := awsProvider.FetchECSClusters(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to fetch clusters: %v\n", err)
			os.Exit(1)
		}

		if len(clusters) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no ECS clusters found in workspace %q to set as default\n", wsName)
			os.Exit(1)
		}

		var defaultCluster string
		if len(clusters) == 1 {
			defaultCluster = clusters[0].Name
		} else {
			var rows []picker.Row
			for _, c := range clusters {
				rows = append(rows, picker.Row{
					ID:     c.Name,
					Fields: []string{c.Name, c.ARN},
				})
			}
			selectedID, err := picker.SingleSelect(fmt.Sprintf("Select default ECS Cluster for workspace %q", wsName), []string{"Cluster Name", "ARN"}, rows)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if selectedID == "" {
				fmt.Println("Canceled setting default cluster")
				os.Exit(0)
			}
			defaultCluster = selectedID
		}
		ecsCfg.DefaultCluster = defaultCluster
	}

	config.SetECSConfig(ws, ecsCfg)
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save config: %v\n", err)
		os.Exit(1)
	}

	if cacheEnabled {
		fmt.Printf("✔ ECS caching enabled with default cluster %q for workspace %q\n", ecsCfg.DefaultCluster, wsName)
	} else {
		fmt.Printf("✔ ECS caching disabled for workspace %q\n", wsName)
	}
}

func initAWSProviderForWorkspace(ctx context.Context, wsName string, skipEnsure bool) (*aws.Provider, *config.Workspace, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	ws, exists := cfg.Workspaces[wsName]
	if !exists {
		return nil, nil, fmt.Errorf("workspace %q not found in workspaces", wsName)
	}

	if ws.Provider != "aws" {
		return nil, nil, fmt.Errorf("unsupported cloud provider %q. v0.1 only supports aws", ws.Provider)
	}

	profileStr, _ := ws.Raw["profile"].(string)
	regionStr, _ := ws.Raw["region"].(string)

	awsProvider := aws.New(profileStr, regionStr)

	if !skipEnsure {
		if err := awsProvider.EnsureCredentials(ctx, func(prompt string, secret bool) (string, error) {
			fmt.Print(prompt)
			var input string
			_, err := fmt.Scanln(&input)
			return input, err
		}); err != nil {
			return nil, nil, fmt.Errorf("credentials negotiation failed: %w", err)
		}
	}

	return awsProvider, ws, nil
}

func updateCacheData(wsName, clusterName string, clusters []aws.ECSCluster, services []aws.ECSService) {
	cache, err := aws.LoadCache()
	if err != nil {
		return
	}
	wsCache, ok := cache.Workspaces[wsName]
	if !ok {
		wsCache = &aws.WorkspaceCache{
			Services: make(map[string][]aws.ECSService),
		}
		cache.Workspaces[wsName] = wsCache
	}
	if wsCache.Services == nil {
		wsCache.Services = make(map[string][]aws.ECSService)
	}

	if len(clusters) > 0 {
		wsCache.Clusters = clusters
	}
	if len(services) > 0 {
		wsCache.Services[clusterName] = services
	}
	wsCache.LastUpdated = time.Now()

	_ = aws.SaveCache(cache)
}

func runWatchMode(ctx context.Context, p *aws.Provider, cluster, service string) {
	// Use alternate screen buffer, hide cursor, and clear screen
	fmt.Print("\033[?1049h\033[?25l\033[H\033[2J")
	defer fmt.Print("\033[?1049l\033[?25h") // Restore screen and cursor on exit

	for {
		tasks, err := p.FetchECSTasks(ctx, cluster, service)

		// Reposition cursor to top-left and clear the screen before rendering
		fmt.Print("\033[H\033[J")

		if err != nil {
			if ctx.Err() != nil {
				return
			}
			fmt.Print("\033[J") // Clear screen below cursor
			fmt.Fprintf(os.Stderr, "Error: failed to fetch tasks: %v\n", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		if len(tasks) == 0 {
			fmt.Printf("No active or stopped tasks found for service %q.\n", service)
		} else {
			renderTasksTable(tasks, service)
		}

		fmt.Println("\nPress Ctrl+C to exit watch mode.")

		// Clear any remaining old characters/lines below the current output
		fmt.Print("\033[J")

		select {
		case <-ctx.Done():
			fmt.Println("\nWatch mode stopped.")
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func renderTasksTable(tasks []aws.ECSTask, serviceName string) {
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	boldStyle := lipgloss.NewStyle().Bold(true)

	var tableRows [][]string
	for _, t := range tasks {
		statusText := t.LastStatus
		switch t.LastStatus {
		case "RUNNING":
			statusText = greenStyle.Render("● RUNNING")
		case "PENDING":
			statusText = yellowStyle.Render("⠋ PENDING")
		case "STOPPED":
			statusText = redStyle.Render("✖ STOPPED")
		}

		var healthText string
		switch t.HealthStatus {
		case "HEALTHY":
			healthText = greenStyle.Render("HEALTHY")
		case "UNHEALTHY":
			healthText = redStyle.Render("UNHEALTHY")
		default:
			if t.HealthStatus == "UNKNOWN" || t.HealthStatus == "" {
				healthText = dimStyle.Render("-")
			} else {
				healthText = dimStyle.Render(t.HealthStatus)
			}
		}

		var timeCol string
		var detailCol string
		switch t.LastStatus {
		case "RUNNING":
			if !t.StartedAt.IsZero() {
				timeCol = formatDuration(time.Since(t.StartedAt))
			} else {
				timeCol = "N/A"
			}
			if t.ContainerName != "" && t.ContainerName != "N/A" {
				detailCol = dimStyle.Render("Container: " + t.ContainerName)
			} else {
				detailCol = ""
			}
		case "STOPPED":
			if !t.StoppedAt.IsZero() {
				timeCol = fmt.Sprintf("Stopped %s ago", formatDuration(time.Since(t.StoppedAt)))
			} else {
				timeCol = "N/A"
			}
			var parts []string
			if t.ExitCode != nil {
				parts = append(parts, fmt.Sprintf("ExitCode: %d", *t.ExitCode))
			}
			if t.StoppedReason != "" {
				parts = append(parts, t.StoppedReason)
			}
			detailCol = redStyle.Render(strings.Join(parts, " - "))
		default:
			timeCol = "N/A"
			detailCol = ""
		}

		tableRows = append(tableRows, []string{
			boldStyle.Render(t.ID),
			statusText,
			healthText,
			timeCol,
			detailCol,
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("244"))).
		Headers("Task ID", "Status", "Health", "Uptime/Age", "Details / Stopped Reason").
		Rows(tableRows...)

	fmt.Printf("\nTasks for service %q:\n", serviceName)
	fmt.Println(t.Render())
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func resolveClusterAndService(ctx context.Context, awsProvider *aws.Provider, wsName string, ws *config.Workspace, ecsCfg *config.ECSConfig, forceRefresh bool) (string, string, error) {
	var clusterName string
	var serviceName string

	useCache := ecsCfg.Cache && !forceRefresh

	if useCache && ecsCfg.DefaultCluster != "" {
		clusterName = ecsCfg.DefaultCluster
		// Try to load services from ecs_cache.json
		cache, err := aws.LoadCache()
		var cachedServices []aws.ECSService
		if err == nil && cache != nil {
			if wsCache, ok := cache.Workspaces[wsName]; ok {
				if services, ok := wsCache.Services[clusterName]; ok {
					cachedServices = services
				}
			}
		}

		if len(cachedServices) > 0 {
			// 1-step selection: display cached service list directly
			fmt.Printf("Using cached cluster %q for workspace %q.\n", clusterName, wsName)
			var rows []picker.Row
			for _, s := range cachedServices {
				rows = append(rows, picker.Row{
					ID:     s.Name,
					Fields: []string{s.Name, s.ARN},
				})
			}
			headers := []string{"Service Name", "ARN"}
			selectedID, err := picker.SingleSelect("Select ECS Service", headers, rows)
			if err != nil {
				return "", "", err
			}
			if selectedID == "" {
				return "", "", fmt.Errorf("selection cancelled")
			}
			serviceName = selectedID
		} else {
			// Cache is active but empty; fetch services from AWS, cache them, and proceed
			fmt.Printf("Cache empty. Fetching services for default cluster %q from AWS...\n", clusterName)
			services, err := awsProvider.FetchECSServices(ctx, clusterName)
			if err != nil {
				return "", "", err
			}
			if len(services) == 0 {
				return "", "", fmt.Errorf("no ECS services found in cluster %q", clusterName)
			}

			// Update cache file with fetched services
			updateCacheData(wsName, clusterName, nil, services)

			var rows []picker.Row
			for _, s := range services {
				rows = append(rows, picker.Row{
					ID:     s.Name,
					Fields: []string{s.Name, s.ARN},
				})
			}
			headers := []string{"Service Name", "ARN"}
			selectedID, err := picker.SingleSelect("Select ECS Service", headers, rows)
			if err != nil {
				return "", "", err
			}
			if selectedID == "" {
				return "", "", fmt.Errorf("selection cancelled")
			}
			serviceName = selectedID
		}
	} else {
		// standard query flow (no cache or explicit refresh)
		fmt.Println("Fetching ECS clusters from AWS...")
		clusters, err := awsProvider.FetchECSClusters(ctx)
		if err != nil {
			return "", "", err
		}
		if len(clusters) == 0 {
			return "", "", fmt.Errorf("no ECS clusters found in workspace")
		}

		if len(clusters) == 1 {
			clusterName = clusters[0].Name
			fmt.Printf("Using ECS Cluster: %s\n", clusterName)
		} else {
			var rows []picker.Row
			for _, c := range clusters {
				rows = append(rows, picker.Row{
					ID:     c.Name,
					Fields: []string{c.Name, c.ARN},
				})
			}
			headers := []string{"Cluster Name", "ARN"}
			selectedID, err := picker.SingleSelect("Select ECS Cluster", headers, rows)
			if err != nil {
				return "", "", err
			}
			if selectedID == "" {
				return "", "", fmt.Errorf("selection cancelled")
			}
			clusterName = selectedID
		}

		fmt.Println("Fetching ECS services from AWS...")
		services, err := awsProvider.FetchECSServices(ctx, clusterName)
		if err != nil {
			return "", "", err
		}
		if len(services) == 0 {
			return "", "", fmt.Errorf("no ECS services found in cluster %q", clusterName)
		}

		// If cache is enabled (but we bypassed it due to --refresh), update the cache file
		if ecsCfg.Cache {
			updateCacheData(wsName, clusterName, clusters, services)
			fmt.Println("✔ ECS cache updated successfully.")
		}

		if len(services) == 1 {
			serviceName = services[0].Name
			fmt.Printf("Using ECS Service: %s\n", serviceName)
		} else {
			var rows []picker.Row
			for _, s := range services {
				rows = append(rows, picker.Row{
					ID:     s.Name,
					Fields: []string{s.Name, s.ARN},
				})
			}
			headers := []string{"Service Name", "ARN"}
			selectedID, err := picker.SingleSelect("Select ECS Service", headers, rows)
			if err != nil {
				return "", "", err
			}
			if selectedID == "" {
				return "", "", fmt.Errorf("selection cancelled")
			}
			serviceName = selectedID
		}
	}

	return clusterName, serviceName, nil
}

var (
	ecsLogsSinceFlag  string
	ecsLogsFilterFlag string
)

var ecsLogsCmd = &cobra.Command{
	Use:     "logs [service]",
	Aliases: []string{"cloudwatch", "cw"},
	Short:   "Tail CloudWatch logs for an ECS service container in real-time",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		wsName := ecsWorkspaceFlag
		if wsName == "" {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
				os.Exit(1)
			}
			wsName = cfg.Current
		}

		if wsName == "" {
			fmt.Fprintf(os.Stderr, "Error: no active workspace selected. Use \"cx use <workspace>\" first\n")
			os.Exit(1)
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if _, exists := cfg.Workspaces[wsName]; !exists {
			fmt.Fprintf(os.Stderr, "Error: workspace %q not found in workspaces\n", wsName)
			os.Exit(1)
		}

		awsProvider, ws, err := initAWSProviderForWorkspace(ctx, wsName, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		ecsCfg, err := config.GetECSConfig(ws)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var clusterName string
		var serviceName string

		// If service is specified as argument, we still need to know its cluster.
		// If ecsCfg.DefaultCluster is defined, we use it. Otherwise, we must list/prompt or search.
		if len(args) > 0 {
			serviceName = args[0]
			if ecsCfg.DefaultCluster != "" {
				clusterName = ecsCfg.DefaultCluster
			} else {
				// Search for which cluster contains this service, or prompt
				clusters, err := awsProvider.FetchECSClusters(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if len(clusters) == 0 {
					fmt.Println("No ECS clusters found.")
					os.Exit(1)
				}
				if len(clusters) == 1 {
					clusterName = clusters[0].Name
				} else {
					var rows []picker.Row
					for _, c := range clusters {
						rows = append(rows, picker.Row{
							ID:     c.Name,
							Fields: []string{c.Name, c.ARN},
						})
					}
					headers := []string{"Cluster Name", "ARN"}
					selectedID, err := picker.SingleSelect("Select ECS Cluster containing service", headers, rows)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						os.Exit(1)
					}
					if selectedID == "" {
						fmt.Println("Selection cancelled")
						os.Exit(0)
					}
					clusterName = selectedID
				}
			}
		} else {
			// Resolve cluster and service interactively
			clusterName, serviceName, err = resolveClusterAndService(ctx, awsProvider, wsName, ws, ecsCfg, ecsRefreshFlag)
			if err != nil {
				if err.Error() == "selection cancelled" {
					os.Exit(0)
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		// Find log group and stream prefix using cache or fetching live
		var logGroup string
		var streamPrefix string

		cache, err := aws.LoadCache()
		var wsCache *aws.WorkspaceCache
		var serviceIndex = -1

		if err == nil && cache != nil && !ecsRefreshFlag {
			if wc, ok := cache.Workspaces[wsName]; ok {
				wsCache = wc
				if services, ok := wc.Services[clusterName]; ok {
					for i, s := range services {
						if s.Name == serviceName {
							logGroup = s.LogGroup
							streamPrefix = s.LogStreamPrefix
							serviceIndex = i
							break
						}
					}
				}
			}
		}

		if logGroup == "" {
			fmt.Printf("Discovering CloudWatch log configuration for service %q...\n", serviceName)
			lg, prefix, err := awsProvider.FetchECSLogConfig(ctx, clusterName, serviceName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			logGroup = lg
			streamPrefix = prefix
			fmt.Printf("✔ Discovered Log Group: %s\n", logGroup)

			// Update cache entry
			if cache != nil {
				if wsCache == nil {
					wsCache = &aws.WorkspaceCache{
						Services: make(map[string][]aws.ECSService),
					}
					cache.Workspaces[wsName] = wsCache
				}
				services := wsCache.Services[clusterName]
				if serviceIndex >= 0 && serviceIndex < len(services) {
					services[serviceIndex].LogGroup = logGroup
					services[serviceIndex].LogStreamPrefix = streamPrefix
				} else {
					// Service not in cache yet (uncommon if we loaded via picker), add it
					services = append(services, aws.ECSService{
						Name:            serviceName,
						ARN:             "", // optional / unknown here
						LogGroup:        logGroup,
						LogStreamPrefix: streamPrefix,
					})
				}
				wsCache.Services[clusterName] = services
				wsCache.LastUpdated = time.Now()
				if err := aws.SaveCache(cache); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save log config to cache: %v\n", err)
				} else {
					fmt.Println("✔ Saved log config to ecs_cache.json")
				}
			}
		} else {
			if streamPrefix != "" {
				fmt.Printf("Using cached Log Group: %s (stream prefix: %s)\n", logGroup, streamPrefix)
			} else {
				fmt.Printf("Using cached Log Group: %s\n", logGroup)
			}
		}

		// Execute tail command: aws logs tail <logGroup> --follow --since <since>
		argsLogs := []string{"logs", "tail", logGroup, "--follow"}
		if ecsLogsSinceFlag != "" {
			argsLogs = append(argsLogs, "--since", ecsLogsSinceFlag)
		}
		if ecsLogsFilterFlag != "" {
			argsLogs = append(argsLogs, "--filter-pattern", ecsLogsFilterFlag)
		}

		profileName := awsProvider.Profile()
		if profileName != "" {
			argsLogs = append(argsLogs, "--profile", profileName)
		}
		if region := awsProvider.Region(); region != "" {
			argsLogs = append(argsLogs, "--region", region)
		}

		fmt.Printf("Tailing logs for %s (Press Ctrl+C to stop)...\n\n", serviceName)
		runCmd := exec.CommandContext(ctx, "aws", argsLogs...)
		runCmd.Stdin = os.Stdin
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr

		if err := runCmd.Run(); err != nil {
			if ctx.Err() != nil {
				return
			}
			fmt.Fprintf(os.Stderr, "\nError: tail command exited: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	ecsCmd.Flags().BoolVarP(&ecsWatchFlag, "watch", "w", false, "Enable continuous real-time task status updates")
	ecsCmd.Flags().StringVar(&ecsCacheFlag, "cache", "", "Configure ECS caching for the workspace ('true' or 'false')")
	ecsCmd.Flags().StringVar(&ecsWorkspaceFlag, "ws", "", "AWS workspace override")
	ecsCmd.Flags().BoolVarP(&ecsRefreshFlag, "refresh", "r", false, "Force a refresh of cached ECS clusters and services")
	
	ecsLogsCmd.Flags().StringVarP(&ecsLogsSinceFlag, "since", "s", "10m", "Tailing history window (e.g. 10m, 1h, 1d)")
	ecsLogsCmd.Flags().StringVarP(&ecsLogsFilterFlag, "filter", "f", "", "Filter pattern string to search/match")
	ecsCmd.AddCommand(ecsLogsCmd)

	rootCmd.AddCommand(ecsCmd)
}
