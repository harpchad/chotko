package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_BuiltinTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		themeName string
	}{
		{"default theme", ThemeDefault},
		{"nord theme", ThemeNord},
		{"dracula theme", ThemeDracula},
		{"gruvbox theme", ThemeGruvbox},
		{"catppuccin theme", ThemeCatppuccin},
		{"tokyonight theme", ThemeTokyoNight},
		{"solarized theme", ThemeSolarized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			theme, err := Load(tt.themeName, "/nonexistent")
			if err != nil {
				t.Fatalf("Load(%q) error = %v", tt.themeName, err)
			}
			if theme == nil {
				t.Fatalf("Load(%q) returned nil theme", tt.themeName)
			}
			if theme.Name != tt.themeName {
				t.Errorf("Load(%q).Name = %q, want %q", tt.themeName, theme.Name, tt.themeName)
			}
		})
	}
}

func TestLoad_CustomTheme(t *testing.T) {
	t.Parallel()

	// Create temp directory for test
	tmpDir := t.TempDir()
	themesDir := filepath.Join(tmpDir, "themes")
	if err := os.MkdirAll(themesDir, 0o750); err != nil {
		t.Fatalf("Failed to create themes dir: %v", err)
	}

	// Create a custom theme file
	customTheme := `name: "mycustom"
description: "My custom test theme"
colors:
  disaster: "#FF0000"
  high: "#FF6600"
  ok: "#00FF00"
  primary: "#0000FF"
`
	themePath := filepath.Join(themesDir, "mycustom.yaml")
	if err := os.WriteFile(themePath, []byte(customTheme), 0o600); err != nil {
		t.Fatalf("Failed to write custom theme: %v", err)
	}

	theme, err := Load("mycustom", tmpDir)
	if err != nil {
		t.Fatalf("Load(mycustom) error = %v", err)
	}
	if theme == nil {
		t.Fatal("Load(mycustom) returned nil theme")
	}
	if theme.Name != "mycustom" {
		t.Errorf("Load(mycustom).Name = %q, want %q", theme.Name, "mycustom")
	}
	if theme.Description != "My custom test theme" {
		t.Errorf("Load(mycustom).Description = %q, want %q", theme.Description, "My custom test theme")
	}
}

func TestLoad_NonexistentTheme(t *testing.T) {
	t.Parallel()

	_, err := Load("nonexistent", t.TempDir())
	if err == nil {
		t.Error("Load(nonexistent) should return an error")
	}
}

func TestLoadFromFile_ValidFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themePath := filepath.Join(tmpDir, "test.yaml")

	content := `name: "testtheme"
description: "Test theme description"
colors:
  disaster: "#FF0000"
  high: "#FF6600"
  average: "#FFAA00"
  warning: "#FFCC00"
  information: "#6699FF"
  not_classified: "#999999"
  ok: "#00CC00"
  unknown: "#AAAAAA"
  maintenance: "#AA66FF"
  primary: "#6699FF"
  secondary: "#00CC00"
  background: "#1a1a1a"
  foreground: "#EEEEEE"
  muted: "#666666"
  border: "#444444"
  focused_border: "#6699FF"
  highlight: "#333366"
  surface: "#2a2a2a"
`
	if err := os.WriteFile(themePath, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test theme: %v", err)
	}

	theme, err := LoadFromFile(themePath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}
	if theme == nil {
		t.Fatal("LoadFromFile() returned nil theme")
	}
	if theme.Name != "testtheme" {
		t.Errorf("theme.Name = %q, want %q", theme.Name, "testtheme")
	}
	if theme.Description != "Test theme description" {
		t.Errorf("theme.Description = %q, want %q", theme.Description, "Test theme description")
	}
}

func TestLoadFromFile_NonexistentFile(t *testing.T) {
	t.Parallel()

	_, err := LoadFromFile("/nonexistent/path/theme.yaml")
	if err == nil {
		t.Error("LoadFromFile(nonexistent) should return an error")
	}
	if !strings.Contains(err.Error(), "failed to read theme file") {
		t.Errorf("Error message should mention reading file, got: %v", err)
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themePath := filepath.Join(tmpDir, "invalid.yaml")

	if err := os.WriteFile(themePath, []byte("invalid: yaml: content: ["), 0o600); err != nil {
		t.Fatalf("Failed to write invalid yaml: %v", err)
	}

	_, err := LoadFromFile(themePath)
	if err == nil {
		t.Error("LoadFromFile(invalid yaml) should return an error")
	}
	if !strings.Contains(err.Error(), "failed to parse theme file") {
		t.Errorf("Error message should mention parsing, got: %v", err)
	}
}

func TestLoadFromFile_PartialColors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themePath := filepath.Join(tmpDir, "partial.yaml")

	// Only specify a few colors - others should fallback to defaults
	content := `name: "partial"
description: "Partial theme"
colors:
  disaster: "#AA0000"
  ok: "#00AA00"
`
	if err := os.WriteFile(themePath, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write partial theme: %v", err)
	}

	theme, err := LoadFromFile(themePath)
	if err != nil {
		t.Fatalf("LoadFromFile(partial) error = %v", err)
	}

	// Check that specified colors are used
	// Note: We can't directly compare lipgloss.Color values easily,
	// so we just ensure the theme loads without nil colors

	if theme.Colors.Disaster == nil {
		t.Error("Disaster color should not be nil")
	}
	if theme.Colors.OK == nil {
		t.Error("OK color should not be nil")
	}

	// Unspecified colors should fallback to defaults (non-nil)
	if theme.Colors.High == nil {
		t.Error("High color should fallback to default, not nil")
	}
	if theme.Colors.Primary == nil {
		t.Error("Primary color should fallback to default, not nil")
	}
}

