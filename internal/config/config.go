// Package config handles application configuration loading and saving.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Auth    AuthConfig    `yaml:"auth"`
	Display DisplayConfig `yaml:"display"`
	Graphs  GraphsConfig  `yaml:"graphs,omitempty"`
}

// ServerConfig holds Zabbix server connection settings.
type ServerConfig struct {
	URL string `yaml:"url"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	Token    string `yaml:"token"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// DisplayConfig holds display/UI settings.
type DisplayConfig struct {
	RefreshInterval int    `yaml:"refresh_interval"`
	MinSeverity     int    `yaml:"min_severity"`
	Theme           string `yaml:"theme"`
}

// GraphsConfig holds settings for the graphs tab.
type GraphsConfig struct {
	// Categories are key prefixes to filter items by (e.g., "system.cpu", "vm.memory")
	Categories []string `yaml:"categories"`
	// HistoryHours is how many hours of history to display (default: 3)
	HistoryHours int `yaml:"history_hours"`
	// MaxItemsPerHost limits items per host (0 = no limit)
	MaxItemsPerHost int `yaml:"max_items_per_host"`
}

// DefaultGraphCategories returns the default item key prefixes for the graphs tab.
func DefaultGraphCategories() []string {
	return []string{
		"system.cpu",
		"system.load",
		"vm.memory",
		"vfs.fs",
		"net.if",
		"proc",
	}
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			URL: "",
		},
		Auth: AuthConfig{},
		Display: DisplayConfig{
			RefreshInterval: 30,
			MinSeverity:     0,
			Theme:           "nord",
		},
		Graphs: GraphsConfig{
			Categories:      DefaultGraphCategories(),
			HistoryHours:    3,
			MaxItemsPerHost: 50,
		},
	}
}

// Dir returns the XDG config directory for chotko.
func Dir() string {
	var configBase string

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		configBase = xdg
	} else if runtime.GOOS == "windows" {
		configBase = os.Getenv("APPDATA")
	} else {
		home, _ := os.UserHomeDir()
		configBase = filepath.Join(home, ".config")
	}

	return filepath.Join(configBase, "chotko")
}

// Path returns the full path to the config file.
func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

// Load loads configuration from the default config file.
func Load() (*Config, error) {
	return LoadFromFile(Path())
}

// LoadFromFile loads configuration from a specific file path.
func LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Exists checks if the config file exists.
func Exists() bool {
	_, err := os.Stat(Path())
	return err == nil
}

// Save writes the configuration to the default config file.
func Save(cfg *Config) error {
	return SaveToFile(cfg, Path())
}

// SaveToFile writes the configuration to a specific file path.
func SaveToFile(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := `# Chotko Configuration
# https://github.com/harpchad/chotko

`
	content := header + string(data)

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid for connecting to Zabbix.
func (c *Config) Validate() error {
	if c.Server.URL == "" {
		return fmt.Errorf("server URL is required")
	}

	if c.Auth.Token == "" && (c.Auth.Username == "" || c.Auth.Password == "") {
		return fmt.Errorf("authentication required: provide either API token or username/password")
	}

	if c.Display.RefreshInterval < 5 {
		c.Display.RefreshInterval = 5 // Minimum 5 seconds
	}

	if c.Display.MinSeverity < 0 || c.Display.MinSeverity > 5 {
		c.Display.MinSeverity = 0
	}

	return nil
}

// UseToken returns true if API token authentication should be used.
func (c *Config) UseToken() bool {
	return c.Auth.Token != ""
}

// GetGraphCategories returns the graph categories, using defaults if not configured.
func (c *Config) GetGraphCategories() []string {
	if len(c.Graphs.Categories) == 0 {
		return DefaultGraphCategories()
	}
	return c.Graphs.Categories
}

// GetHistoryHours returns the history hours, using default if not configured.
func (c *Config) GetHistoryHours() int {
	if c.Graphs.HistoryHours <= 0 {
		return 3 // default 3 hours
	}
	return c.Graphs.HistoryHours
}
