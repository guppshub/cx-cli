package main

import (
	"fmt"
	"os"

	"github.com/guppshub/cx-cli/internal/ui/picker"
)

func main() {
	items := []picker.Row{
		{ID: "i-0123456789abcdef0", Fields: []string{"bastion-prod", "i-0123456789abcdef0", "running", "10.0.1.25"}},
		{ID: "i-09f87c4f1c901844a", Fields: []string{"payment-worker", "i-09f87c4f1c901844a", "running", "10.0.1.86"}},
		{ID: "i-08a9f24300bfecb21", Fields: []string{"analytics-node", "i-08a9f24300bfecb21", "stopped", "10.0.2.14"}},
		{ID: "i-07c8efb132a0d18bc", Fields: []string{"redis-maintenance", "i-07c8efb132a0d18bc", "running", "10.0.3.41"}},
	}

	headers := []string{"Name", "Instance ID", "State", "Private IP"}

	selected, err := picker.SingleSelect("Select EC2 Instance", headers, items)
	if err != nil {
		fmt.Printf("Error running picker: %v\n", err)
		os.Exit(1)
	}

	if selected == "" {
		fmt.Println("Selection cancelled")
		os.Exit(0)
	}

	fmt.Printf("Successfully selected item ID: %s\n", selected)
}
