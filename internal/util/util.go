package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PackageManager represents a system package manager
type PackageManager string

const (
	Unknown PackageManager = "unknown"
	Pacman  PackageManager = "pacman"
	Yay     PackageManager = "yay"
	Apt     PackageManager = "apt"
	Brew    PackageManager = "brew"
	Dnf     PackageManager = "dnf"
	Zypper  PackageManager = "zypper"
)

// DetectPackageManager detects the system package manager
func DetectPackageManager() PackageManager {
	// Check for yay first (Arch with AUR helper)
	if CommandExists("yay") {
		return Yay
	}

	// Check for pacman (Arch)
	if CommandExists("pacman") {
		return Pacman
	}

	// Check for apt (Debian/Ubuntu)
	if CommandExists("apt") {
		return Apt
	}

	// Check for brew (macOS)
	if CommandExists("brew") {
		return Brew
	}

	// Check for dnf (Fedora)
	if CommandExists("dnf") {
		return Dnf
	}

	// Check for zypper (OpenSUSE)
	if CommandExists("zypper") {
		return Zypper
	}

	return Unknown
}

// ParsePackageManager converts a CLI/config value into a package manager.
func ParsePackageManager(value string) (PackageManager, error) {
	switch strings.ToLower(strings.TrimLeft(value, "-")) {
	case "", "auto":
		return Unknown, nil
	case "yay":
		return Yay, nil
	case "pacman":
		return Pacman, nil
	case "apt":
		return Apt, nil
	case "brew", "homebrew", "hombrew":
		return Brew, nil
	case "dnf":
		return Dnf, nil
	case "zypper":
		return Zypper, nil
	default:
		return Unknown, fmt.Errorf("unsupported package manager: %s", value)
	}
}

// ResolvePackageManager returns the selected manager or falls back to detection.
func ResolvePackageManager(selected PackageManager) PackageManager {
	if selected != Unknown {
		return selected
	}
	return DetectPackageManager()
}

// CommandExists checks if a command exists in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// commandExists is an alias for CommandExists for backward compatibility
func commandExists(cmd string) bool {
	return CommandExists(cmd)
}

// Shell returns the user's default shell
func Shell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return shell
}

// ExecuteCommand executes a command and returns its output
func ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
