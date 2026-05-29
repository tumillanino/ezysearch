package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	header        *tview.TextView
	inputField    *tview.InputField
	resultList    *tview.List
	detailView    *tview.TextView
	statusBar     *tview.TextView
	currentMode   string
	selectedItem  string
	searchResults []search.SearchResult
	pkgManager    util.PackageManager
	inScriptView  bool // Track if we're in script view mode
}

const (
	colorAccent = tcell.ColorSteelBlue
	colorMuted  = tcell.ColorGray
	colorPanel  = tcell.ColorDarkSlateGray
)

// New creates a new UI application
func New(conf *config.Settings, pkgManager ...util.PackageManager) (*App, error) {
	selectedManager := util.Unknown
	if len(pkgManager) > 0 {
		selectedManager = pkgManager[0]
	}

	app := &App{
		conf:        conf,
		app:         tview.NewApplication(),
		currentMode: "package",
		pkgManager:  selectedManager,
	}

	app.createComponents()
	app.setupKeyBindings()
	return app, nil
}

// createComponents sets up all UI components
func (a *App) createComponents() {
	a.header = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	a.header.SetText("[::b] ezysearch [-:-:-] [gray]Package, GitHub, and file search")

	a.inputField = tview.NewInputField().
		SetFieldWidth(0).
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(colorAccent).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				a.performSearch()
				a.app.SetFocus(a.resultList)
			}
		})
	a.inputField.SetBorder(true).
		SetBorderColor(colorPanel).
		SetTitleColor(colorAccent)

	a.resultList = tview.NewList().
		ShowSecondaryText(true).
		SetMainTextColor(tcell.ColorWhite).
		SetSecondaryTextColor(colorMuted).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(colorAccent).
		SetHighlightFullLine(true)
	a.resultList.SetBorder(true).
		SetBorderColor(colorPanel).
		SetTitle(" Results ").
		SetTitleColor(colorAccent)

	a.detailView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	a.detailView.SetBorder(true).
		SetBorderColor(colorPanel).
		SetTitle(" Details ").
		SetTitleColor(colorAccent)

	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextColor(colorMuted)

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

	a.flexRoot = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.header, 1, 0, false).
		AddItem(a.inputField, 3, 0, true).
		AddItem(tview.NewFlex().
			AddItem(a.resultList, 0, 2, false).
			AddItem(a.detailView, 0, 3, false),
			0, 1, false).
		AddItem(a.statusBar, 1, 0, false)

	a.setMode("package")
	a.setStatus("Type a query, press Enter to search. Ctrl+P packages, Ctrl+G GitHub, Ctrl+T files.")
	a.detailView.SetText("[gray]Search results will appear on the left. Select an item to preview details here.")
	a.app.SetFocus(a.inputField)
}

func (a *App) setMode(mode string) {
	a.currentMode = mode

	var label, title string
	switch mode {
	case "github":
		label = " GitHub "
		title = " Search GitHub repositories "
	case "directory":
		label = " Files "
		title = " Search files and directories "
	default:
		a.currentMode = "package"
		label = " Package "
		title = fmt.Sprintf(" Search packages (%s) ", a.packageManagerLabel())
	}

	a.inputField.SetLabel(label)
	a.inputField.SetTitle(title)
	a.setStatus("Enter searches, j/k moves, / focuses search, v views package scripts, q quits.")
}

func (a *App) setStatus(message string) {
	a.statusBar.SetText("[gray] " + message)
}

func (a *App) packageManager() util.PackageManager {
	return util.ResolvePackageManager(a.pkgManager)
}

func (a *App) packageManagerLabel() string {
	if a.pkgManager == util.Unknown {
		return "auto"
	}
	return string(a.pkgManager)
}

