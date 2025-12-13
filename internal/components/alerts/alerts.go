// Package alerts provides the alerts list component displaying Zabbix problems.
package alerts

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// Model represents the alerts list component.
type Model struct {
	styles   *theme.Styles
	problems []zabbix.Problem
	filtered []zabbix.Problem
	cursor   int
	offset   int
	width    int
	height   int
	focused  bool

	// Filter state
	minSeverity int
	textFilter  string
}

// New creates a new alerts list model.
func New(styles *theme.Styles) Model {
	return Model{
		styles: styles,
	}
}

// SetSize sets the component dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetFocused sets the focus state.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// SetProblems updates the problems list.
func (m *Model) SetProblems(problems []zabbix.Problem) {
	m.problems = problems
	m.applyFilter()
}

// SetMinSeverity sets the minimum severity filter.
func (m *Model) SetMinSeverity(severity int) {
	m.minSeverity = severity
	m.applyFilter()
}

// SetTextFilter sets the text filter.
func (m *Model) SetTextFilter(filter string) {
	m.textFilter = strings.ToLower(filter)
	m.applyFilter()
}

// applyFilter filters problems based on current filter settings.
func (m *Model) applyFilter() {
	m.filtered = nil
	for _, p := range m.problems {
		if p.SeverityInt() < m.minSeverity {
			continue
		}
		if m.textFilter != "" {
			name := strings.ToLower(p.Name)
			host := strings.ToLower(p.HostName())
			if !strings.Contains(name, m.textFilter) && !strings.Contains(host, m.textFilter) {
				continue
			}
		}
		m.filtered = append(m.filtered, p)
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m.ensureVisible()
}

// Selected returns the currently selected problem.
func (m Model) Selected() *zabbix.Problem {
	if m.cursor >= 0 && m.cursor < len(m.filtered) {
		return &m.filtered[m.cursor]
	}
	return nil
}

// SelectedIndex returns the index of the selected problem in the original list.
func (m Model) SelectedIndex() int {
	if selected := m.Selected(); selected != nil {
		for i, p := range m.problems {
			if p.EventID == selected.EventID {
				return i
			}
		}
	}
	return -1
}

// Count returns the total and filtered problem counts.
func (m Model) Count() (total, filtered int) {
	return len(m.problems), len(m.filtered)
}

// MoveUp moves the cursor up.
func (m *Model) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
		m.ensureVisible()
	}
}

// MoveDown moves the cursor down.
func (m *Model) MoveDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
		m.ensureVisible()
	}
}

// PageUp moves the cursor up by one page.
func (m *Model) PageUp() {
	m.cursor -= m.visibleRows()
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

// PageDown moves the cursor down by one page.
func (m *Model) PageDown() {
	m.cursor += m.visibleRows()
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

// GoToTop moves the cursor to the first item.
func (m *Model) GoToTop() {
	m.cursor = 0
	m.offset = 0
}

// GoToBottom moves the cursor to the last item.
func (m *Model) GoToBottom() {
	m.cursor = max(0, len(m.filtered)-1)
	m.ensureVisible()
}

// visibleRows returns the number of visible rows.
func (m Model) visibleRows() int {
	return m.height - 2 // Account for header and border
}

// ensureVisible ensures the cursor is visible in the viewport.
func (m *Model) ensureVisible() {
	visible := m.visibleRows()
	if visible <= 0 {
		return
	}

	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.MoveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.MoveDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup", "ctrl+u"))):
			m.PageUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown", "ctrl+d"))):
			m.PageDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys("home", "g"))):
			m.GoToTop()
		case key.Matches(msg, key.NewBinding(key.WithKeys("end", "G"))):
			m.GoToBottom()
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	// Handle zero-size case
	if m.width < 10 || m.height < 5 {
		return ""
	}

	var b strings.Builder

	// Header
	total, filtered := m.Count()
	header := fmt.Sprintf("ALERTS (%d", filtered)
	if total != filtered {
		header += fmt.Sprintf("/%d", total)
	}
	header += ")"
	b.WriteString(m.styles.PaneTitle.Render(header))
	b.WriteString("\n")

	// Calculate visible range
	visible := m.visibleRows()
	if visible < 1 {
		visible = 1
	}

	endIdx := min(m.offset+visible, len(m.filtered))

	// Render rows
	for i := m.offset; i < endIdx; i++ {
		p := m.filtered[i]
		row := m.renderRow(p, i == m.cursor)
		b.WriteString(row)
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Pad remaining space
	rendered := endIdx - m.offset
	for i := rendered; i < visible; i++ {
		b.WriteString("\n")
	}

	// Apply pane style
	content := b.String()
	if m.focused {
		return m.styles.PaneFocused.Width(m.width).Height(m.height).Render(content)
	}
	return m.styles.PaneBlurred.Width(m.width).Height(m.height).Render(content)
}

// renderRow renders a single problem row.
func (m Model) renderRow(p zabbix.Problem, selected bool) string {
	// Severity indicator
	severity := p.SeverityInt()

	var indicator string
	switch severity {
	case 5:
		indicator = "●"
	case 4:
		indicator = "●"
	case 3:
		indicator = "◐"
	case 2:
		indicator = "○"
	case 1:
		indicator = "○"
	default:
		indicator = "○"
	}

	// Host name
	host := p.HostName()
	if len(host) > 15 {
		host = host[:12] + "..."
	}

	// Problem name
	name := p.Name
	nameWidth := m.width - 15 - 12 - 6 // host, duration, icon, padding
	if nameWidth < 10 {
		nameWidth = 10
	}
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	// Duration
	duration := p.DurationString()

	// Ack indicator
	var ackIndicator string
	if p.IsAcknowledged() {
		ackIndicator = "✓"
	} else {
		ackIndicator = " "
	}

	if selected {
		// For selected rows, use selected style for all elements
		// Keep severity color for the indicator only
		severityIcon := m.styles.AlertSeverity[severity].Render(indicator)
		hostStr := m.styles.AlertSelected.Width(15).Render(host)
		nameStr := m.styles.AlertSelected.Width(nameWidth).Render(name)
		durationStr := m.styles.AlertSelected.Width(10).Align(lipgloss.Right).Render(duration)
		ackStr := m.styles.AlertSelected.Render(ackIndicator)

		row := fmt.Sprintf("%s %s %s %s %s", severityIcon, hostStr, nameStr, durationStr, ackStr)
		return m.styles.AlertSelected.Width(m.width - 2).Render(row)
	}

	// Normal row rendering with individual styles
	severityIcon := m.styles.AlertSeverity[severity].Render(indicator)
	hostStr := m.styles.AlertHost.Width(15).Render(host)
	nameStr := m.styles.AlertName.Width(nameWidth).Render(name)
	durationStr := m.styles.AlertDuration.Width(10).Align(lipgloss.Right).Render(duration)
	ackStr := m.styles.AlertAcked.Render(ackIndicator)

	row := fmt.Sprintf("%s %s %s %s %s", severityIcon, hostStr, nameStr, durationStr, ackStr)
	return m.styles.AlertNormal.Width(m.width - 2).Render(row)
}
