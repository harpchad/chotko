// Package hosts provides the hosts list component displaying Zabbix hosts.
package hosts

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// Model represents the hosts list component.
type Model struct {
	styles   *theme.Styles
	hosts    []zabbix.Host
	filtered []zabbix.Host
	cursor   int
	offset   int
	width    int
	height   int
	focused  bool

	// Filter state
	textFilter string
}

// New creates a new hosts list model.
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

// SetHosts updates the hosts list.
func (m *Model) SetHosts(hosts []zabbix.Host) {
	m.hosts = hosts
	m.applyFilter()
}

// SetTextFilter sets the text filter.
func (m *Model) SetTextFilter(filter string) {
	m.textFilter = strings.ToLower(filter)
	m.applyFilter()
}

// applyFilter filters hosts based on current filter settings.
func (m *Model) applyFilter() {
	m.filtered = nil
	for _, h := range m.hosts {
		if m.textFilter != "" {
			name := strings.ToLower(h.DisplayName())
			host := strings.ToLower(h.Host)
			ip := strings.ToLower(m.getHostIP(h))
			if !strings.Contains(name, m.textFilter) &&
				!strings.Contains(host, m.textFilter) &&
				!strings.Contains(ip, m.textFilter) {
				continue
			}
		}
		m.filtered = append(m.filtered, h)
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m.ensureVisible()
}

// Selected returns the currently selected host.
func (m Model) Selected() *zabbix.Host {
	if m.cursor >= 0 && m.cursor < len(m.filtered) {
		return &m.filtered[m.cursor]
	}
	return nil
}

// SelectedIndex returns the index of the selected host in the original list.
func (m Model) SelectedIndex() int {
	if selected := m.Selected(); selected != nil {
		for i, h := range m.hosts {
			if h.HostID == selected.HostID {
				return i
			}
		}
	}
	return -1
}

// Count returns the total and filtered host counts.
func (m Model) Count() (total, filtered int) {
	return len(m.hosts), len(m.filtered)
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

// Scroll scrolls the list by delta lines (positive = down, negative = up).
func (m *Model) Scroll(delta int) {
	m.offset += delta
	if m.offset < 0 {
		m.offset = 0
	}
	maxOffset := len(m.filtered) - m.visibleRows()
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

// FilteredCount returns the number of filtered items.
func (m Model) FilteredCount() int {
	return len(m.filtered)
}

// SetCursor sets the cursor to a specific index.
func (m *Model) SetCursor(index int) {
	if index >= 0 && index < len(m.filtered) {
		m.cursor = index
		m.ensureVisible()
	}
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
	header := fmt.Sprintf("HOSTS (%d", filtered)
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
		h := m.filtered[i]
		row := m.renderRow(h, i == m.cursor)
		// Mark row with zone for mouse click detection
		rowID := fmt.Sprintf("host_%d", i)
		b.WriteString(zone.Mark(rowID, row))
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

// getHostIP returns the primary IP address of a host.
func (m Model) getHostIP(h zabbix.Host) string {
	for _, iface := range h.Interfaces {
		if iface.Main == "1" {
			if iface.IP != "" {
				return iface.IP
			}
			return iface.DNS
		}
	}
	if len(h.Interfaces) > 0 {
		if h.Interfaces[0].IP != "" {
			return h.Interfaces[0].IP
		}
		return h.Interfaces[0].DNS
	}
	return ""
}

// renderRow renders a single host row.
func (m Model) renderRow(h zabbix.Host, selected bool) string {
	// Status indicator based on availability
	var indicator string
	var statusStyle lipgloss.Style

	if h.InMaintenance() {
		indicator = "M"
		statusStyle = m.styles.StatusMaint
	} else {
		switch h.IsAvailable() {
		case 1: // Available
			indicator = "+"
			statusStyle = m.styles.StatusOK
		case 2: // Unavailable
			indicator = "!"
			statusStyle = m.styles.StatusProblem
		default: // Unknown
			indicator = "?"
			statusStyle = m.styles.StatusUnknown
		}
	}

	// Host name
	name := h.DisplayName()
	nameWidth := m.width - 20 - 18 - 6 // IP width, status width, padding
	if nameWidth < 10 {
		nameWidth = 10
	}
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	// IP address
	ip := m.getHostIP(h)
	if len(ip) > 18 {
		ip = ip[:15] + "..."
	}

	// Host groups (show first one)
	group := ""
	if len(h.Groups) > 0 {
		group = h.Groups[0].Name
		if len(group) > 15 {
			group = group[:12] + "..."
		}
	}

	if selected {
		// For selected rows, use selected style
		statusIcon := statusStyle.Render(indicator)
		nameStr := m.styles.AlertSelected.Width(nameWidth).Render(name)
		ipStr := m.styles.AlertSelected.Width(18).Render(ip)
		groupStr := m.styles.AlertSelected.Width(15).Align(lipgloss.Right).Render(group)

		row := fmt.Sprintf("%s %s %s %s", statusIcon, nameStr, ipStr, groupStr)
		return m.styles.AlertSelected.Width(m.width - 2).Render(row)
	}

	// Normal row rendering
	statusIcon := statusStyle.Render(indicator)
	nameStr := m.styles.AlertHost.Width(nameWidth).Render(name)
	ipStr := m.styles.Subtle.Width(18).Render(ip)
	groupStr := m.styles.Subtle.Width(15).Align(lipgloss.Right).Render(group)

	row := fmt.Sprintf("%s %s %s %s", statusIcon, nameStr, ipStr, groupStr)
	return m.styles.AlertNormal.Width(m.width - 2).Render(row)
}
