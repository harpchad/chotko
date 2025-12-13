package theme

import "github.com/charmbracelet/lipgloss"

// ColorPalette defines all colors used throughout the application.
// Colors are organized by semantic meaning to ensure consistent theming.
type ColorPalette struct {
	// Severity colors (aligned with Zabbix severity levels)
	Disaster      lipgloss.TerminalColor // Severity 5 - Critical/Disaster
	High          lipgloss.TerminalColor // Severity 4 - High
	Average       lipgloss.TerminalColor // Severity 3 - Average
	Warning       lipgloss.TerminalColor // Severity 2 - Warning
	Information   lipgloss.TerminalColor // Severity 1 - Information
	NotClassified lipgloss.TerminalColor // Severity 0 - Not classified

	// Status colors
	OK          lipgloss.TerminalColor // Healthy/resolved state
	Unknown     lipgloss.TerminalColor // Unknown status
	Maintenance lipgloss.TerminalColor // Under maintenance

	// UI colors
	Primary       lipgloss.TerminalColor // Primary accent, focused elements
	Secondary     lipgloss.TerminalColor // Secondary accent
	Background    lipgloss.TerminalColor // Main background
	Foreground    lipgloss.TerminalColor // Main text color
	Muted         lipgloss.TerminalColor // Subtle text, disabled elements
	Border        lipgloss.TerminalColor // Unfocused borders
	FocusedBorder lipgloss.TerminalColor // Focused pane borders
	Highlight     lipgloss.TerminalColor // Selected/highlighted items
	Surface       lipgloss.TerminalColor // Elevated surfaces (modals, etc.)
}

// SeverityColor returns the appropriate color for a Zabbix severity level (0-5).
func (c ColorPalette) SeverityColor(severity int) lipgloss.TerminalColor {
	switch severity {
	case 5:
		return c.Disaster
	case 4:
		return c.High
	case 3:
		return c.Average
	case 2:
		return c.Warning
	case 1:
		return c.Information
	default:
		return c.NotClassified
	}
}

// SeverityName returns the human-readable name for a severity level.
func SeverityName(severity int) string {
	switch severity {
	case 5:
		return "Disaster"
	case 4:
		return "High"
	case 3:
		return "Average"
	case 2:
		return "Warning"
	case 1:
		return "Information"
	default:
		return "Not classified"
	}
}
