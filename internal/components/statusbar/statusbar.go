// Package statusbar provides the status bar component showing host counts and connection status.
package statusbar

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// Model represents the status bar component.
type Model struct {
	styles      *theme.Styles
	width       int
	counts      *zabbix.HostCounts
	connected   bool
	version     string
	loading     bool
	lastUpdate  string
	minSeverity int
	textFilter  string
}

// New creates a new status bar model.
func New(styles *theme.Styles) Model {
	return Model{
		styles: styles,
	}
}

// SetWidth sets the status bar width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetCounts updates the host counts.
func (m *Model) SetCounts(counts *zabbix.HostCounts) {
	m.counts = counts
}

// SetConnected updates the connection status.
func (m *Model) SetConnected(connected bool, version string) {
	m.connected = connected
	m.version = version
}

// SetLoading sets the loading state.
func (m *Model) SetLoading(loading bool) {
	m.loading = loading
}

// SetLastUpdate sets the last update time string.
func (m *Model) SetLastUpdate(t string) {
	m.lastUpdate = t
}

// SetFilter sets the current filter state.
func (m *Model) SetFilter(minSeverity int, textFilter string) {
	m.minSeverity = minSeverity
	m.textFilter = textFilter
}

// HasActiveFilter returns true if any filter is active.
func (m Model) HasActiveFilter() bool {
	return m.minSeverity > 0 || m.textFilter != ""
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	// Left side: host status counts
	var left string
	if m.counts != nil {
		ok := m.styles.StatusOK.Render(fmt.Sprintf("▲ %d OK", m.counts.OK))
		problem := m.styles.StatusProblem.Render(fmt.Sprintf("▼ %d Problem", m.counts.Problem))
		unknown := m.styles.StatusUnknown.Render(fmt.Sprintf("? %d Unknown", m.counts.Unknown))
		maint := m.styles.StatusMaint.Render(fmt.Sprintf("⚙ %d Maint", m.counts.Maintenance))
		left = fmt.Sprintf("Hosts: %s │ %s │ %s │ %s", ok, problem, unknown, maint)
	} else {
		left = "Hosts: Loading..."
	}

	// Center: filter indicator (if active)
	var center string
	if m.HasActiveFilter() {
		var parts []string
		if m.minSeverity > 0 {
			severityNames := []string{"", "Info+", "Warn+", "Avg+", "High+", "Disaster"}
			if m.minSeverity <= 5 {
				parts = append(parts, severityNames[m.minSeverity])
			}
		}
		if m.textFilter != "" {
			parts = append(parts, fmt.Sprintf("\"%s\"", m.textFilter))
		}
		filterText := "⚡ Filter: " + joinParts(parts, ", ")
		center = m.styles.StatusFilter.Render(filterText)
	}

	// Right side: connection status and refresh indicator
	var right string
	switch {
	case m.loading:
		right = "⟳ Refreshing..."
	case m.connected:
		right = fmt.Sprintf("✓ Zabbix %s", m.version)
		if m.lastUpdate != "" {
			right += fmt.Sprintf(" │ Updated: %s", m.lastUpdate)
		}
	default:
		right = "✗ Disconnected"
	}

	// Calculate padding
	leftLen := lipgloss.Width(left)
	centerLen := lipgloss.Width(center)
	rightLen := lipgloss.Width(right)

	// Distribute padding: some before center, some after
	totalPadding := m.width - leftLen - centerLen - rightLen - 2
	if totalPadding < 2 {
		totalPadding = 2
	}

	var content string
	if center != "" {
		leftPad := totalPadding / 2
		rightPad := totalPadding - leftPad
		content = left + fmt.Sprintf("%*s", leftPad, "") + center + fmt.Sprintf("%*s", rightPad, "") + right
	} else {
		content = left + fmt.Sprintf("%*s", totalPadding, "") + right
	}

	return m.styles.StatusBar.Width(m.width).Render(content)
}

// joinParts joins non-empty strings with a separator.
func joinParts(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if p != "" {
			if result != "" {
				result += sep
			}
			result += p
		}
		_ = i // suppress unused warning
	}
	return result
}