// setupKeyBindings sets up key bindings for the application
func (a *App) setupKeyBindings() {
	// Input field key bindings
	a.inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlP:
			a.setMode("package")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyCtrlG:
			a.setMode("github")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyCtrlT:
			a.setMode("directory")
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
				a.setMode("package")
				a.app.SetFocus(a.inputField)
				return nil
			}
		case 'g':
			if event.Modifiers() == tcell.ModCtrl {
				a.setMode("github")
				a.app.SetFocus(a.inputField)
				return nil
			}
		case 't':
			if event.Modifiers() == tcell.ModCtrl {
				a.setMode("directory")
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
			a.setMode("package")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyCtrlG:
			a.setMode("github")
			a.app.SetFocus(a.inputField)
			return nil
		case tcell.KeyCtrlT:
			a.setMode("directory")
			a.app.SetFocus(a.inputField)
			return nil
		}

		// Vim keybindings for switching modes globally
		switch event.Rune() {
		case 'p':
			if event.Modifiers() == tcell.ModCtrl {
				a.setMode("package")
				a.app.SetFocus(a.inputField)
				return nil
			}
		case 'g':
			if event.Modifiers() == tcell.ModCtrl {
				a.setMode("github")
				a.app.SetFocus(a.inputField)
				return nil
			}
		case 't':
			if event.Modifiers() == tcell.ModCtrl {
				a.setMode("directory")
				a.app.SetFocus(a.inputField)
				return nil
			}
		}

		return event
	})
}

