package util

import (
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

	return Unknown
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