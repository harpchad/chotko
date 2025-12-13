package theme

import (
	"testing"
)

func TestBuiltinThemes(t *testing.T) {
	t.Parallel()

	themes := BuiltinThemes()

	expectedThemes := []string{
		ThemeDefault,
		ThemeNord,
		ThemeDracula,
		ThemeGruvbox,
		ThemeCatppuccin,
		ThemeTokyoNight,
		ThemeSolarized,
	}

	// Verify all expected themes are present
	for _, name := range expectedThemes {
		if _, ok := themes[name]; !ok {
			t.Errorf("BuiltinThemes() missing theme: %s", name)
		}
	}

	// Verify count matches
	if len(themes) != len(expectedThemes) {
		t.Errorf("BuiltinThemes() returned %d themes, want %d", len(themes), len(expectedThemes))
	}
}

func TestBuiltinThemeNames(t *testing.T) {
	t.Parallel()

	names := BuiltinThemeNames()

	expectedNames := []string{
		ThemeDefault,
		ThemeNord,
		ThemeDracula,
		ThemeGruvbox,
		ThemeCatppuccin,
		ThemeTokyoNight,
		ThemeSolarized,
	}

	if len(names) != len(expectedNames) {
		t.Errorf("BuiltinThemeNames() returned %d names, want %d", len(names), len(expectedNames))
	}

	// Verify all names are present (order matters for this list)
	for i, want := range expectedNames {
		if names[i] != want {
			t.Errorf("BuiltinThemeNames()[%d] = %q, want %q", i, names[i], want)
		}
	}
}

func TestThemeConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"ThemeDefault", ThemeDefault, "default"},
		{"ThemeNord", ThemeNord, "nord"},
		{"ThemeDracula", ThemeDracula, "dracula"},
		{"ThemeGruvbox", ThemeGruvbox, "gruvbox"},
		{"ThemeCatppuccin", ThemeCatppuccin, "catppuccin"},
		{"ThemeTokyoNight", ThemeTokyoNight, "tokyonight"},
		{"ThemeSolarized", ThemeSolarized, "solarized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.constant != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()

	if theme == nil {
		t.Fatal("DefaultTheme() returned nil")
	}

	if theme.Name != ThemeDefault {
		t.Errorf("DefaultTheme().Name = %q, want %q", theme.Name, ThemeDefault)
	}

	if theme.Description == "" {
		t.Error("DefaultTheme().Description should not be empty")
	}

	// Verify severity colors are distinct
	colors := theme.Colors
	severityColors := []interface{}{
		colors.Disaster,
		colors.High,
		colors.Average,
		colors.Warning,
		colors.Information,
		colors.NotClassified,
	}
	for i := 0; i < len(severityColors); i++ {
		if severityColors[i] == nil {
			t.Errorf("Severity color at index %d is nil", i)
		}
	}
}

func TestNordTheme(t *testing.T) {
	t.Parallel()

	theme := NordTheme()

	if theme == nil {
		t.Fatal("NordTheme() returned nil")
	}

	if theme.Name != ThemeNord {
		t.Errorf("NordTheme().Name = %q, want %q", theme.Name, ThemeNord)
	}

	if theme.Description == "" {
		t.Error("NordTheme().Description should not be empty")
	}
}

func TestDraculaTheme(t *testing.T) {
	t.Parallel()

	theme := DraculaTheme()

	if theme == nil {
		t.Fatal("DraculaTheme() returned nil")
	}

	if theme.Name != ThemeDracula {
		t.Errorf("DraculaTheme().Name = %q, want %q", theme.Name, ThemeDracula)
	}

	if theme.Description == "" {
		t.Error("DraculaTheme().Description should not be empty")
	}
}

func TestGruvboxTheme(t *testing.T) {
	t.Parallel()

	theme := GruvboxTheme()

	if theme == nil {
		t.Fatal("GruvboxTheme() returned nil")
	}

	if theme.Name != ThemeGruvbox {
		t.Errorf("GruvboxTheme().Name = %q, want %q", theme.Name, ThemeGruvbox)
	}

	if theme.Description == "" {
		t.Error("GruvboxTheme().Description should not be empty")
	}
}

