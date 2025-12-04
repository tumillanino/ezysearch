package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tumillanino/ezysearch/internal/config"
	"github.com/tumillanino/ezysearch/internal/search"
	"github.com/tumillanino/ezysearch/internal/util"
)

// App represents the main application UI
type App struct {
	conf          *config.Settings
	app           *tview.Application
	flexRoot      *tview.Flex
	inputField    *tview.InputField
	resultList    *tview.List
	detailView    *tview.TextView
	currentMode   string
	selectedItem  string
	searchResults []search.SearchResult
	inScriptView  bool  // Track if we're in script view mode
}

// New creates a new UI application
func New(conf *config.Settings) (*App, error) {
	app := &App{
		conf: conf,
		app:  tview.NewApplication(),
	}

	app.createComponents()
	app.setupKeyBindings()
	return app, nil
}

// createComponents sets up all UI components
func (a *App) createComponents() {
	// Create input field
	a.inputField = tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(50)

	// Create result list
	a.resultList = tview.NewList().
		ShowSecondaryText(true)

	// Create detail view
	a.detailView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	// Add key bindings for scrolling when detail view is focused
	a.detailView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle scrolling and navigation when detail view is focused
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'v':
				// Return to normal view only if we're in script view (not installation)
				if a.inScriptView {
					a.inScriptView = false
					a.app.SetFocus(a.resultList)
					// Refresh the package details
					currentIndex := a.resultList.GetCurrentItem()
					if currentIndex >= 0 {
						a.showDetail(currentIndex)
					}
					return nil
				}
			case 'q':
				// Quit application
				a.app.Stop()
				return nil
			case 'j':
				// Scroll down
				row, col := a.detailView.GetScrollOffset()
				a.detailView.ScrollTo(row+1, col)
				return nil
			case 'k':
				// Scroll up
				row, col := a.detailView.GetScrollOffset()
				if row > 0 {
					a.detailView.ScrollTo(row-1, col)
				}
				return nil
			}
		case tcell.KeyEsc:
			// Return to normal view from script view or installation view
			if a.inScriptView {
				a.inScriptView = false
				a.app.SetFocus(a.resultList)
				// Refresh the package details
				currentIndex := a.resultList.GetCurrentItem()
				if currentIndex >= 0 {
					a.showDetail(currentIndex)
				}
				return nil
			}
			// If we're in installation view, return to results
			if a.currentMode == "package" {
				a.app.SetFocus(a.resultList)
				return nil
			}
		}
		return event
	})

	// Create layout
	a.flexRoot = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.inputField, 3, 1, true).
		AddItem(tview.NewFlex().
			AddItem(a.resultList, 0, 1, false).
			AddItem(a.detailView, 0, 1, false),
			0, 1, false)

	// Set focus to input field
	a.app.SetFocus(a.inputField)
}

