package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/guppshub/cx-cli/internal/update"
	"github.com/spf13/cobra"
)

var (
	checkFlag bool
	yesFlag   bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check and update the cx CLI to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		fmt.Println("Checking for updates...")
		release, err := update.FetchLatestRelease(ctx, update.LatestReleaseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to fetch latest release: %v\n", err)
			os.Exit(1)
		}

		isNewer := update.IsNewerVersion(Version, release.TagName)

		if checkFlag {
			if isNewer {
				fmt.Printf("A newer version of cx is available: %s (current: %s).\n", release.TagName, Version)
			} else {
				fmt.Printf("You are already running the latest version of cx (%s).\n", Version)
			}
			return
		}

		if !isNewer {
			fmt.Printf("You are already running the latest version of cx (%s).\n", Version)
			return
		}

		fmt.Printf("A newer version of cx is available: %s (current: %s).\n", release.TagName, Version)

		downloadURL, assetName, err := update.FindAsset(release)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if !yesFlag {
			fmt.Print("Would you like to upgrade? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				os.Exit(1)
			}
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Println("Update canceled.")
				return
			}
		}

		fmt.Printf("Downloading latest release asset %q...\n", assetName)
		err = update.SelfUpdate(ctx, downloadURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to install update: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated cx to %s!\n", release.TagName)
	},
}

func init() {
	updateCmd.Flags().BoolVar(&checkFlag, "check", false, "Check for updates without downloading")
	updateCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatic yes to prompts; assume yes to all questions")
	rootCmd.AddCommand(updateCmd)
}
