# ezysearch

A universal package manager and search tool written in Go.

ezysearch is a terminal-based application that provides a unified interface for searching and installing packages across different operating systems, searching GitHub repositories, and finding files in your filesystem.

## Features

- **Cross-platform package search**: Works with pacman/yay (Arch), apt (Debian/Ubuntu), brew (macOS), and dnf (Fedora)
- **GitHub repository search**: Search and clone repositories directly from the terminal using `gh`
- **File/directory search**: Find files in your filesystem with preview and open actions
- **Interactive TUI**: User-friendly terminal interface built with tview
- **Fuzzy search**: Quickly find what you're looking for

## Installation

### Go install

```bash
go install github.com/tumillanino/ezysearch@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`, then run `ezysearch`.

### From source

```bash
# Clone the repository
git clone https://github.com/tumillanino/ezysearch.git
cd ezysearch

# Build
make build

# Install
sudo make install
```

### Requirements

- Go 1.21 or later
- Supported package manager (pacman/yay, apt, brew, or dnf)
- Optional: GitHub CLI for GitHub search functionality
- Optional: fd and bat for enhanced directory search

## Usage

Run ezysearch from your terminal:

```bash
ezysearch
```

To force a package manager instead of using auto-detection:

```bash
ezysearch --homebrew
ezysearch --dnf
ezysearch --package-manager brew
```

Supported package-manager flags are `--auto`, `--yay`, `--pacman`, `--apt`, `--brew`, `--homebrew`, `--hombrew`, `--dnf`, and `--zypper`.

### Key Bindings

- `Ctrl+P` - Package search
- `Ctrl+G` - GitHub repository search
- `Ctrl+T` - Directory/file search
- `Enter` - Execute search or select item
- `Esc` - Return to previous view
- `Ctrl+C` - Quit

## Configuration

ezysearch creates a TOML configuration file at `~/.config/ezysearch/config.toml` with the following options:

```toml
package_search_key = "Ctrl+P"
github_search_key = "Ctrl+G"
directory_search_key = "Ctrl+T"
github_limit = 50
directory_command = "fd --hidden --strip-cwd-prefix --exclude .git"
preview_command = "bat --color=always -n --line-range :500 {}"
cache_expiry = 60

[package_manager]
sudo = "sudo"
confirm_install = true
pacman_flags = []
apt_flags = []
dnf_flags = []
zypper_flags = []
brew_flags = []
yay_flags = []

[ui]
color_scheme = "default"
show_package_count = true
```

## License

MIT
