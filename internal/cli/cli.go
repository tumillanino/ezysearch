package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/tumillanino/ezysearch/internal/config"
	"github.com/tumillanino/ezysearch/internal/ui"
	"github.com/tumillanino/ezysearch/internal/util"
)

var Version = "1.0.0"

func Main() {
	pkgManager := util.Unknown
	action := ""
	actionValue := ""

	setAction := func(name, value string) {
		if action != "" {
			fmt.Fprintf(os.Stderr, "Error: cannot combine --%s with --%s\n", name, action)
			os.Exit(1)
		}
		action = name
		actionValue = value
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--help" || arg == "-h":
			setAction("help", "")
		case arg == "--version" || arg == "-v" || arg == "-V":
			setAction("version", "")
		case arg == "--install":
			setAction("install", "")
		case arg == "--config":
			setAction("config", "")
		case arg == "--config-path":
			setAction("config-path", "")
		case arg == "--print-config":
			setAction("print-config", "")
		case arg == "--default-config":
			setAction("default-config", "")
		case arg == "--list-package-managers":
			setAction("list-package-managers", "")
		case arg == "--doctor" || arg == "--check":
			setAction("doctor", "")
		case arg == "--completion" || arg == "--completions":
			if i+1 >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "Error: %s requires a shell name (bash or zsh)\n", arg)
				os.Exit(1)
			}
			setAction("completion", os.Args[i+1])
			i++
		case strings.HasPrefix(arg, "--completion="):
			setAction("completion", strings.TrimPrefix(arg, "--completion="))
		case strings.HasPrefix(arg, "--completions="):
			setAction("completion", strings.TrimPrefix(arg, "--completions="))
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

	switch action {
	case "help":
		printHelp()
		return
	case "version":
		fmt.Printf("ezysearch version %s\n", Version)
		return
	case "install":
		// Handle installation
		fmt.Println("Installation functionality will be implemented here")
		return
	case "config":
		generateConfig()
		return
	case "config-path":
		fmt.Println(config.ConfigPath())
		return
	case "print-config":
		printCurrentConfig()
		return
	case "default-config":
		printConfig(config.Default())
		return
	case "list-package-managers":
		printPackageManagers()
		return
	case "doctor":
		runDoctor(pkgManager)
		return
	case "completion":
		printCompletion(actionValue)
		return
	}

	conf, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

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
	fmt.Printf(`ezysearch - A cross-platform package explorer

Usage: ezysearch [OPTIONS]

Options:
  -h, --help                      Show this help message
  -v, -V, --version               Show version information
  --install                       Install ezysearch
  --config                        Write the default configuration file
  --config-path                   Print the configuration file path
  --print-config                  Print the active configuration
  --default-config                Print the default configuration
  --doctor, --check               Check configuration and optional tools
  --completion <shell>            Print shell completions (bash or zsh)
  --list-package-managers         List supported package managers
  --package-manager <manager>     Use a package manager instead of auto-detect
  --manager <manager>             Alias for --package-manager

Package manager flags:
  --auto                          Auto-detect package manager
  --yay, --pacman, --apt          Use a Linux package manager
  --dnf, --zypper                 Use a Linux package manager
  --brew, --homebrew, --hombrew   Use Homebrew packages

Key Bindings:
  Ctrl+P         Search for packages
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

func printCurrentConfig() {
	conf, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	printConfig(conf)
}

func printConfig(conf *config.Settings) {
	if err := toml.NewEncoder(os.Stdout).Encode(conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error printing configuration: %v\n", err)
		os.Exit(1)
	}
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

func printPackageManagers() {
	detected := util.DetectPackageManager()
	managers := []util.PackageManager{
		util.Yay,
		util.Pacman,
		util.Apt,
		util.Brew,
		util.Dnf,
		util.Zypper,
	}

	fmt.Println("Supported package managers:")
	fmt.Println("  auto")
	for _, manager := range managers {
		suffix := ""
		if detected == manager {
			suffix = " (detected)"
		}
		fmt.Printf("  %s%s\n", manager, suffix)
	}

	if detected == util.Unknown {
		fmt.Println("\nNo supported package manager was detected on PATH.")
	}
}

func runDoctor(selected util.PackageManager) {
	hasFailure := false
	configPath := config.ConfigPath()

	fmt.Println("ezysearch doctor")

	conf := config.Default()
	if _, err := os.Stat(configPath); err == nil {
		loaded, loadErr := config.Load()
		if loadErr != nil {
			hasFailure = true
			fmt.Printf("[fail] config: %s (%v)\n", configPath, loadErr)
		} else {
			conf = loaded
			fmt.Printf("[ok]   config: %s\n", configPath)
		}
	} else if os.IsNotExist(err) {
		fmt.Printf("[warn] config: %s does not exist yet; it will be created on first run\n", configPath)
	} else {
		hasFailure = true
		fmt.Printf("[fail] config: cannot inspect %s (%v)\n", configPath, err)
	}

	manager := util.ResolvePackageManager(selected)
	if manager == util.Unknown {
		hasFailure = true
		fmt.Println("[fail] package manager: none detected; use --package-manager <manager> to override")
	} else if selected == util.Unknown {
		fmt.Printf("[ok]   package manager: %s detected\n", manager)
	} else {
		fmt.Printf("[ok]   package manager: %s selected\n", manager)
	}

	for _, command := range requiredCommands(manager) {
		if command == "" {
			continue
		}
		if util.CommandExists(command) {
			fmt.Printf("[ok]   required command: %s\n", command)
		} else {
			hasFailure = true
			fmt.Printf("[fail] required command: %s not found on PATH\n", command)
		}
	}

	if conf.PackageManager.Sudo != "" && manager != util.Unknown && manager != util.Brew {
		printOptionalCommand(conf.PackageManager.Sudo, "configured privilege command")
	}

	if hasFailure {
		os.Exit(1)
	}
}

func requiredCommands(manager util.PackageManager) []string {
	switch manager {
	case util.Apt:
		return []string{"apt", "apt-cache"}
	case util.Pacman:
		return []string{"pacman"}
	case util.Yay:
		return []string{"yay"}
	case util.Brew:
		return []string{"brew"}
	case util.Dnf:
		return []string{"dnf"}
	case util.Zypper:
		return []string{"zypper"}
	default:
		return nil
	}
}

func printOptionalCommand(command, label string) {
	if util.CommandExists(command) {
		fmt.Printf("[ok]   optional command: %s (%s)\n", command, label)
		return
	}
	fmt.Printf("[warn] optional command: %s not found on PATH (%s)\n", command, label)
}

func printCompletion(shell string) {
	switch strings.ToLower(shell) {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported completion shell: %s\n", shell)
		os.Exit(1)
	}
}

const bashCompletion = `#!/usr/bin/env bash

_ezysearch_completion() {
    local cur prev opts managers shells
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="--help --version --install --config --config-path --print-config --default-config --doctor --check --completion --completions --list-package-managers --package-manager --manager --auto --yay --pacman --apt --brew --homebrew --hombrew --dnf --zypper -h -v -V"
    managers="auto yay pacman apt brew homebrew dnf zypper"
    shells="bash zsh"

    if [[ ${prev} == "--package-manager" || ${prev} == "--manager" ]] ; then
        COMPREPLY=( $(compgen -W "${managers}" -- ${cur}) )
        return 0
    fi

    if [[ ${prev} == "--completion" || ${prev} == "--completions" ]] ; then
        COMPREPLY=( $(compgen -W "${shells}" -- ${cur}) )
        return 0
    fi

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}

complete -F _ezysearch_completion ezysearch
`

const zshCompletion = `#compdef ezysearch

local -a opts
opts=(
  '--help[Show help message]'
  '--version[Show version information]'
  '--install[Install ezysearch]'
  '--config[Write the default configuration file]'
  '--config-path[Print the configuration file path]'
  '--print-config[Print the active configuration]'
  '--default-config[Print the default configuration]'
  '--doctor[Check configuration and optional tools]'
  '--check[Check configuration and optional tools]'
  '--completion[Print shell completions]:shell:(bash zsh)'
  '--completions[Print shell completions]:shell:(bash zsh)'
  '--list-package-managers[List supported package managers]'
  '--package-manager[Use a package manager instead of auto-detect]:manager:(auto yay pacman apt brew homebrew dnf zypper)'
  '--manager[Alias for --package-manager]:manager:(auto yay pacman apt brew homebrew dnf zypper)'
  '--auto[Auto-detect package manager]'
  '--yay[Use yay packages]'
  '--pacman[Use pacman packages]'
  '--apt[Use apt packages]'
  '--brew[Use Homebrew packages]'
  '--homebrew[Use Homebrew packages]'
  '--hombrew[Use Homebrew packages]'
  '--dnf[Use dnf packages]'
  '--zypper[Use zypper packages]'
  '-h[Show help message]'
  '-v[Show version information]'
  '-V[Show version information]'
)

_arguments -s $opts
`
