package main

import (
	"fmt"
	"os"

	"github.com/tumillanino/ezysearch/internal/config"
	"github.com/tumillanino/ezysearch/internal/ui"
)

const version = "1.0.0"

func main() {
	// Check if we're running with any special flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("ezysearch version %s\n", version)
			return
		case "--help", "-h":
			printHelp()
			return
		case "--install":
			// Handle installation
			fmt.Println("Installation functionality will be implemented here")
			return
		case "--config":
			// Generate default config
			generateConfig()
			return
		}
	}

	// Load configuration
	conf, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create and start the UI
	app, err := ui.New(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing UI: %v\n", err)
		os.Exit(1)
	}

	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting application: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`ezysearch - A universal package manager and search tool

Usage: ezysearch [OPTIONS]

Options:
  -h, --help     Show this help message
  -v, --version  Show version information
  --install      Install ezysearch
  --config       Generate default configuration file

Key Bindings:
  Ctrl+P         Search for packages
  Ctrl+G         Search for GitHub repositories
  Ctrl+T         Search for files/directories
  Ctrl+C         Quit

Navigation:
  Arrow keys     Navigate through results
  j/k            Vim-style navigation (up/down)
  g/G            Move to top/bottom of list
  v/V            View package script
  q              Quit application
  /              Focus on search input
  Enter          Execute search or select item
  Esc            Return to previous view

For more information, visit: https://github.com/tumillanino/ezysearch
`)
}

func generateConfig() {
	conf := config.Default()
	if err := conf.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating configuration: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Default configuration file generated at: %s\n", config.ConfigPath())
	fmt.Println("You can now edit this file to customize ezysearch settings.")
}