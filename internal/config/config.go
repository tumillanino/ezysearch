package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Settings holds the application configuration
type Settings struct {
	// Key bindings
	PackageSearchKey   string `toml:"package_search_key"`
	GitHubSearchKey    string `toml:"github_search_key"`
	DirectorySearchKey string `toml:"directory_search_key"`

	// GitHub settings
	GitHubLimit int `toml:"github_limit"`

	// Directory search settings
	DirectoryCommand string `toml:"directory_command"`
	PreviewCommand   string `toml:"preview_command"`

	// Cache settings
	CacheExpiry int `toml:"cache_expiry"`

	// Package manager settings
	PackageManager PackageManagerConfig `toml:"package_manager"`

	// UI settings
	UI UIConfig `toml:"ui"`
}

// PackageManagerConfig holds package manager specific settings
type PackageManagerConfig struct {
	// Sudo command (sudo, doas, etc.)
	Sudo string `toml:"sudo"`
	
	// Confirm before installing packages
	ConfirmInstall bool `toml:"confirm_install"`
	
	// Additional flags for package managers
	PacmanFlags []string `toml:"pacman_flags"`
	AptFlags    []string `toml:"apt_flags"`
	DnfFlags    []string `toml:"dnf_flags"`
	ZypperFlags []string `toml:"zypper_flags"`
	BrewFlags   []string `toml:"brew_flags"`
	YayFlags    []string `toml:"yay_flags"`
}

// UIConfig holds UI specific settings
type UIConfig struct {
	// Color scheme
	ColorScheme string `toml:"color_scheme"`
	
	// Show package counts in search results
	ShowPackageCount bool `toml:"show_package_count"`
}

// Default returns the default configuration
func Default() *Settings {
	return &Settings{
		PackageSearchKey:   "Ctrl+P",
		GitHubSearchKey:    "Ctrl+G",
		DirectorySearchKey: "Ctrl+T",
		GitHubLimit:        50,
		DirectoryCommand:   "fd --hidden --strip-cwd-prefix --exclude .git",
		PreviewCommand:     "bat --color=always -n --line-range :500 {}",
		CacheExpiry:        60, // minutes
		PackageManager: PackageManagerConfig{
			Sudo:           "sudo",
			ConfirmInstall: true,
			PacmanFlags:    []string{},
			AptFlags:       []string{},
			DnfFlags:       []string{},
			ZypperFlags:    []string{},
			BrewFlags:      []string{},
			YayFlags:       []string{},
		},
		UI: UIConfig{
			ColorScheme:      "default",
			ShowPackageCount: true,
		},
	}
}

// ConfigPath returns the path to the configuration file
func ConfigPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "ezysearch", "config.toml")
}

// Load loads the configuration from file
func Load() (*Settings, error) {
	path := ConfigPath()
	
	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config
		conf := Default()
		if err := conf.Save(); err != nil {
			return nil, err
		}
		return conf, nil
	}

	// Load existing config
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf Settings
	if _, err := toml.Decode(string(data), &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

// Save saves the configuration to file
func (s *Settings) Save() error {
	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(s)
}