package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/guppshub/cx-cli/internal/provider/aws"
	"github.com/guppshub/cx-cli/internal/ui/picker"
	"github.com/spf13/cobra"
)

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "Connect to an EC2 instance via SSM",
	Long:  `Retrieve EC2 instances in the active workspace and connect via SSM.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// 1. Initialize AWS provider
		awsProvider, ws, err := initAWSProvider(ctx, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// 2. Fetch EC2 instances
		fmt.Println("Fetching EC2 instances...")
		instances, err := awsProvider.FetchEC2Instances(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(instances) == 0 {
			fmt.Println("No EC2 instances found in workspace.")
			os.Exit(0)
		}

		// 3. Map instances to picker rows
		var rows []picker.Row
		for _, inst := range instances {
			rows = append(rows, picker.Row{
				ID: inst.InstanceID,
				Fields: []string{
					inst.Name,
					inst.InstanceID,
					inst.State,
					inst.PrivateIPAddress,
				},
			})
		}

		// 4. Launch interactive picker
		headers := []string{"Name", "Instance ID", "State", "Private IP"}
		selectedID, err := picker.SingleSelect("Select EC2 Instance", headers, rows)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to run picker: %v\n", err)
			os.Exit(1)
		}

		if selectedID == "" {
			fmt.Println("Selection cancelled")
			os.Exit(0)
		}

		// Find the selected instance to inspect its state
		var targetInst *aws.EC2Instance
		for _, inst := range instances {
			if inst.InstanceID == selectedID {
				targetInst = &inst
				break
			}
		}

		if targetInst == nil {
			fmt.Fprintf(os.Stderr, "Error: selected instance not found\n")
			os.Exit(1)
		}

		if targetInst.State == "stopped" {
			fmt.Fprintf(os.Stderr, "Error: instance %s is stopped. Start the instance before attempting connection.\n", targetInst.InstanceID)
			os.Exit(1)
		}

		// 5. Connect via SSM
		startupCmd := ec2CommandFlag
		if startupCmd == "" && ws != nil {
			if val, ok := ws.Raw["ec2_startup_command"].(string); ok {
				startupCmd = val
			} else if val, ok := ws.Raw["startup_command"].(string); ok {
				startupCmd = val
			}
		}

		fmt.Printf("Connecting to %s (%s) via SSM...\n", targetInst.Name, targetInst.InstanceID)
		err = awsProvider.ConnectSSM(targetInst.InstanceID, startupCmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: connection failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var ec2CommandFlag string

func init() {
	ec2Cmd.Flags().StringVarP(&ec2CommandFlag, "command", "c", "", "Command to execute upon starting the interactive session")
	rootCmd.AddCommand(ec2Cmd)
}
