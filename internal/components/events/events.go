// Package events provides the events list component displaying Zabbix event history.
package events

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

// Model represents the events list component.
type Model struct {
	*listview.State
	styles   *theme.Styles
	events   []zabbix.Event
	filtered []zabbix.Event

	// Filter state
	textFilter string
}

// New creates a new events list model.
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

// SetEvents updates the events list.
func (m *Model) SetEvents(events []zabbix.Event) {
	m.events = events
	m.applyFilter()
}

// SetTextFilter sets the text filter.
func (m *Model) SetTextFilter(filter string) {
	m.textFilter = strings.ToLower(filter)
	m.applyFilter()
}

// applyFilter filters events based on current filter settings.
func (m *Model) applyFilter() {
	m.filtered = nil
	for _, e := range m.events {
		if m.textFilter != "" {
			name := strings.ToLower(e.Name)
			host := strings.ToLower(e.HostName())
			if !strings.Contains(name, m.textFilter) && !strings.Contains(host, m.textFilter) {
				continue
			}
		}
		m.filtered = append(m.filtered, e)
	}

	// Update count in State (handles cursor bounds automatically)
	m.SetCount(len(m.filtered))
}

// Selected returns the currently selected event.
// Returns a pointer to the element in the filtered slice. The pointer remains
// valid until the next call to SetEvents or filter changes. Callers should
// not store this pointer long-term.
func (m Model) Selected() *zabbix.Event {
	cursor := m.Cursor()
	if cursor >= 0 && cursor < len(m.filtered) {
		return &m.filtered[cursor]
	}
	return nil
}

// SelectedIndex returns the index of the selected event in the original list.
func (m Model) SelectedIndex() int {
	if selected := m.Selected(); selected != nil {
		for i, e := range m.events {
			if e.EventID == selected.EventID {
				return i
			}
		}
	}
	return -1
}

// ItemCount returns the total and filtered event counts.
func (m Model) ItemCount() (total, filtered int) {
	return len(m.events), len(m.filtered)
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
	header := fmt.Sprintf("EVENTS (%d", filtered)
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
		e := m.filtered[i]
		row := m.renderRow(e, i == cursor)
		// Mark row with zone for mouse click detection
		rowID := fmt.Sprintf("event_%d", i)
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

// renderRow renders a single event row.
func (m Model) renderRow(e zabbix.Event, selected bool) string {
	width := m.Width()

	// Status indicator - recovery (OK) or problem
	var indicator string
	var statusStyle lipgloss.Style

	if e.IsRecovery() {
		indicator = "OK"
		statusStyle = m.styles.StatusOK
	} else {
		// Use severity colors for problems
		severity := e.SeverityInt()
		indicator = "!!"
		statusStyle = m.styles.AlertSeverity[severity]
	}

	// Time
	timeStr := e.StartTime().Format("15:04:05")

	// Host name
	host := e.HostName()
	if len(host) > 12 {
		host = host[:9] + "..."
	}

	// Event name
	name := e.Name
	nameWidth := width - 12 - 12 - 8 - 8 // time, host, status, padding
	if nameWidth < 10 {
		nameWidth = 10
	}
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	// Duration (for resolved events, show how long it lasted)
	var duration string
	if e.IsRecovery() {
		duration = e.ResolvedDurationString()
	} else {
		duration = e.DurationString()
	}
	if len(duration) > 8 {
		duration = duration[:8]
	}

	if selected {
		// Build plain text row, then apply highlight style to the whole thing
		// This prevents ANSI code fragmentation from individual column styles
		timePadded := fmt.Sprintf("%-8s", timeStr)
		hostPadded := fmt.Sprintf("%-12s", host)
		namePadded := fmt.Sprintf("%-*s", nameWidth, name)
		durationPadded := fmt.Sprintf("%8s", duration)

		row := fmt.Sprintf("%s %s %s %s %s", indicator, timePadded, hostPadded, namePadded, durationPadded)
		// Pad to full width for consistent highlight
		if len(row) < width-2 {
			row += strings.Repeat(" ", width-2-len(row))
		}
		return m.styles.AlertSelected.Render(row)
	}

	// Normal row rendering
	statusIcon := statusStyle.Render(indicator)
	timeStrStyled := m.styles.Subtle.Width(8).Render(timeStr)
	hostStr := m.styles.AlertHost.Width(12).Render(host)
	nameStr := m.styles.AlertName.Width(nameWidth).Render(name)
	durationStr := m.styles.AlertDuration.Width(8).Align(lipgloss.Right).Render(duration)

	row := fmt.Sprintf("%s %s %s %s %s", statusIcon, timeStrStyled, hostStr, nameStr, durationStr)
	return m.styles.AlertNormal.Width(width - 2).Render(row)
}