// setupKeyBindings sets up key bindings for the application
func (a *App) setupKeyBindings() {
	// Input field key bindings
	a.inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlP:
			a.currentMode = "package"
			a.inputField.SetLabel("Package Search: ")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyCtrlG:
			a.currentMode = "github"
			a.inputField.SetLabel("GitHub Search: ")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyCtrlT:
			a.currentMode = "directory"
			a.inputField.SetLabel("Directory Search: ")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyEnter:
			a.performSearch()
			a.app.SetFocus(a.resultList)
			return nil
		case tcell.KeyEsc:
			a.app.Stop()
			return nil
		}
		
		// Vim keybindings for switching modes
		switch event.Rune() {
		case 'p':
			if event.Modifiers() == tcell.ModCtrl {
				a.currentMode = "package"
				a.inputField.SetLabel("Package Search: ")
				a.app.SetFocus(a.inputField)
				return nil
			}
		case 'g':
			if event.Modifiers() == tcell.ModCtrl {
				a.currentMode = "github"
				a.inputField.SetLabel("GitHub Search: ")
				a.app.SetFocus(a.inputField)
				return nil
			}
		case 't':
			if event.Modifiers() == tcell.ModCtrl {
				a.currentMode = "directory"
				a.inputField.SetLabel("Directory Search: ")
				a.app.SetFocus(a.inputField)
				return nil
			}
		}
		
		return event
	})

	// Result list key bindings
	a.resultList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			a.selectItem()
			return nil
		case tcell.KeyEsc:
			a.app.SetFocus(a.inputField)
			return nil
		}
		
		// Vim keybindings for navigation
		switch event.Rune() {
		case 'j':
			// Move down
			current := a.resultList.GetCurrentItem()
			a.resultList.SetCurrentItem(current + 1)
			return nil
		case 'k':
			// Move up
			current := a.resultList.GetCurrentItem()
			if current > 0 {
				a.resultList.SetCurrentItem(current - 1)
			}
			return nil
		case 'g':
			// Move to top
			a.resultList.SetCurrentItem(0)
			return nil
		case 'G':
			// Move to bottom
			items := a.resultList.GetItemCount()
			if items > 0 {
				a.resultList.SetCurrentItem(items - 1)
			}
			return nil
		case 'q':
			// Quit application
			a.app.Stop()
			return nil
		case '/':
			// Focus on search input
			a.app.SetFocus(a.inputField)
			return nil
		case 'v':
			// View package script (vim-style :view command)
			a.viewPackageScript()
			return nil
		case 'V':
			// View package script (alternative)
			a.viewPackageScript()
			return nil
		}
		
		return event
	})

	// Global key bindings
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		case tcell.KeyCtrlP:
			a.currentMode = "package"
			a.inputField.SetLabel("Package Search: ")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyCtrlG:
			a.currentMode = "github"
			a.inputField.SetLabel("GitHub Search: ")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyCtrlT:
			a.currentMode = "directory"
			a.inputField.SetLabel("Directory Search: ")
			a.app.SetFocus(a.inputField)
			return nil
		}
		
		// Vim keybindings for switching modes globally
		switch event.Rune() {
		case 'p':
			if event.Modifiers() == tcell.ModCtrl {
				a.currentMode = "package"
				a.inputField.SetLabel("Package Search: ")
				a.app.SetFocus(a.inputField)
				return nil
			}
		case 'g':
			if event.Modifiers() == tcell.ModCtrl {
				a.currentMode = "github"
				a.inputField.SetLabel("GitHub Search: ")
				a.app.SetFocus(a.inputField)
				return nil
			}
		case 't':
			if event.Modifiers() == tcell.ModCtrl {
				a.currentMode = "directory"
				a.inputField.SetLabel("Directory Search: ")
				a.app.SetFocus(a.inputField)
				return nil
			}
		}
		
		return event
	})
}

// performSearch executes the search based on current mode
func (a *App) performSearch() {
	query := a.inputField.GetText()
	if query == "" {
		return
	}

	a.resultList.Clear()
	a.detailView.SetText("Searching...")

	go func() {
		var results []search.SearchResult
		var err error

		switch a.currentMode {
		case "package":
			results, err = search.PackageSearch(query)
		case "github":
			results, err = search.GitHubSearch(query, a.conf.GitHubLimit)
		case "directory":
			results, err = search.DirectorySearch(query, a.conf.DirectoryCommand)
		default:
			results, err = search.PackageSearch(query)
		}

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.detailView.SetText(fmt.Sprintf("Error: %v", err))
				return
			}

			a.searchResults = results
			a.displayResults(results)
		})
	}()
}

// displayResults shows search results in the list
func (a *App) displayResults(results []search.SearchResult) {
	a.resultList.Clear()

	for _, result := range results {
		// Add item to list
		a.resultList.AddItem(result.Title, result.Description, 0, nil)
	}

	if len(results) > 0 {
		a.resultList.SetCurrentItem(0)
		// Only load details for the first item
		a.showDetail(0)
	}

	// Set selection handler
	a.resultList.SetSelectedFunc(func(index int, title string, description string, shortcut rune) {
		a.selectItem()
	})

	// Set changed handler to update detail view
	a.resultList.SetChangedFunc(func(index int, title string, description string, shortcut rune) {
		a.showDetail(index)
	})
}

