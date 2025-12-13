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
	styles     *theme.Styles
	width      int
	counts     *zabbix.HostCounts
	connected  bool
	version    string
	loading    bool
	lastUpdate string
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
	rightLen := lipgloss.Width(right)
	padding := m.width - leftLen - rightLen - 2
	if padding < 1 {
		padding = 1
	}

	return m.styles.StatusBar.Width(m.width).Render(
		left + fmt.Sprintf("%*s", padding, "") + right,
	)
}
