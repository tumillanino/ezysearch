package search

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tumillanino/ezysearch/internal/util"
)

// Package represents a package from any package manager
type Package struct {
	Name        string
	Description string
	Version     string
	Repository  string
}

// SearchResult represents a generic search result
type SearchResult struct {
	Title       string
	Description string
	Value       string
}

// PackageSearch searches for packages using the system package manager
func PackageSearch(query string) ([]SearchResult, error) {
	return PackageSearchWithManager(query, util.Unknown)
}

// PackageSearchWithManager searches packages using the selected package manager.
func PackageSearchWithManager(query string, pkgManager util.PackageManager) ([]SearchResult, error) {
	pkgManager = util.ResolvePackageManager(pkgManager)

	switch pkgManager {
	case util.Yay:
		return yaySearch(query)
	case util.Pacman:
		return pacmanSearch(query)
	case util.Apt:
		return aptSearch(query)
	case util.Brew:
		return brewSearch(query)
	case util.Dnf:
		return dnfSearch(query)
	case util.Zypper:
		return zypperSearch(query)
	default:
		return nil, fmt.Errorf("no supported package manager found")
	}
}

// yaySearch searches for packages using yay
func yaySearch(query string) ([]SearchResult, error) {
	cmd := exec.Command("yay", "-Slq")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return filterPackages(strings.Split(strings.TrimSpace(string(output)), "\n"), query), nil
}

// pacmanSearch searches for packages using pacman
func pacmanSearch(query string) ([]SearchResult, error) {
	cmd := exec.Command("pacman", "-Sl")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var packages []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			packages = append(packages, parts[1])
		}
	}

	return filterPackages(packages, query), nil
}

// aptSearch searches for packages using apt
func aptSearch(query string) ([]SearchResult, error) {
	cmd := exec.Command("apt-cache", "pkgnames")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return filterPackages(strings.Split(strings.TrimSpace(string(output)), "\n"), query), nil
}

// brewSearch searches for packages using brew
func brewSearch(query string) ([]SearchResult, error) {
	cmd := exec.Command("brew", "formulae")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return filterPackages(strings.Split(strings.TrimSpace(string(output)), "\n"), query), nil
}

// dnfSearch searches for packages using dnf
func dnfSearch(query string) ([]SearchResult, error) {
	cmd := exec.Command("dnf", "list", "available")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var packages []string
	// Skip the first line (header)
	for i, line := range lines {
		if i == 0 {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			// Remove architecture suffix if present
			pkgName := strings.Split(parts[0], ".")[0]
			packages = append(packages, pkgName)
		}
	}

	return filterPackages(packages, query), nil
}

// zypperSearch searches for packages using zypper
func zypperSearch(query string) ([]SearchResult, error) {
	cmd := exec.Command("zypper", "search", query)
	output, err := cmd.Output()
	if err != nil {
		// Zypper returns non-zero exit code even for successful searches with no results
		// So we need to check if it's actually an error
		if !strings.Contains(err.Error(), "exit status 104") {
			return nil, err
		}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var packages []string

	// Parse zypper search output
	// Skip header lines and parse package names
	for _, line := range lines {
		// Skip empty lines and header lines
		if line == "" || strings.HasPrefix(line, "S |") || strings.HasPrefix(line, "--+") {
			continue
		}

		// Parse the line: "i | package-name | summary"
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			pkgName := strings.TrimSpace(parts[1])
			packages = append(packages, pkgName)
		}
	}

	return filterPackages(packages, query), nil
}

// filterPackages filters packages based on query and sorts them
func filterPackages(packages []string, query string) []SearchResult {
	var exactMatches []SearchResult
	var prefixMatches []SearchResult
	var substringMatches []SearchResult

	query = strings.ToLower(query)

	for _, pkg := range packages {
		pkgLower := strings.ToLower(pkg)
		if query == "" {
			substringMatches = append(substringMatches, SearchResult{
				Title:       pkg,
				Description: "",
				Value:       pkg,
			})
		} else if pkgLower == query {
			// Exact match
			exactMatches = append(exactMatches, SearchResult{
				Title:       pkg,
				Description: "",
				Value:       pkg,
			})
		} else if strings.HasPrefix(pkgLower, query) {
			// Prefix match
			prefixMatches = append(prefixMatches, SearchResult{
				Title:       pkg,
				Description: "",
				Value:       pkg,
			})
		} else if strings.Contains(pkgLower, query) {
			// Substring match
			substringMatches = append(substringMatches, SearchResult{
				Title:       pkg,
				Description: "",
				Value:       pkg,
			})
		}
	}

	// Combine results in order of preference: exact, prefix, substring
	results := append(exactMatches, prefixMatches...)
	results = append(results, substringMatches...)

	return results
}