// showDetail displays details for the selected item
func (a *App) showDetail(index int) {
	if index < 0 || index >= len(a.searchResults) {
		a.detailView.SetText("")
		return
	}

	result := a.searchResults[index]
	
	// For package mode, show detailed package information
	if a.currentMode == "package" {
		a.showPackageDetail(result.Value)
		return
	}

	detail := fmt.Sprintf("[green]Title:[white] %s\n\n[green]Description:[white] %s\n\n[green]Value:[white] %s",
		result.Title, result.Description, result.Value)
	a.detailView.SetText(detail)
}

// showPackageDetail displays detailed package information
func (a *App) showPackageDetail(pkgName string) {
	a.detailView.SetText("Loading package details...")
	
	go func() {
		details, err := search.GetPackageDetails(pkgName)
		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.detailView.SetText(fmt.Sprintf("Error loading package details: %v", err))
				return
			}
			
			detail := formatPackageDetails(details)
			a.detailView.SetText(detail)
		})
	}()
}

// formatPackageDetails formats package details for display
func formatPackageDetails(pkg *search.DetailedPackage) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("[green]Package:[white] %s\n", pkg.Name))
	sb.WriteString(fmt.Sprintf("[green]Version:[white] %s\n", pkg.Version))
	
	if pkg.Description != "" {
		sb.WriteString(fmt.Sprintf("[green]Description:[white] %s\n", pkg.Description))
	}
	
	if pkg.Repository != "" {
		sb.WriteString(fmt.Sprintf("[green]Repository:[white] %s\n", pkg.Repository))
	}
	
	if pkg.Architecture != "" {
		sb.WriteString(fmt.Sprintf("[green]Architecture:[white] %s\n", pkg.Architecture))
	}
	
	if pkg.URL != "" {
		sb.WriteString(fmt.Sprintf("[green]URL:[white] %s\n", pkg.URL))
	}
	
	if len(pkg.Licenses) > 0 {
		sb.WriteString(fmt.Sprintf("[green]Licenses:[white] %s\n", strings.Join(pkg.Licenses, ", ")))
	}
	
	if pkg.DownloadSize != "" {
		sb.WriteString(fmt.Sprintf("[green]Download Size:[white] %s\n", pkg.DownloadSize))
	}
	
	if pkg.InstalledSize != "" {
		sb.WriteString(fmt.Sprintf("[green]Installed Size:[white] %s\n", pkg.InstalledSize))
	}
	
	if pkg.Packager != "" {
		sb.WriteString(fmt.Sprintf("[green]Packager:[white] %s\n", pkg.Packager))
	}
	
	if pkg.BuildDate != "" {
		sb.WriteString(fmt.Sprintf("[green]Build Date:[white] %s\n", pkg.BuildDate))
	}
	
	if pkg.InstallDate != "" {
		sb.WriteString(fmt.Sprintf("[green]Install Date:[white] %s\n", pkg.InstallDate))
	}
	
	if len(pkg.Groups) > 0 {
		sb.WriteString(fmt.Sprintf("[green]Groups:[white] %s\n", strings.Join(pkg.Groups, ", ")))
	}
	
	if len(pkg.Provides) > 0 {
		sb.WriteString(fmt.Sprintf("[green]Provides:[white] %s\n", strings.Join(pkg.Provides, ", ")))
	}
	
	if len(pkg.DependsOn) > 0 {
		sb.WriteString(fmt.Sprintf("[green]Depends On:[white] %s\n", strings.Join(pkg.DependsOn, ", ")))
	}
	
	if len(pkg.Conflicts) > 0 {
		sb.WriteString(fmt.Sprintf("[green]Conflicts With:[white] %s\n", strings.Join(pkg.Conflicts, ", ")))
	}
	
	if len(pkg.Replaces) > 0 {
		sb.WriteString(fmt.Sprintf("[green]Replaces:[white] %s\n", strings.Join(pkg.Replaces, ", ")))
	}
	
	return sb.String()
}

