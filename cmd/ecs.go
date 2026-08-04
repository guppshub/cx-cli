package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/guppshub/cx-cli/internal/provider/aws"
	"github.com/guppshub/cx-cli/internal/ui/picker"
	"github.com/spf13/cobra"
)

var (
	ecsWatchFlag bool
)

var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "Monitor the state of ECS service tasks",
	Long:  `Retrieve ECS clusters and services in the active workspace and monitor their task states in real-time.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set up interrupt signal cancel context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		go func() {
			<-sigChan
			cancel()
		}()

		// 1. Initialize AWS provider
		awsProvider, _, err := initAWSProvider(ctx, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// 2. Fetch clusters
		fmt.Println("Fetching ECS clusters...")
		clusters, err := awsProvider.FetchECSClusters(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to fetch clusters: %v\n", err)
			os.Exit(1)
		}

		if len(clusters) == 0 {
			fmt.Println("No ECS clusters found in workspace.")
			os.Exit(0)
		}

		var clusterName string
		if len(clusters) == 1 {
			clusterName = clusters[0].Name
			fmt.Printf("Using ECS Cluster: %s\n", clusterName)
		} else {
			var rows []picker.Row
			for _, c := range clusters {
				rows = append(rows, picker.Row{
					ID: c.Name,
					Fields: []string{
						c.Name,
						c.ARN,
					},
				})
			}
			headers := []string{"Cluster Name", "ARN"}
			selectedID, err := picker.SingleSelect("Select ECS Cluster", headers, rows)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to select cluster: %v\n", err)
				os.Exit(1)
			}
			if selectedID == "" {
				fmt.Println("Selection cancelled")
				os.Exit(0)
			}
			clusterName = selectedID
		}

		// 3. Fetch services
		fmt.Println("Fetching ECS services...")
		services, err := awsProvider.FetchECSServices(ctx, clusterName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to fetch services: %v\n", err)
			os.Exit(1)
		}

		if len(services) == 0 {
			fmt.Printf("No ECS services found in cluster %q.\n", clusterName)
			os.Exit(0)
		}

		var serviceName string
		if len(services) == 1 {
			serviceName = services[0].Name
			fmt.Printf("Using ECS Service: %s\n", serviceName)
		} else {
			var rows []picker.Row
			for _, s := range services {
				rows = append(rows, picker.Row{
					ID: s.Name,
					Fields: []string{
						s.Name,
						s.ARN,
					},
				})
			}
			headers := []string{"Service Name", "ARN"}
			selectedID, err := picker.SingleSelect("Select ECS Service", headers, rows)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to select service: %v\n", err)
				os.Exit(1)
			}
			if selectedID == "" {
				fmt.Println("Selection cancelled")
				os.Exit(0)
			}
			serviceName = selectedID
		}

		// 4. Tasks monitor logic
		if ecsWatchFlag {
			runWatchMode(ctx, awsProvider, clusterName, serviceName)
		} else {
			runSingleView(ctx, awsProvider, clusterName, serviceName)
		}
	},
}

func runSingleView(ctx context.Context, p *aws.Provider, cluster, service string) {
	fmt.Println("Fetching task states...")
	tasks, err := p.FetchECSTasks(ctx, cluster, service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to fetch tasks: %v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Printf("No tasks found for service %q.\n", service)
		os.Exit(0)
	}

	renderTasksTable(tasks, service)
}

func runWatchMode(ctx context.Context, p *aws.Provider, cluster, service string) {
	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")

		tasks, err := p.FetchECSTasks(ctx, cluster, service)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to fetch tasks: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(tasks) == 0 {
			fmt.Printf("No active or stopped tasks found for service %q.\n", service)
		} else {
			renderTasksTable(tasks, service)
		}

		fmt.Println("\nPress Ctrl+C to exit watch mode.")

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
			healthText = dimStyle.Render(t.HealthStatus)
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
			detailCol = dimStyle.Render("Uptime")
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

func init() {
	ecsCmd.Flags().BoolVarP(&ecsWatchFlag, "watch", "w", false, "Enable continuous real-time task status updates")
	rootCmd.AddCommand(ecsCmd)
}
