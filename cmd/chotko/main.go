// Package main is the entry point for the chotko Zabbix TUI application.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	flag "github.com/spf13/pflag"

	"github.com/harpchad/chotko/internal/app"
	"github.com/harpchad/chotko/internal/config"
	"github.com/harpchad/chotko/internal/theme"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Command line flags
	var (
		configPath  string
		serverURL   string
		apiToken    string
		username    string
		password    string
		themeName   string
		refresh     int
		minSeverity int
		showVersion bool
		showHelp    bool
	)

	flag.StringVarP(&configPath, "config", "c", "", "Path to config file")
	flag.StringVarP(&serverURL, "server", "s", "", "Zabbix server URL")
	flag.StringVarP(&apiToken, "token", "t", "", "API token")
	flag.StringVarP(&username, "user", "u", "", "Username")
	flag.StringVarP(&password, "password", "p", "", "Password")
	flag.StringVar(&themeName, "theme", "", "Theme name")
	flag.IntVarP(&refresh, "refresh", "r", 0, "Refresh interval in seconds")
	flag.IntVar(&minSeverity, "min-severity", -1, "Minimum severity (0-5)")
	flag.BoolVarP(&showVersion, "version", "v", false, "Show version")
	flag.BoolVarP(&showHelp, "help", "h", false, "Show help")

	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("chotko %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Load or create configuration
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Apply command line overrides
	if serverURL != "" {
		cfg.Server.URL = serverURL
	}
	if apiToken != "" {
		cfg.Auth.Token = apiToken
		cfg.Auth.Username = ""
		cfg.Auth.Password = ""
	}
	if username != "" {
		cfg.Auth.Username = username
		cfg.Auth.Token = ""
	}
	if password != "" {
		cfg.Auth.Password = password
	}
	if themeName != "" {
		cfg.Display.Theme = themeName
	}
	if refresh > 0 {
		cfg.Display.RefreshInterval = refresh
	}
	if minSeverity >= 0 {
		cfg.Display.MinSeverity = minSeverity
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Load theme
	t, err := theme.Load(cfg.Display.Theme, config.Dir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load theme '%s', using default: %v\n", cfg.Display.Theme, err)
		t = theme.DefaultTheme()
	}

	// Initialize mouse zone manager for click detection
	zone.NewGlobal()

	// Create and run the application
	model := app.New(cfg, t)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// loadConfig loads configuration from file or runs the setup wizard.
func loadConfig(path string) (*config.Config, error) {
	// Use specified path or default
	if path == "" {
		path = config.Path()
	}

	// Try to load existing config
	cfg, err := config.LoadFromFile(path)
	if err == nil {
		return cfg, nil
	}

	// Config doesn't exist - prompt for wizard
	if !config.PromptForConfig() {
		return nil, fmt.Errorf("configuration required. Run with --help for options")
	}

	// Run setup wizard
	cfg, err = config.RunWizard()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func printUsage() {
	fmt.Println(`Chotko - Zabbix Terminal UI

Usage:
  chotko [flags]

Flags:
  -c, --config string     Path to config file (default ~/.config/chotko/config.yaml)
  -s, --server string     Zabbix server URL (overrides config)
  -t, --token string      API token (overrides config)
  -u, --user string       Username for auth (overrides config)
  -p, --password string   Password for auth (overrides config)
      --theme string      Theme name (default "nord")
  -r, --refresh int       Refresh interval in seconds (default 30)
      --min-severity int  Minimum severity to display (0-5)
  -h, --help              Show this help
  -v, --version           Show version

Examples:
  # Run with config file (or setup wizard if none exists)
  chotko

  # Connect with API token
  chotko -s https://zabbix.example.com -t YOUR_API_TOKEN

  # Connect with username/password
  chotko -s https://zabbix.example.com -u Admin -p password

  # Use specific theme
  chotko --theme dracula

  # Show only high severity alerts
  chotko --min-severity 4

Available Themes:
  default, nord, dracula, gruvbox, catppuccin, tokyonight, solarized

Key Bindings (press ? in app for full list):
  j/k, ↑/↓    Navigate alerts
  Tab         Switch panes
  a           Acknowledge alert
  r           Refresh
  /           Filter
  :           Command mode
  ?           Help
  q           Quit

Configuration:
  Config file: ~/.config/chotko/config.yaml
  Custom themes: ~/.config/chotko/themes/`)
}
