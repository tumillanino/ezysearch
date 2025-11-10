package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings holds the application configuration
type Settings struct {
	// Key bindings
	PackageSearchKey   string `json:"package_search_key"`
	GitHubSearchKey    string `json:"github_search_key"`
	DirectorySearchKey string `json:"directory_search_key"`

	// GitHub settings
	GitHubLimit int `json:"github_limit"`

	// Directory search settings
	DirectoryCommand string `json:"directory_command"`
	PreviewCommand   string `json:"preview_command"`

	// Cache settings
	CacheExpiry int `json:"cache_expiry"`
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
	}
}

// ConfigPath returns the path to the configuration file
func ConfigPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "ezysearch", "config.json")
}

// Load loads the configuration from file
func Load() (*Settings, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf Settings
	if err := json.Unmarshal(data, &conf); err != nil {
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

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}