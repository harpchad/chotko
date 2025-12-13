package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorPalette_SeverityColor(t *testing.T) {
	t.Parallel()

	palette := ColorPalette{
		Disaster:      lipgloss.Color("#FF0000"),
		High:          lipgloss.Color("#FF6600"),
		Average:       lipgloss.Color("#FFAA00"),
		Warning:       lipgloss.Color("#FFCC00"),
		Information:   lipgloss.Color("#6699FF"),
		NotClassified: lipgloss.Color("#999999"),
	}

	tests := []struct {
		name     string
		severity int
		want     lipgloss.TerminalColor
	}{
		{"severity 5 returns Disaster", 5, palette.Disaster},
		{"severity 4 returns High", 4, palette.High},
		{"severity 3 returns Average", 3, palette.Average},
		{"severity 2 returns Warning", 2, palette.Warning},
		{"severity 1 returns Information", 1, palette.Information},
		{"severity 0 returns NotClassified", 0, palette.NotClassified},
		{"negative severity returns NotClassified", -1, palette.NotClassified},
		{"severity 6 returns NotClassified", 6, palette.NotClassified},
		{"severity 100 returns NotClassified", 100, palette.NotClassified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := palette.SeverityColor(tt.severity)
			if got != tt.want {
				t.Errorf("SeverityColor(%d) = %v, want %v", tt.severity, got, tt.want)
			}
		})
	}
}

func TestSeverityName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity int
		want     string
	}{
		{"severity 5", 5, "Disaster"},
		{"severity 4", 4, "High"},
		{"severity 3", 3, "Average"},
		{"severity 2", 2, "Warning"},
		{"severity 1", 1, "Information"},
		{"severity 0", 0, "Not classified"},
		{"negative severity", -1, "Not classified"},
		{"severity 6", 6, "Not classified"},
		{"severity 100", 100, "Not classified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SeverityName(tt.severity)
			if got != tt.want {
				t.Errorf("SeverityName(%d) = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

func TestColorPalette_AllFieldsUsed(t *testing.T) {
	t.Parallel()

	// Ensure all ColorPalette fields are non-nil when properly initialized
	palette := DefaultTheme().Colors

	// Severity colors
	if palette.Disaster == nil {
		t.Error("Disaster color should not be nil")
	}
	if palette.High == nil {
		t.Error("High color should not be nil")
	}
	if palette.Average == nil {
		t.Error("Average color should not be nil")
	}
	if palette.Warning == nil {
		t.Error("Warning color should not be nil")
	}
	if palette.Information == nil {
		t.Error("Information color should not be nil")
	}
	if palette.NotClassified == nil {
		t.Error("NotClassified color should not be nil")
	}

	// Status colors
	if palette.OK == nil {
		t.Error("OK color should not be nil")
	}
	if palette.Unknown == nil {
		t.Error("Unknown color should not be nil")
	}
	if palette.Maintenance == nil {
		t.Error("Maintenance color should not be nil")
	}

	// UI colors
	if palette.Primary == nil {
		t.Error("Primary color should not be nil")
	}
	if palette.Secondary == nil {
		t.Error("Secondary color should not be nil")
	}
	if palette.Background == nil {
		t.Error("Background color should not be nil")
	}
	if palette.Foreground == nil {
		t.Error("Foreground color should not be nil")
	}
	if palette.Muted == nil {
		t.Error("Muted color should not be nil")
	}
	if palette.Border == nil {
		t.Error("Border color should not be nil")
	}
	if palette.FocusedBorder == nil {
		t.Error("FocusedBorder color should not be nil")
	}
	if palette.Highlight == nil {
		t.Error("Highlight color should not be nil")
	}
	if palette.Surface == nil {
		t.Error("Surface color should not be nil")
	}
}
