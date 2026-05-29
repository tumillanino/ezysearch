package search

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tumillanino/ezysearch/internal/util"
)

// GetPackageScript retrieves the package script/content for viewing
func GetPackageScript(pkgName string) (string, error) {
	return GetPackageScriptWithManager(pkgName, util.Unknown)
}

// GetPackageScriptWithManager retrieves package script/content using the selected package manager.
func GetPackageScriptWithManager(pkgName string, pkgManager util.PackageManager) (string, error) {
	pkgManager = util.ResolvePackageManager(pkgManager)

	switch pkgManager {
	case util.Yay:
		return getYayPackageScript(pkgName)
	case util.Pacman:
		return getPacmanPackageScript(pkgName)
	case util.Apt:
		return getAptPackageScript(pkgName)
	case util.Brew:
		return getBrewPackageScript(pkgName)
	case util.Dnf:
		return getDnfPackageScript(pkgName)
	case util.Zypper:
		return getZypperPackageScript(pkgName)
	default:
		return "", fmt.Errorf("package script viewing not supported for this package manager")
	}
}

// getYayPackageScript gets the PKGBUILD for a yay package
func getYayPackageScript(pkgName string) (string, error) {
	// Try to get from AUR first
	cmd := exec.Command("yay", "-G", pkgName)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}

	// If that fails, try to get PKGBUILD from local cache
	cmd = exec.Command("yay", "-p", pkgName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not retrieve PKGBUILD for %s (try running yay -Sy to refresh package databases): %v", pkgName, err)
	}

	return string(output), nil
}

// getPacmanPackageScript gets the PKGBUILD for a pacman package
func getPacmanPackageScript(pkgName string) (string, error) {
	// Try to get PKGBUILD from local cache
	cmd := exec.Command("pacman", "-p", pkgName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try alternative method
		cmd = exec.Command("pacman", "-Si", pkgName)
		infoOutput, infoErr := cmd.Output()
		if infoErr != nil {
			return "", fmt.Errorf("could not retrieve package info for %s (try running pacman -Sy to refresh package databases): %v", pkgName, infoErr)
		}

		// Parse the repository to construct URL
		lines := strings.Split(string(infoOutput), "\n")
		var repo, pkgBase string

		for _, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Repository":
				repo = value
			case "Name":
				pkgBase = value
			}
		}

		if repo != "" && pkgBase != "" {
			// Try to fetch PKGBUILD from ABS (Arch Build System)
			url := fmt.Sprintf("https://github.com/archlinux/svntogit-packages/raw/master/%s/trunk/PKGBUILD", pkgBase)
			curlCmd := exec.Command("curl", "-s", url)
			curlOutput, curlErr := curlCmd.Output()
			if curlErr == nil {
				return string(curlOutput), nil
			}
		}

		return "", fmt.Errorf("could not retrieve PKGBUILD for %s", pkgName)
	}

	return string(output), nil
}

// getAptPackageScript gets control information for an apt package
func getAptPackageScript(pkgName string) (string, error) {
	// Get package control information
	cmd := exec.Command("apt-cache", "show", pkgName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not retrieve package information for %s: %v", pkgName, err)
	}

	// Also try to get pre/post installation scripts if available
	// This requires the package to be downloaded first, so we'll just note that
	controlInfo := string(output)

	// Add note about installation scripts
	note := "\n# Note: Pre/Post installation scripts are only available after package download\n"
	note += "# To see installation scripts, you would need to download the package first:\n"
	note += fmt.Sprintf("# apt download %s\n", pkgName)
	note += "# dpkg-deb --control <package>.deb DEBIAN/\n"

	return controlInfo + note, nil
}

// getBrewPackageScript gets the formula for a brew package
func getBrewPackageScript(pkgName string) (string, error) {
	cmd := exec.Command("brew", "formula", pkgName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not retrieve formula for %s: %v", pkgName, err)
	}

	// Get the formula file path
	formulaPath := strings.TrimSpace(string(output))

	// Read the formula file
	catCmd := exec.Command("cat", formulaPath)
	catOutput, catErr := catCmd.Output()
	if catErr != nil {
		return "", fmt.Errorf("could not read formula file for %s: %v", pkgName, catErr)
	}

	return string(catOutput), nil
}

// getDnfPackageScript gets RPM spec information for a dnf package
func getDnfPackageScript(pkgName string) (string, error) {
	// Get package information
	cmd := exec.Command("dnf", "info", pkgName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not retrieve package information for %s: %v", pkgName, err)
	}

	// Try to get source RPM and spec file if available
	// This is more complex and may require downloading the source RPM
	info := string(output)

	// Add note about spec files
	note := "\n# Note: RPM spec files are part of source RPMs\n"
	note += "# To see the full spec file, you would need to:\n"
	note += fmt.Sprintf("# dnf download --source %s\n", pkgName)
	note += "# rpm2cpio <source-package.src.rpm> | cpio -idmv\n"

	return info + note, nil
}

// getZypperPackageScript gets package information for zypper
func getZypperPackageScript(pkgName string) (string, error) {
	// Get package information
	cmd := exec.Command("zypper", "info", pkgName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not retrieve package information for %s: %v", pkgName, err)
	}

	info := string(output)

	// Try to get source information if available
	sourceCmd := exec.Command("zypper", "source-download", "--dry-run", pkgName)
	sourceOutput, sourceErr := sourceCmd.CombinedOutput()
	if sourceErr == nil {
		info += "\n\n# Source Package Information:\n" + string(sourceOutput)
	}

	// Add note about accessing source packages
	note := "\n# Note: To download and examine source packages:\n"
	note += fmt.Sprintf("# zypper source-download %s\n", pkgName)
	note += "# zypper si %s  # Show source info\n"

	return info + note, nil
}