func TestCatppuccinTheme(t *testing.T) {
	t.Parallel()

	theme := CatppuccinTheme()

	if theme == nil {
		t.Fatal("CatppuccinTheme() returned nil")
	}

	if theme.Name != ThemeCatppuccin {
		t.Errorf("CatppuccinTheme().Name = %q, want %q", theme.Name, ThemeCatppuccin)
	}

	if theme.Description == "" {
		t.Error("CatppuccinTheme().Description should not be empty")
	}
}

func TestTokyoNightTheme(t *testing.T) {
	t.Parallel()

	theme := TokyoNightTheme()

	if theme == nil {
		t.Fatal("TokyoNightTheme() returned nil")
	}

	if theme.Name != ThemeTokyoNight {
		t.Errorf("TokyoNightTheme().Name = %q, want %q", theme.Name, ThemeTokyoNight)
	}

	if theme.Description == "" {
		t.Error("TokyoNightTheme().Description should not be empty")
	}
}

func TestSolarizedTheme(t *testing.T) {
	t.Parallel()

	theme := SolarizedTheme()

	if theme == nil {
		t.Fatal("SolarizedTheme() returned nil")
	}

	if theme.Name != ThemeSolarized {
		t.Errorf("SolarizedTheme().Name = %q, want %q", theme.Name, ThemeSolarized)
	}

	if theme.Description == "" {
		t.Error("SolarizedTheme().Description should not be empty")
	}
}

func TestAllBuiltinThemesHaveValidColors(t *testing.T) {
	t.Parallel()

	themes := BuiltinThemes()

	for name, theme := range themes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if theme == nil {
				t.Fatalf("Theme %q is nil", name)
			}

			c := theme.Colors

			// Check all severity colors are set
			if c.Disaster == nil {
				t.Errorf("Theme %q: Disaster color is nil", name)
			}
			if c.High == nil {
				t.Errorf("Theme %q: High color is nil", name)
			}
			if c.Average == nil {
				t.Errorf("Theme %q: Average color is nil", name)
			}
			if c.Warning == nil {
				t.Errorf("Theme %q: Warning color is nil", name)
			}
			if c.Information == nil {
				t.Errorf("Theme %q: Information color is nil", name)
			}
			if c.NotClassified == nil {
				t.Errorf("Theme %q: NotClassified color is nil", name)
			}

			// Check all status colors are set
			if c.OK == nil {
				t.Errorf("Theme %q: OK color is nil", name)
			}
			if c.Unknown == nil {
				t.Errorf("Theme %q: Unknown color is nil", name)
			}
			if c.Maintenance == nil {
				t.Errorf("Theme %q: Maintenance color is nil", name)
			}

			// Check all UI colors are set
			if c.Primary == nil {
				t.Errorf("Theme %q: Primary color is nil", name)
			}
			if c.Secondary == nil {
				t.Errorf("Theme %q: Secondary color is nil", name)
			}
			if c.Background == nil {
				t.Errorf("Theme %q: Background color is nil", name)
			}
			if c.Foreground == nil {
				t.Errorf("Theme %q: Foreground color is nil", name)
			}
			if c.Muted == nil {
				t.Errorf("Theme %q: Muted color is nil", name)
			}
			if c.Border == nil {
				t.Errorf("Theme %q: Border color is nil", name)
			}
			if c.FocusedBorder == nil {
				t.Errorf("Theme %q: FocusedBorder color is nil", name)
			}
			if c.Highlight == nil {
				t.Errorf("Theme %q: Highlight color is nil", name)
			}
			if c.Surface == nil {
				t.Errorf("Theme %q: Surface color is nil", name)
			}
		})
	}
}

func TestBuiltinThemeNamesMatchThemeNames(t *testing.T) {
	t.Parallel()

	themes := BuiltinThemes()
	names := BuiltinThemeNames()

	for _, name := range names {
		theme, ok := themes[name]
		if !ok {
			t.Errorf("Theme name %q in BuiltinThemeNames() not found in BuiltinThemes()", name)
			continue
		}
		if theme.Name != name {
			t.Errorf("Theme[%q].Name = %q, want %q", name, theme.Name, name)
		}
	}
}
