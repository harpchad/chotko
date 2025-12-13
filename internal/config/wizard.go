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
	url, _ := reader.ReadString('\n')
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
	authChoice, _ := reader.ReadString('\n')
	authChoice = strings.TrimSpace(authChoice)

	if authChoice == "1" || authChoice == "" {
		fmt.Println()
		fmt.Println("To create an API token:")
		fmt.Println("  1. Log in to Zabbix web interface")
		fmt.Println("  2. Go to User settings > API tokens")
		fmt.Println("  3. Create a new token")
		fmt.Println()
		fmt.Print("API Token: ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)
		if token == "" {
			return nil, fmt.Errorf("API token is required")
		}
		cfg.Auth.Token = token
	} else {
		fmt.Println()
		fmt.Print("Username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)
		if username == "" {
			return nil, fmt.Errorf("username is required")
		}
		cfg.Auth.Username = username

		fmt.Print("Password: ")
		password, _ := reader.ReadString('\n')
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
	themeChoice, _ := reader.ReadString('\n')
	themeChoice = strings.TrimSpace(themeChoice)

	if themeChoice == "" {
		cfg.Display.Theme = "nord"
	} else {
		var idx int
		_, _ = fmt.Sscanf(themeChoice, "%d", &idx)
		if idx >= 1 && idx <= len(themes) {
			cfg.Display.Theme = themes[idx-1]
		} else {
			cfg.Display.Theme = "nord"
		}
	}

	// Refresh interval
	fmt.Println()
	fmt.Print("Refresh interval in seconds [30]: ")
	refreshStr, _ := reader.ReadString('\n')
	refreshStr = strings.TrimSpace(refreshStr)
	if refreshStr != "" {
		var refresh int
		_, _ = fmt.Sscanf(refreshStr, "%d", &refresh)
		if refresh >= 5 {
			cfg.Display.RefreshInterval = refresh
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

	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "" || response == "y" || response == "yes"
}