// selectItem handles item selection
func (a *App) selectItem() {
	index := a.resultList.GetCurrentItem()
	if index < 0 || index >= len(a.searchResults) {
		return
	}

	result := a.searchResults[index]
	a.selectedItem = result.Value

	// Handle selection based on mode
	switch a.currentMode {
	case "package":
		// Show confirmation dialog before installation
		a.showInstallConfirmation(result.Value)
	case "github":
		a.cloneRepository(result.Value)
	case "directory":
		a.openFile(result.Value)
	}
}

// showInstallConfirmation displays a confirmation dialog before installation
func (a *App) showInstallConfirmation(pkgName string) {
	// Create a modal dialog for confirmation
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to install package '%s'?", pkgName)).
		AddButtons([]string{"Install", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Install" {
				a.installPackage(pkgName)
			}
			// Remove the modal and return focus to the main view
			a.app.SetRoot(a.flexRoot, true)
			a.app.SetFocus(a.resultList)
		})

	// Display the modal
	a.app.SetRoot(modal, false)
}

// installPackage installs a package using the appropriate package manager
func (a *App) installPackage(pkgName string) {
	pkgManager := util.DetectPackageManager()
	var cmd *exec.Cmd

	// Use the configured sudo command
	sudoCmd := a.conf.PackageManager.Sudo

	// Check if sudo is needed and provide user guidance
	needsSudo := false
	var fullCommand string

	switch pkgManager {
	case util.Yay:
		// Yay doesn't need sudo
		args := append([]string{"-S", pkgName}, a.conf.PackageManager.YayFlags...)
		cmd = exec.Command("yay", args...)
		fullCommand = fmt.Sprintf("yay -S %s", pkgName)
	case util.Pacman:
		args := append([]string{"pacman", "-S", pkgName}, a.conf.PackageManager.PacmanFlags...)
		cmd = exec.Command(sudoCmd, args...)
		fullCommand = fmt.Sprintf("%s pacman -S %s", sudoCmd, pkgName)
		needsSudo = true
	case util.Apt:
		args := append([]string{"apt", "install", pkgName}, a.conf.PackageManager.AptFlags...)
		cmd = exec.Command(sudoCmd, args...)
		fullCommand = fmt.Sprintf("%s apt install %s", sudoCmd, pkgName)
		needsSudo = true
	case util.Brew:
		// Brew doesn't need sudo on modern macOS
		args := append([]string{"install", pkgName}, a.conf.PackageManager.BrewFlags...)
		cmd = exec.Command("brew", args...)
		fullCommand = fmt.Sprintf("brew install %s", pkgName)
	case util.Dnf:
		args := append([]string{"dnf", "install", pkgName}, a.conf.PackageManager.DnfFlags...)
		cmd = exec.Command(sudoCmd, args...)
		fullCommand = fmt.Sprintf("%s dnf install %s", sudoCmd, pkgName)
		needsSudo = true
	case util.Zypper:
		args := append([]string{"zypper", "install", pkgName}, a.conf.PackageManager.ZypperFlags...)
		cmd = exec.Command(sudoCmd, args...)
		fullCommand = fmt.Sprintf("%s zypper install %s", sudoCmd, pkgName)
		needsSudo = true
	default:
		a.detailView.SetText("No supported package manager found")
		return
	}

	// Show installation header with navigation instructions
	header := fmt.Sprintf("[:: Installing package: %s ::]\n", pkgName)
	header += "[:: View real-time output below ::]\n"
	header += "[:: Press 'q' to quit, 'Esc' to return to search ::]\n\n"
	
	if needsSudo {
		header += fmt.Sprintf("Command: %s\n", fullCommand)
		header += "Note: System may prompt for password in terminal\n\n"
	} else {
		header += fmt.Sprintf("Command: %s\n\n", fullCommand)
	}
	
	a.detailView.SetText(header)
	
	// Create a pipe to capture command output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.detailView.SetText(fmt.Sprintf("%sError creating stdout pipe: %v", header, err))
		return
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.detailView.SetText(fmt.Sprintf("%sError creating stderr pipe: %v", header, err))
		return
	}
	
	// Set proper environment to handle terminal interaction
	cmd.Stdin = os.Stdin
	
	// Start the command
	if err := cmd.Start(); err != nil {
		a.detailView.SetText(fmt.Sprintf("%sError starting installation: %v", header, err))
		return
	}
	
	// Move focus to detail view to show output
	a.app.SetFocus(a.detailView)
	
	// Create a multi-reader to combine stdout and stderr
	outputReader := io.MultiReader(stdout, stderr)
	
	// Read output in a goroutine and update the UI
	go func() {
		scanner := bufio.NewScanner(outputReader)
		for scanner.Scan() {
			line := scanner.Text()
			a.app.QueueUpdateDraw(func() {
				currentText := a.detailView.GetText(true)
				// Append new line to existing content
				newText := currentText + line + "\n"
				a.detailView.SetText(newText)
				// Scroll to bottom
				a.detailView.ScrollToEnd()
			})
		}
		
		// Wait for command to finish
		err := cmd.Wait()
		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.detailView.SetText(fmt.Sprintf("%s\nInstallation completed with errors: %v", a.detailView.GetText(true), err))
			} else {
				a.detailView.SetText(fmt.Sprintf("%s\nInstallation completed successfully!", a.detailView.GetText(true)))
			}
			
			// Add final instructions
			finalText := a.detailView.GetText(true)
			finalText += "\n[:: Press 'Esc' to return to search ::]\n"
			a.detailView.SetText(finalText)
		})
	}()
}