// performSearch executes the search based on current mode
func (a *App) performSearch() {
	query := strings.TrimSpace(a.inputField.GetText())
	if query == "" {
		a.setStatus("Enter a search term first.")
		return
	}

	searchMode := a.currentMode
	a.resultList.Clear()
	a.searchResults = nil
	a.detailView.SetText(fmt.Sprintf("[yellow]Searching %s for:[white] %s", searchMode, tview.Escape(query)))
	a.setStatus("Searching...")

	go func() {
		var results []search.SearchResult
		var err error

		switch searchMode {
		case "package":
			results, err = search.PackageSearchWithManager(query, a.packageManager())
		case "github":
			results, err = search.GitHubSearch(query, a.conf.GitHubLimit)
		case "directory":
			results, err = search.DirectorySearch(query, a.conf.DirectoryCommand)
		default:
			results, err = search.PackageSearchWithManager(query, a.packageManager())
		}

		a.app.QueueUpdateDraw(func() {
			if a.currentMode != searchMode {
				return
			}
			if err != nil {
				a.detailView.SetText(fmt.Sprintf("[red]Error:[white] %v", err))
				a.setStatus("Search failed.")
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

	showSecondaryText := false
	for _, result := range results {
		if strings.TrimSpace(result.Description) != "" {
			showSecondaryText = true
			break
		}
	}
	a.resultList.ShowSecondaryText(showSecondaryText)

	for _, result := range results {
		description := result.Description
		a.resultList.AddItem(result.Title, description, 0, nil)
	}

	if len(results) > 0 {
		a.resultList.SetCurrentItem(0)
		// Only load details for the first item
		a.showDetail(0)
		a.setStatus(fmt.Sprintf("%d result(s). Enter selects, / edits search, q quits.", len(results)))
	} else {
		a.detailView.SetText("[gray]No results found.")
		a.setStatus("No results found.")
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

	if a.currentMode == "directory" {
		a.showDirectoryDetail(result.Value)
		return
	}

	detail := fmt.Sprintf("[blue]Repository[white]\n%s\n\n[blue]Description[white]\n%s\n\n[blue]Clone URL[white]\n%s",
		tview.Escape(result.Title), tview.Escape(result.Description), tview.Escape(result.Value))
	a.detailView.SetText(detail)
}

// showPackageDetail displays detailed package information
func (a *App) showPackageDetail(pkgName string) {
	a.detailView.SetText("Loading package details...")

	go func() {
		details, err := search.GetPackageDetailsWithManager(pkgName, a.packageManager())
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

func (a *App) showDirectoryDetail(path string) {
	info, err := os.Stat(path)
	if err != nil {
		a.detailView.SetText(fmt.Sprintf("[red]Could not inspect path:[white] %v", err))
		return
	}

	var sb strings.Builder
	sb.WriteString("[blue]Path[white]\n")
	sb.WriteString(tview.Escape(path))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("[blue]Type:[white] %s\n", fileKind(info)))
	sb.WriteString(fmt.Sprintf("[blue]Size:[white] %s\n", formatBytes(info.Size())))
	sb.WriteString(fmt.Sprintf("[blue]Modified:[white] %s\n", info.ModTime().Format(time.RFC1123)))

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err == nil {
			sb.WriteString("\n[blue]Contents[white]\n")
			for i, entry := range entries {
				if i >= 40 {
					sb.WriteString("...\n")
					break
				}
				name := entry.Name()
				if entry.IsDir() {
					name += string(os.PathSeparator)
				}
				sb.WriteString(tview.Escape(name))
				sb.WriteString("\n")
			}
		}
		a.detailView.SetText(sb.String())
		return
	}

	if a.conf.PreviewCommand == "" {
		a.detailView.SetText(sb.String())
		return
	}

	a.detailView.SetText(sb.String() + "\n[gray]Loading preview...")
	go func() {
		preview, err := previewFile(path, a.conf.PreviewCommand)
		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.detailView.SetText(sb.String() + fmt.Sprintf("\n[yellow]Preview unavailable:[white] %v", err))
				return
			}
			a.detailView.SetText(sb.String() + "\n[blue]Preview[white]\n" + tview.Escape(preview))
		})
	}()
}

// formatPackageDetails formats package details for display
func formatPackageDetails(pkg *search.DetailedPackage) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[blue]Package[white]\n%s\n\n", tview.Escape(pkg.Name)))
	sb.WriteString(fmt.Sprintf("[blue]Version[white]\n%s\n", tview.Escape(pkg.Version)))

	if pkg.Description != "" {
		sb.WriteString(fmt.Sprintf("\n[blue]Description[white]\n%s\n", tview.Escape(pkg.Description)))
	}

	if pkg.Repository != "" {
		sb.WriteString(fmt.Sprintf("\n[blue]Repository[white]\n%s\n", tview.Escape(pkg.Repository)))
	}

	if pkg.Architecture != "" {
		sb.WriteString(fmt.Sprintf("[blue]Architecture:[white] %s\n", tview.Escape(pkg.Architecture)))
	}

	if pkg.URL != "" {
		sb.WriteString(fmt.Sprintf("[blue]URL:[white] %s\n", tview.Escape(pkg.URL)))
	}

	if len(pkg.Licenses) > 0 {
		sb.WriteString(fmt.Sprintf("[blue]Licenses:[white] %s\n", tview.Escape(strings.Join(pkg.Licenses, ", "))))
	}

	if pkg.DownloadSize != "" {
		sb.WriteString(fmt.Sprintf("[blue]Download Size:[white] %s\n", tview.Escape(pkg.DownloadSize)))
	}

	if pkg.InstalledSize != "" {
		sb.WriteString(fmt.Sprintf("[blue]Installed Size:[white] %s\n", tview.Escape(pkg.InstalledSize)))
	}

	if pkg.Packager != "" {
		sb.WriteString(fmt.Sprintf("[blue]Packager:[white] %s\n", tview.Escape(pkg.Packager)))
	}

	if pkg.BuildDate != "" {
		sb.WriteString(fmt.Sprintf("[blue]Build Date:[white] %s\n", tview.Escape(pkg.BuildDate)))
	}

	if pkg.InstallDate != "" {
		sb.WriteString(fmt.Sprintf("[blue]Install Date:[white] %s\n", tview.Escape(pkg.InstallDate)))
	}

	if len(pkg.Groups) > 0 {
		sb.WriteString(fmt.Sprintf("[blue]Groups:[white] %s\n", tview.Escape(strings.Join(pkg.Groups, ", "))))
	}

	if len(pkg.Provides) > 0 {
		sb.WriteString(fmt.Sprintf("[blue]Provides:[white] %s\n", tview.Escape(strings.Join(pkg.Provides, ", "))))
	}

	if len(pkg.DependsOn) > 0 {
		sb.WriteString(fmt.Sprintf("[blue]Depends On:[white] %s\n", tview.Escape(strings.Join(pkg.DependsOn, ", "))))
	}

	if len(pkg.Conflicts) > 0 {
		sb.WriteString(fmt.Sprintf("[blue]Conflicts With:[white] %s\n", tview.Escape(strings.Join(pkg.Conflicts, ", "))))
	}

	if len(pkg.Replaces) > 0 {
		sb.WriteString(fmt.Sprintf("[blue]Replaces:[white] %s\n", tview.Escape(strings.Join(pkg.Replaces, ", "))))
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
		if a.conf.PackageManager.ConfirmInstall {
			a.showInstallConfirmation(result.Value)
			return
		}
		a.installPackage(result.Value)
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
	pkgManager := a.packageManager()
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
	a.runCommandInDetail(cmd, "Cloning repository", "Repository cloned successfully.")
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
		script, err := search.GetPackageScriptWithManager(pkgName, a.packageManager())
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
	cmd, err := openCommand(path)
	if err != nil {
		a.detailView.SetText(fmt.Sprintf("[yellow]Selected path[white]\n%s\n\n[yellow]%v", tview.Escape(path), err))
		return
	}

	a.detailView.SetText(fmt.Sprintf("[blue]Opening[white]\n%s\n\n[gray]Running: %s", tview.Escape(path), tview.Escape(strings.Join(cmd.Args, " "))))
	go func() {
		err := cmd.Start()
		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.detailView.SetText(fmt.Sprintf("[red]Open failed:[white] %v\n\n%s", err, tview.Escape(path)))
				return
			}
			a.detailView.SetText(fmt.Sprintf("[green]Opened:[white] %s", tview.Escape(path)))
		})
	}()
}

