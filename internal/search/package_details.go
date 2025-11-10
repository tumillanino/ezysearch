package search

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/tumillanino/ezysearch/internal/util"
)

// DetailedPackage represents a package with detailed information
type DetailedPackage struct {
	Name         string
	Version      string
	Description  string
	Repository   string
	Architecture string
	URL          string
	Licenses     []string
	Groups       []string
	Provides     []string
	DependsOn    []string
	Conflicts    []string
	Replaces     []string
	DownloadSize string
	InstalledSize string
	Packager     string
	BuildDate    string
	InstallDate  string
	InstallReason string
	InstallScript bool
	ValidatedBy  []string
}

// GetPackageDetails retrieves detailed information about a package
func GetPackageDetails(pkgName string) (*DetailedPackage, error) {
	pkgManager := util.DetectPackageManager()
	
	switch pkgManager {
	case util.Yay:
		return getYayPackageDetails(pkgName)
	case util.Pacman:
		return getPacmanPackageDetails(pkgName)
	case util.Apt:
		return getAptPackageDetails(pkgName)
	case util.Brew:
		return getBrewPackageDetails(pkgName)
	case util.Dnf:
		return getDnfPackageDetails(pkgName)
	default:
		return nil, fmt.Errorf("unsupported package manager")
	}
}

// getYayPackageDetails gets detailed package information for yay
func getYayPackageDetails(pkgName string) (*DetailedPackage, error) {
	cmd := exec.Command("yay", "-Si", pkgName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return parsePacmanInfo(string(output)), nil
}

// getPacmanPackageDetails gets detailed package information for pacman
func getPacmanPackageDetails(pkgName string) (*DetailedPackage, error) {
	cmd := exec.Command("pacman", "-Si", pkgName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return parsePacmanInfo(string(output)), nil
}

// getAptPackageDetails gets detailed package information for apt
func getAptPackageDetails(pkgName string) (*DetailedPackage, error) {
	cmd := exec.Command("apt-cache", "show", pkgName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return parseAptInfo(string(output)), nil
}

// getBrewPackageDetails gets detailed package information for brew
func getBrewPackageDetails(pkgName string) (*DetailedPackage, error) {
	cmd := exec.Command("brew", "info", "--json", pkgName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return parseBrewInfo(string(output))
}

// getDnfPackageDetails gets detailed package information for dnf
func getDnfPackageDetails(pkgName string) (*DetailedPackage, error) {
	cmd := exec.Command("dnf", "info", pkgName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return parseDnfInfo(string(output)), nil
}

// parsePacmanInfo parses pacman package information
func parsePacmanInfo(output string) *DetailedPackage {
	pkg := &DetailedPackage{}
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "Name":
			pkg.Name = value
		case "Version":
			pkg.Version = value
		case "Description":
			pkg.Description = value
		case "Architecture":
			pkg.Architecture = value
		case "URL":
			pkg.URL = value
		case "Licenses":
			pkg.Licenses = strings.Fields(value)
		case "Groups":
			if value != "None" {
				pkg.Groups = strings.Fields(value)
			}
		case "Provides":
			if value != "None" {
				pkg.Provides = strings.Fields(value)
			}
		case "Depends On":
			if value != "None" {
				pkg.DependsOn = strings.Fields(value)
			}
		case "Conflicts With":
			if value != "None" {
				pkg.Conflicts = strings.Fields(value)
			}
		case "Replaces":
			if value != "None" {
				pkg.Replaces = strings.Fields(value)
			}
		case "Download Size":
			pkg.DownloadSize = value
		case "Installed Size":
			pkg.InstalledSize = value
		case "Packager":
			pkg.Packager = value
		case "Build Date":
			pkg.BuildDate = value
		}
	}
	
	return pkg
}

// parseAptInfo parses apt package information
func parseAptInfo(output string) *DetailedPackage {
	pkg := &DetailedPackage{}
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "Package":
			pkg.Name = value
		case "Version":
			pkg.Version = value
		case "Description":
			if pkg.Description == "" {
				pkg.Description = value
			}
		case "Homepage":
			pkg.URL = value
		case "Installed-Size":
			pkg.InstalledSize = value
		}
	}
	
	return pkg
}

// parseBrewInfo parses brew package information
func parseBrewInfo(output string) (*DetailedPackage, error) {
	var brewInfo []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &brewInfo); err != nil {
		return nil, err
	}
	
	if len(brewInfo) == 0 {
		return nil, fmt.Errorf("no package information found")
	}
	
	pkgInfo := brewInfo[0]
	pkg := &DetailedPackage{
		Name:        getStringValue(pkgInfo, "name"),
		Version:     getStringValue(pkgInfo, "versions.stable"),
		Description: getStringValue(pkgInfo, "desc"),
		URL:         getStringValue(pkgInfo, "homepage"),
	}
	
	return pkg, nil
}

// parseDnfInfo parses dnf package information
func parseDnfInfo(output string) *DetailedPackage {
	pkg := &DetailedPackage{}
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "Name":
			pkg.Name = value
		case "Version":
			pkg.Version = value
		case "Release":
			if pkg.Version != "" {
				pkg.Version += "-" + value
			} else {
				pkg.Version = value
			}
		case "Summary":
			pkg.Description = value
		case "URL":
			pkg.URL = value
		case "License":
			pkg.Licenses = []string{value}
		case "Architecture":
			pkg.Architecture = value
		case "Size":
			pkg.InstalledSize = value
		}
	}
	
	return pkg
}

// getStringValue safely extracts a string value from a nested map
func getStringValue(m map[string]interface{}, path string) string {
	keys := strings.Split(path, ".")
	current := m
	
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key]; ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
			return ""
		}
		
		if val, ok := current[key]; ok {
			if next, ok := val.(map[string]interface{}); ok {
				current = next
			} else {
				return ""
			}
		} else {
			return ""
		}
	}
	
	return ""
}