package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/harpchad/chotko/internal/theme"
)

// RunWizard runs the interactive setup wizard.
func RunWizard() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           Welcome to Chotko Setup Wizard                 ║")
	fmt.Println("║                                                          ║")
	fmt.Println("║  Let's configure your Zabbix connection.                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	cfg := DefaultConfig()

	// Server URL
	fmt.Print("Zabbix Server URL (e.g., https://zabbix.example.com): ")
	url, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read server URL: %w", err)
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")
	cfg.Server.URL = url

	// Authentication method
	fmt.Println()
	fmt.Println("Authentication method:")
	fmt.Println("  1. API Token (recommended for Zabbix 5.4+)")
	fmt.Println("  2. Username/Password")
	fmt.Print("Select [1/2]: ")
	authChoice, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read auth choice: %w", err)
	}
	authChoice = strings.TrimSpace(authChoice)

	if authChoice == "1" || authChoice == "" {
		fmt.Println()
		fmt.Println("To create an API token:")
		fmt.Println("  1. Log in to Zabbix web interface")
		fmt.Println("  2. Go to User settings > API tokens")
		fmt.Println("  3. Create a new token")
		fmt.Println()
		fmt.Print("API Token: ")
		token, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read API token: %w", err)
		}
		token = strings.TrimSpace(token)
		if token == "" {
			return nil, fmt.Errorf("API token is required")
		}
		cfg.Auth.Token = token
	} else {
		fmt.Println()
		fmt.Print("Username: ")
		username, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read username: %w", err)
		}
		username = strings.TrimSpace(username)
		if username == "" {
			return nil, fmt.Errorf("username is required")
		}
		cfg.Auth.Username = username

		fmt.Print("Password: ")
		password, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(password)
		if password == "" {
			return nil, fmt.Errorf("password is required")
		}
		cfg.Auth.Password = password
	}

	// Theme selection
	fmt.Println()
	fmt.Println("Available themes:")
	themes := theme.BuiltinThemeNames()
	for i, t := range themes {
		fmt.Printf("  %d. %s\n", i+1, t)
	}
	fmt.Print("Select theme [1-7, default=2 (nord)]: ")
	themeChoice, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read theme choice: %w", err)
	}
	themeChoice = strings.TrimSpace(themeChoice)

	cfg.Display.Theme = "nord" // default
	if themeChoice != "" {
		var idx int
		if _, parseErr := fmt.Sscanf(themeChoice, "%d", &idx); parseErr == nil {
			if idx >= 1 && idx <= len(themes) {
				cfg.Display.Theme = themes[idx-1]
			}
		}
	}

	// Refresh interval
	fmt.Println()
	fmt.Print("Refresh interval in seconds [30]: ")
	refreshStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh interval: %w", err)
	}
	refreshStr = strings.TrimSpace(refreshStr)
	if refreshStr != "" {
		var refresh int
		if _, parseErr := fmt.Sscanf(refreshStr, "%d", &refresh); parseErr == nil {
			if refresh >= MinRefreshInterval {
				cfg.Display.RefreshInterval = refresh
			}
		}
	}

	// Save configuration
	fmt.Println()
	fmt.Printf("Saving configuration to: %s\n", Path())

	if err := Save(cfg); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	// Save theme template
	if err := theme.SaveThemeTemplate(Dir()); err != nil {
		fmt.Printf("Warning: Could not save theme template: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved successfully!")
	fmt.Println()
	fmt.Println("You can edit the configuration file at any time:")
	fmt.Printf("  %s\n", Path())
	fmt.Println()
	fmt.Println("Custom themes can be added to:")
	fmt.Printf("  %s/themes/\n", Dir())
	fmt.Println()

	return cfg, nil
}

// PromptForConfig prompts the user to run the wizard or exit.
func PromptForConfig() bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("No configuration file found.")
	fmt.Println()
	fmt.Print("Would you like to run the setup wizard? [Y/n]: ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "" || response == "y" || response == "yes"
}