// cloneRepository clones a GitHub repository
func (a *App) cloneRepository(url string) {
	cmd := exec.Command("git", "clone", url)
	a.detailView.SetText(fmt.Sprintf("Cloning repository...\n\nRunning: %s", strings.Join(cmd.Args, " ")))
	
	// In a real implementation, you would execute the command and show output
	// For now, we'll just show what would be executed
}

// viewPackageScript displays the package script/content for the selected package
func (a *App) viewPackageScript() {
	if a.currentMode != "package" {
		a.detailView.SetText("Package script viewing is only available in package search mode")
		return
	}

	// If we're already in script view, toggle back to normal view
	if a.inScriptView {
		a.inScriptView = false
		// Return focus to result list
		a.app.SetFocus(a.resultList)
		// Refresh the package details
		currentIndex := a.resultList.GetCurrentItem()
		if currentIndex >= 0 {
			a.showDetail(currentIndex)
		}
		return
	}

	index := a.resultList.GetCurrentItem()
	if index < 0 || index >= len(a.searchResults) {
		return
	}

	result := a.searchResults[index]
	pkgName := result.Value

	a.detailView.SetText("Loading package script...")

	// Set script view mode
	a.inScriptView = true

	go func() {
		script, err := search.GetPackageScript(pkgName)
		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.detailView.SetText(fmt.Sprintf("Error loading package script: %v", err))
				return
			}

			// Format the script with a header
			header := fmt.Sprintf("[:: Package Script for %s ::]\n", pkgName)
			header += "[:: Use arrow keys to scroll, press 'v' again to return ::]\n\n"
			a.detailView.SetText(header + script)
			
			// Move focus to detail view for scrolling
			a.app.SetFocus(a.detailView)
		})
	}()
}

// openFile opens a file or directory
func (a *App) openFile(path string) {
	// For now, just show the path
	a.detailView.SetText(fmt.Sprintf("Selected file: %s", path))
}

// Start runs the application
func (a *App) Start() error {
	a.currentMode = "package"
	a.inputField.SetLabel("Package Search: ")
	return a.app.SetRoot(a.flexRoot, true).EnableMouse(true).Run()
}