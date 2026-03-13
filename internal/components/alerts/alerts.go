// Package alerts provides the alerts list component displaying Zabbix problems.
package alerts

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/harpchad/chotko/internal/components/listview"
	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// Model represents the alerts list component.
type Model struct {
	*listview.State
	styles   *theme.Styles
	problems []zabbix.Problem
	filtered []zabbix.Problem

	// Filter state
	minSeverity      int
	textFilter       string
	hideAcknowledged bool
	ignoredCount     int         // Number of alerts hidden by ignore rules
	hiddenAckCount   int         // Number of acknowledged alerts hidden by filter
	severityCounts   map[int]int // Cached counts by severity level

	// Ignore checker function - returns true if hostID+triggerID should be hidden
	isIgnored func(hostID, triggerID string) bool
}

// New creates a new alerts list model.
func New(styles *theme.Styles) Model {
	return Model{
		State:  listview.New(),
		styles: styles,
	}
}

// SetStyles updates the component's styles (for runtime theme changes).
func (m *Model) SetStyles(styles *theme.Styles) {
	m.styles = styles
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

// SetHideAcknowledged controls whether acknowledged alerts are hidden.
func (m *Model) SetHideAcknowledged(hide bool) {
	m.hideAcknowledged = hide
	m.applyFilter()
}

// SetIgnoreChecker sets the function used to determine if an alert should be hidden.
// The function takes hostID and triggerID and returns true if the alert should be ignored.
func (m *Model) SetIgnoreChecker(fn func(hostID, triggerID string) bool) {
	m.isIgnored = fn
	m.applyFilter()
}

// applyFilter filters problems based on current filter settings.
func (m *Model) applyFilter() {
	m.filtered = nil
	m.ignoredCount = 0
	m.hiddenAckCount = 0
	m.severityCounts = make(map[int]int)
	for _, p := range m.problems {
		// Check ignore list first - skip if host+trigger is ignored
		if m.isIgnored != nil {
			hostID := ""
			triggerID := ""
			if len(p.Hosts) > 0 {
				hostID = p.Hosts[0].HostID
			}
			// Object "0" means trigger-based problem
			if p.Object == "0" {
				triggerID = p.ObjectID
			}
			if triggerID == "" && p.RelatedObject.TriggerID != "" {
				triggerID = p.RelatedObject.TriggerID
			}
			if hostID != "" && triggerID != "" && m.isIgnored(hostID, triggerID) {
				m.ignoredCount++
				continue
			}
		}

		if m.hideAcknowledged && p.IsAcknowledged() {
			m.hiddenAckCount++
			continue
		}

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
		m.severityCounts[p.SeverityInt()]++
	}

	// Update count in State (handles cursor bounds automatically)
	m.SetCount(len(m.filtered))
}

// Selected returns the currently selected problem.
// Returns a pointer to the element in the filtered slice. The pointer remains
// valid until the next call to SetProblems or filter changes. Callers should
// not store this pointer long-term.
func (m Model) Selected() *zabbix.Problem {
	cursor := m.Cursor()
	if cursor >= 0 && cursor < len(m.filtered) {
		return &m.filtered[cursor]
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

// ItemCount returns the total and filtered problem counts.
// Total excludes ignored alerts (they are not counted as real alerts).
func (m Model) ItemCount() (total, filtered int) {
	return len(m.problems) - m.ignoredCount - m.hiddenAckCount, len(m.filtered)
}

// SeverityCounts returns a copy of the cached severity counts.
// The map keys are severity levels (0-5), values are counts.
func (m Model) SeverityCounts() map[int]int {
	counts := make(map[int]int, len(m.severityCounts))
	for k, v := range m.severityCounts {
		counts[k] = v
	}
	return counts
}

// FilteredCount returns the number of filtered items.
func (m Model) FilteredCount() int {
	return len(m.filtered)
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Focused() {
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
	width := m.Width()
	height := m.Height()

	// Handle zero-size case
	if width < 10 || height < 5 {
		return ""
	}

	var b strings.Builder

	// Header
	total, filtered := m.ItemCount()
	header := fmt.Sprintf("ALERTS (%d", filtered)
	if total != filtered {
		header += fmt.Sprintf("/%d", total)
	}
	header += ")"
	b.WriteString(m.styles.PaneTitle.Render(header))
	b.WriteString("\n")

	// Calculate visible range
	visible := m.VisibleRows()
	if visible < 1 {
		visible = 1
	}

	offset := m.Offset()
	cursor := m.Cursor()
	endIdx := min(offset+visible, len(m.filtered))

	// Render rows
	for i := offset; i < endIdx; i++ {
		p := m.filtered[i]
		row := m.renderRow(p, i == cursor)
		// Mark row with zone for mouse click detection
		rowID := fmt.Sprintf("alert_%d", i)
		b.WriteString(zone.Mark(rowID, row))
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Pad remaining space
	rendered := endIdx - offset
	for i := rendered; i < visible; i++ {
		b.WriteString("\n")
	}

	// Apply pane style
	content := b.String()
	if m.Focused() {
		return m.styles.PaneFocused.Width(width).Height(height).Render(content)
	}
	return m.styles.PaneBlurred.Width(width).Height(height).Render(content)
}

// renderRow renders a single problem row.
func (m Model) renderRow(p zabbix.Problem, selected bool) string {
	width := m.Width()

	// Severity indicator
	severity := p.SeverityInt()

	// Use distinct symbols for each severity level for accessibility
	// (allows differentiation without relying solely on color)
	var indicator string
	switch severity {
	case 5:
		indicator = "⬤" // Disaster - large filled circle
	case 4:
		indicator = "▲" // High - triangle
	case 3:
		indicator = "◆" // Average - diamond
	case 2:
		indicator = "■" // Warning - square
	case 1:
		indicator = "●" // Information - small filled circle
	default:
		indicator = "○" // Not classified - empty circle
	}

	// Host name
	host := p.HostName()
	if len(host) > 15 {
		host = host[:12] + "..."
	}

	// Problem name
	name := p.Name
	nameWidth := width - 15 - 12 - 6 // host, duration, icon, padding
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
		// Build plain text row, then apply highlight style to the whole thing
		// This prevents ANSI code fragmentation from individual column styles
		hostPadded := fmt.Sprintf("%-15s", host)
		namePadded := fmt.Sprintf("%-*s", nameWidth, name)
		durationPadded := fmt.Sprintf("%10s", duration)

		row := fmt.Sprintf("%s %s %s %s %s", indicator, hostPadded, namePadded, durationPadded, ackIndicator)
		// Pad to full width for consistent highlight
		if len(row) < width-2 {
			row += strings.Repeat(" ", width-2-len(row))
		}
		return m.styles.AlertSelected.Render(row)
	}

	// Normal row rendering with individual styles
	severityIcon := m.styles.AlertSeverity[severity].Render(indicator)
	hostStr := m.styles.AlertHost.Width(15).Render(host)
	nameStr := m.styles.AlertName.Width(nameWidth).Render(name)
	durationStr := m.styles.AlertDuration.Width(10).Align(lipgloss.Right).Render(duration)
	ackStr := m.styles.AlertAcked.Render(ackIndicator)

	row := fmt.Sprintf("%s %s %s %s %s", severityIcon, hostStr, nameStr, durationStr, ackStr)
	return m.styles.AlertNormal.Width(width - 2).Render(row)
}
