package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/guppshub/cx-cli/internal/config"
	"github.com/guppshub/cx-cli/internal/state"
)

func main() {
	action := flag.String("action", "load", "action to perform: load, save-default, print-paths")
	flag.Parse()

	switch *action {
	case "print-paths":
		cPath, err := config.Path()
		if err != nil {
			log.Fatalf("Error getting config path: %v", err)
		}
		sPath, err := state.Path()
		if err != nil {
			log.Fatalf("Error getting state path: %v", err)
		}
		fmt.Printf("Config Path: %s\n", cPath)
		fmt.Printf("State Path: %s\n", sPath)
	case "load":
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
		fmt.Printf("Loaded Config Version: %s\n", cfg.Version)
		fmt.Printf("Contexts count: %d\n", len(cfg.Contexts))
	case "save-default":
		cfg := config.Default()
		err := config.Save(cfg)
		if err != nil {
			log.Fatalf("Error saving config: %v", err)
		}
		fmt.Println("Saved default configuration successfully.")
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}
