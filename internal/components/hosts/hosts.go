// Package hosts provides the hosts list component displaying Zabbix hosts.
package hosts

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

// Model represents the hosts list component.
type Model struct {
	*listview.State
	styles   *theme.Styles
	hosts    []zabbix.Host
	filtered []zabbix.Host

	// Filter state
	textFilter string
}

// New creates a new hosts list model.
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

	// Update count in State (handles cursor bounds automatically)
	m.SetCount(len(m.filtered))
}

// Selected returns the currently selected host.
// Returns a pointer to the element in the filtered slice. The pointer remains
// valid until the next call to SetHosts or filter changes. Callers should
// not store this pointer long-term.
func (m Model) Selected() *zabbix.Host {
	cursor := m.Cursor()
	if cursor >= 0 && cursor < len(m.filtered) {
		return &m.filtered[cursor]
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

// ItemCount returns the total and filtered host counts.
func (m Model) ItemCount() (total, filtered int) {
	return len(m.hosts), len(m.filtered)
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
	header := fmt.Sprintf("HOSTS (%d", filtered)
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
		h := m.filtered[i]
		row := m.renderRow(h, i == cursor)
		// Mark row with zone for mouse click detection
		rowID := fmt.Sprintf("host_%d", i)
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
	width := m.Width()

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
	nameWidth := width - 20 - 18 - 6 // IP width, status width, padding
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
		// Build plain text row, then apply highlight style to the whole thing
		// This prevents ANSI code fragmentation from individual column styles
		namePadded := fmt.Sprintf("%-*s", nameWidth, name)
		ipPadded := fmt.Sprintf("%-18s", ip)
		groupPadded := fmt.Sprintf("%15s", group)

		row := fmt.Sprintf("%s %s %s %s", indicator, namePadded, ipPadded, groupPadded)
		// Pad to full width for consistent highlight
		if len(row) < width-2 {
			row += strings.Repeat(" ", width-2-len(row))
		}
		return m.styles.AlertSelected.Render(row)
	}

	// Normal row rendering
	statusIcon := statusStyle.Render(indicator)
	nameStr := m.styles.AlertHost.Width(nameWidth).Render(name)
	ipStr := m.styles.Subtle.Width(18).Render(ip)
	groupStr := m.styles.Subtle.Width(15).Align(lipgloss.Right).Render(group)

	row := fmt.Sprintf("%s %s %s %s", statusIcon, nameStr, ipStr, groupStr)
	return m.styles.AlertNormal.Width(width - 2).Render(row)
}
