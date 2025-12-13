package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// CustomThemeConfig represents a custom theme loaded from YAML.
type CustomThemeConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Colors      CustomColorConfig `yaml:"colors"`
}

// CustomColorConfig holds color hex values from a custom theme file.
type CustomColorConfig struct {
	// Severity colors
	Disaster      string `yaml:"disaster"`
	High          string `yaml:"high"`
	Average       string `yaml:"average"`
	Warning       string `yaml:"warning"`
	Information   string `yaml:"information"`
	NotClassified string `yaml:"not_classified"`

	// Status colors
	OK          string `yaml:"ok"`
	Unknown     string `yaml:"unknown"`
	Maintenance string `yaml:"maintenance"`

	// UI colors
	Primary       string `yaml:"primary"`
	Secondary     string `yaml:"secondary"`
	Background    string `yaml:"background"`
	Foreground    string `yaml:"foreground"`
	Muted         string `yaml:"muted"`
	Border        string `yaml:"border"`
	FocusedBorder string `yaml:"focused_border"`
	Highlight     string `yaml:"highlight"`
	Surface       string `yaml:"surface"`
}

// Load attempts to load a theme by name.
// It first checks built-in themes, then looks for custom theme files.
func Load(name string, configDir string) (*Theme, error) {
	// Check built-in themes first
	if theme, ok := BuiltinThemes()[name]; ok {
		return theme, nil
	}

	// Try to load custom theme from config directory
	themePath := filepath.Join(configDir, "themes", name+".yaml")
	return LoadFromFile(themePath)
}

// LoadFromFile loads a custom theme from a YAML file.
func LoadFromFile(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	var cfg CustomThemeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	return buildThemeFromConfig(&cfg), nil
}

// buildThemeFromConfig creates a Theme from a CustomThemeConfig.
// Missing colors fallback to the default theme.
func buildThemeFromConfig(cfg *CustomThemeConfig) *Theme {
	def := DefaultTheme()

	return &Theme{
		Name:        cfg.Name,
		Description: cfg.Description,
		Colors: ColorPalette{
			Disaster:      colorOrDefault(cfg.Colors.Disaster, def.Colors.Disaster),
			High:          colorOrDefault(cfg.Colors.High, def.Colors.High),
			Average:       colorOrDefault(cfg.Colors.Average, def.Colors.Average),
			Warning:       colorOrDefault(cfg.Colors.Warning, def.Colors.Warning),
			Information:   colorOrDefault(cfg.Colors.Information, def.Colors.Information),
			NotClassified: colorOrDefault(cfg.Colors.NotClassified, def.Colors.NotClassified),
			OK:            colorOrDefault(cfg.Colors.OK, def.Colors.OK),
			Unknown:       colorOrDefault(cfg.Colors.Unknown, def.Colors.Unknown),
			Maintenance:   colorOrDefault(cfg.Colors.Maintenance, def.Colors.Maintenance),
			Primary:       colorOrDefault(cfg.Colors.Primary, def.Colors.Primary),
			Secondary:     colorOrDefault(cfg.Colors.Secondary, def.Colors.Secondary),
			Background:    colorOrDefault(cfg.Colors.Background, def.Colors.Background),
			Foreground:    colorOrDefault(cfg.Colors.Foreground, def.Colors.Foreground),
			Muted:         colorOrDefault(cfg.Colors.Muted, def.Colors.Muted),
			Border:        colorOrDefault(cfg.Colors.Border, def.Colors.Border),
			FocusedBorder: colorOrDefault(cfg.Colors.FocusedBorder, def.Colors.FocusedBorder),
			Highlight:     colorOrDefault(cfg.Colors.Highlight, def.Colors.Highlight),
			Surface:       colorOrDefault(cfg.Colors.Surface, def.Colors.Surface),
		},
	}
}

// colorOrDefault returns a lipgloss.Color from a hex string,
// or the default color if the hex string is empty.
func colorOrDefault(hex string, def lipgloss.TerminalColor) lipgloss.TerminalColor {
	if hex == "" {
		return def
	}
	return lipgloss.Color(hex)
}

// SaveThemeTemplate saves a template theme file for user customization.
func SaveThemeTemplate(dir string) error {
	themesDir := filepath.Join(dir, "themes")
	if err := os.MkdirAll(themesDir, 0o755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	template := `# Custom Chotko Theme
# Copy this file and modify the colors to create your own theme.
# Colors should be specified as hex values (e.g., "#FF0000" for red).

name: "custom"
description: "My custom theme"

colors:
  # Severity colors (Zabbix severity levels)
  disaster: "#FF0000"       # Severity 5 - Critical/Disaster
  high: "#FF6600"           # Severity 4 - High
  average: "#FFAA00"        # Severity 3 - Average
  warning: "#FFCC00"        # Severity 2 - Warning
  information: "#6699FF"    # Severity 1 - Information
  not_classified: "#999999" # Severity 0 - Not classified

  # Status colors
  ok: "#00CC00"             # Healthy/resolved
  unknown: "#AAAAAA"        # Unknown status
  maintenance: "#AA66FF"    # Under maintenance

  # UI colors
  primary: "#6699FF"        # Primary accent, focused elements
  secondary: "#00CC00"      # Secondary accent
  background: "#1a1a1a"     # Main background
  foreground: "#EEEEEE"     # Main text color
  muted: "#666666"          # Subtle text, disabled elements
  border: "#444444"         # Unfocused borders
  focused_border: "#6699FF" # Focused pane borders
  highlight: "#333366"      # Selected/highlighted items
  surface: "#2a2a2a"        # Elevated surfaces (modals, etc.)
`

	templatePath := filepath.Join(themesDir, "custom.yaml.example")
	if err := os.WriteFile(templatePath, []byte(template), 0o644); err != nil {
		return fmt.Errorf("failed to write theme template: %w", err)
	}

	return nil
}
