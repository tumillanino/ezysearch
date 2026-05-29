package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tumillanino/ezysearch/internal/config"
	"github.com/tumillanino/ezysearch/internal/ui"
	"github.com/tumillanino/ezysearch/internal/util"
)

const version = "1.0.0"

func main() {
	pkgManager := util.Unknown

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--version" || arg == "-v":
			fmt.Printf("ezysearch version %s\n", version)
			return
		case arg == "--help" || arg == "-h":
			printHelp()
			return
		case arg == "--install":
			// Handle installation
			fmt.Println("Installation functionality will be implemented here")
			return
		case arg == "--config":
			// Generate default config
			generateConfig()
			return
		case arg == "--package-manager" || arg == "--manager":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "Error: --package-manager requires a value")
				os.Exit(1)
			}
			selected, err := util.ParsePackageManager(os.Args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			pkgManager = selected
			i++
		case strings.HasPrefix(arg, "--package-manager="):
			selected, err := util.ParsePackageManager(strings.TrimPrefix(arg, "--package-manager="))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			pkgManager = selected
		case strings.HasPrefix(arg, "--manager="):
			selected, err := util.ParsePackageManager(strings.TrimPrefix(arg, "--manager="))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			pkgManager = selected
		case strings.HasPrefix(arg, "--"):
			selected, err := util.ParsePackageManager(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			pkgManager = selected
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown argument: %s\n", arg)
			os.Exit(1)
		}
	}

	// Load configuration
	conf, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create and start the UI
	app, err := ui.New(conf, pkgManager)
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
  -h, --help                      Show this help message
  -v, --version                   Show version information
  --install                       Install ezysearch
  --config                        Generate default configuration file
  --package-manager <manager>     Use a package manager instead of auto-detect
  --manager <manager>             Alias for --package-manager

Package manager flags:
  --auto                          Auto-detect package manager
  --yay, --pacman, --apt          Use a Linux package manager
  --dnf, --zypper                 Use a Linux package manager
  --brew, --homebrew, --hombrew   Use Homebrew packages

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
