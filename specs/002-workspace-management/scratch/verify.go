package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/workspace"
)

func main() {
	action := flag.String("action", "list", "action to perform: init, add, delete, rename, use, list")
	name := flag.String("name", "", "workspace name")
	provider := flag.String("provider", "", "cloud provider")
	newName := flag.String("new-name", "", "new workspace name for rename")
	flag.Parse()

	cPath, err := config.Path()
	if err != nil {
		log.Fatalf("Error getting config path: %v", err)
	}

	store := config.New(cPath)
	mgr := workspace.New(store)

	switch *action {
	case "init":
		cfg := config.Default()
		err := store.Save(cfg)
		if err != nil {
			log.Fatalf("Error initializing config: %v", err)
		}
		fmt.Println("Config initialized successfully.")
	case "add":
		if *name == "" || *provider == "" {
			log.Fatal("add action requires --name and --provider")
		}
		err := mgr.Add(*name, *provider, map[string]any{"profile": "default"})
		if err != nil {
			log.Fatalf("Error adding workspace: %v", err)
		}
		fmt.Printf("Workspace %q added successfully.\n", *name)
	case "delete":
		if *name == "" {
			log.Fatal("delete action requires --name")
		}
		err := mgr.Delete(*name)
		if err != nil {
			log.Fatalf("Error deleting workspace: %v", err)
		}
		fmt.Printf("Workspace %q deleted successfully.\n", *name)
	case "rename":
		if *name == "" || *newName == "" {
			log.Fatal("rename action requires --name and --new-name")
		}
		err := mgr.Rename(*name, *newName)
		if err != nil {
			log.Fatalf("Error renaming workspace: %v", err)
		}
		fmt.Printf("Workspace %q renamed to %q successfully.\n", *name, *newName)
	case "use":
		if *name == "" {
			log.Fatal("use action requires --name")
		}
		err := mgr.Use(*name)
		if err != nil {
			log.Fatalf("Error selecting workspace: %v", err)
		}
		fmt.Printf("Now using workspace %q.\n", *name)
	case "list":
		list, err := mgr.List()
		if err != nil {
			log.Fatalf("Error listing workspaces: %v", err)
		}
		fmt.Println("Workspaces:")
		for _, ws := range list {
			activeStr := ""
			if ws.IsActive {
				activeStr = " (active)"
			}
			fmt.Printf("- %s [%s]%s\n", ws.Name, ws.Provider, activeStr)
		}
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}