func (a *App) runCommandInDetail(cmd *exec.Cmd, title, successMessage string) {
	header := fmt.Sprintf("[blue]%s[white]\n[gray]Running: %s\n\n", tview.Escape(title), tview.Escape(strings.Join(cmd.Args, " ")))
	a.detailView.SetText(header)
	a.app.SetFocus(a.detailView)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.detailView.SetText(fmt.Sprintf("%s[red]Error:[white] %v", header, err))
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.detailView.SetText(fmt.Sprintf("%s[red]Error:[white] %v", header, err))
		return
	}

	if err := cmd.Start(); err != nil {
		a.detailView.SetText(fmt.Sprintf("%s[red]Error:[white] %v", header, err))
		return
	}

	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		var output strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line)
			output.WriteString("\n")
			snapshot := output.String()
			a.app.QueueUpdateDraw(func() {
				a.detailView.SetText(header + tview.Escape(snapshot))
				a.detailView.ScrollToEnd()
			})
		}

		err := cmd.Wait()
		a.app.QueueUpdateDraw(func() {
			body := tview.Escape(output.String())
			if err != nil {
				a.detailView.SetText(fmt.Sprintf("%s%s\n[red]Command failed:[white] %v", header, body, err))
				return
			}
			a.detailView.SetText(fmt.Sprintf("%s%s\n[green]%s", header, body, tview.Escape(successMessage)))
		})
	}()
}

func previewFile(path, previewCommand string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command := strings.ReplaceAll(previewCommand, "{}", shellQuote(path))
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("preview timed out")
	}
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func fileKind(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	if info.Mode().IsRegular() {
		return "file"
	}
	return info.Mode().Type().String()
}

func formatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func openCommand(path string) (*exec.Cmd, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", absPath), nil
	case "windows":
		return exec.Command("cmd", "/c", "start", "", absPath), nil
	default:
		if util.CommandExists("xdg-open") {
			return exec.Command("xdg-open", absPath), nil
		}
		return nil, fmt.Errorf("xdg-open was not found")
	}
}

// Start runs the application
func (a *App) Start() error {
	a.setMode("package")
	return a.app.SetRoot(a.flexRoot, true).EnableMouse(true).Run()
}