func TestSaveThemeTemplate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	err := SaveThemeTemplate(tmpDir)
	if err != nil {
		t.Fatalf("SaveThemeTemplate() error = %v", err)
	}

	// Check that the template file was created
	templatePath := filepath.Join(tmpDir, "themes", "custom.yaml.example")
	if _, statErr := os.Stat(templatePath); os.IsNotExist(statErr) {
		t.Error("Template file was not created")
	}

	// Read and verify content has expected sections
	content, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	contentStr := string(content)
	expectedStrings := []string{
		"name:",
		"description:",
		"colors:",
		"disaster:",
		"high:",
		"average:",
		"warning:",
		"information:",
		"not_classified:",
		"ok:",
		"unknown:",
		"maintenance:",
		"primary:",
		"secondary:",
		"background:",
		"foreground:",
		"muted:",
		"border:",
		"focused_border:",
		"highlight:",
		"surface:",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(contentStr, s) {
			t.Errorf("Template should contain %q", s)
		}
	}
}

func TestSaveThemeTemplate_CreatesThemesDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Don't pre-create themes dir

	err := SaveThemeTemplate(tmpDir)
	if err != nil {
		t.Fatalf("SaveThemeTemplate() error = %v", err)
	}

	themesDir := filepath.Join(tmpDir, "themes")
	info, err := os.Stat(themesDir)
	if os.IsNotExist(err) {
		t.Error("themes directory was not created")
	}
	if !info.IsDir() {
		t.Error("themes should be a directory")
	}
}

func TestSaveThemeTemplate_InvalidDir(t *testing.T) {
	t.Parallel()

	// Use a path that can't be created (file exists where dir needed)
	tmpDir := t.TempDir()
	blockingFile := filepath.Join(tmpDir, "themes")
	if err := os.WriteFile(blockingFile, []byte("blocking"), 0o600); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}

	err := SaveThemeTemplate(tmpDir)
	if err == nil {
		t.Error("SaveThemeTemplate() should fail when themes path is a file")
	}
}

func TestColorOrDefault(t *testing.T) {
	t.Parallel()

	defaultColor := DefaultTheme().Colors.Disaster

	tests := []struct {
		name       string
		hex        string
		wantCustom bool
	}{
		{"empty string uses default", "", false},
		{"hex string uses custom", "#FF0000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := colorOrDefault(tt.hex, defaultColor)
			if result == nil {
				t.Error("colorOrDefault should never return nil")
			}

			// When hex is empty, result should be the default
			// When hex is provided, result should be different from default
			// (though we can't easily verify the exact color)
			if tt.hex == "" && result != defaultColor {
				t.Error("Empty hex should return default color")
			}
		})
	}
}

func TestBuildThemeFromConfig(t *testing.T) {
	t.Parallel()

	cfg := &CustomThemeConfig{
		Name:        "testbuild",
		Description: "Test build theme",
		Colors: CustomColorConfig{
			Disaster: "#FF0000",
			OK:       "#00FF00",
			// Other fields empty - should use defaults
		},
	}

	theme := buildThemeFromConfig(cfg)

	if theme.Name != "testbuild" {
		t.Errorf("theme.Name = %q, want %q", theme.Name, "testbuild")
	}
	if theme.Description != "Test build theme" {
		t.Errorf("theme.Description = %q, want %q", theme.Description, "Test build theme")
	}

	// Verify all colors are non-nil (either from config or defaults)
	c := theme.Colors
	if c.Disaster == nil {
		t.Error("Disaster should not be nil")
	}
	if c.High == nil {
		t.Error("High should not be nil (should use default)")
	}
	if c.OK == nil {
		t.Error("OK should not be nil")
	}
	if c.Primary == nil {
		t.Error("Primary should not be nil (should use default)")
	}
}
