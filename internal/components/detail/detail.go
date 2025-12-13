// Package detail provides the detail pane component for displaying problem and host information.
package detail

import (
	"fmt"
	"strings"

	"github.com/NimbleMarkets/ntcharts/linechart"
	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// ViewMode represents what type of data is being displayed.
type ViewMode int

// ViewMode constants.
const (
	ViewModeProblem ViewMode = iota
	ViewModeHost
	ViewModeEvent
	ViewModeGraph
)

// Model represents the detail pane component.
type Model struct {
	styles  *theme.Styles
	mode    ViewMode
	problem *zabbix.Problem
	host    *zabbix.Host
	event   *zabbix.Event
	item    *zabbix.Item
	history []zabbix.History
	width   int
	height  int
	focused bool
	scroll  int
}

// New creates a new detail pane model.
func New(styles *theme.Styles) Model {
	return Model{
		styles: styles,
		mode:   ViewModeProblem,
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

// SetProblem sets the problem to display.
func (m *Model) SetProblem(p *zabbix.Problem) {
	m.mode = ViewModeProblem
	m.problem = p
	m.host = nil
	m.scroll = 0
}

// SetHost sets the host to display.
func (m *Model) SetHost(h *zabbix.Host) {
	m.mode = ViewModeHost
	m.host = h
	m.problem = nil
	m.event = nil
	m.scroll = 0
}

// SetEvent sets the event to display.
func (m *Model) SetEvent(e *zabbix.Event) {
	m.mode = ViewModeEvent
	m.event = e
	m.problem = nil
	m.host = nil
	m.item = nil
	m.history = nil
	m.scroll = 0
}

// SetItem sets the item to display with its history data.
func (m *Model) SetItem(i *zabbix.Item, history []zabbix.History) {
	m.mode = ViewModeGraph
	m.item = i
	m.history = history
	m.problem = nil
	m.host = nil
	m.event = nil
	m.scroll = 0
}

// Clear clears the displayed content.
func (m *Model) Clear() {
	m.problem = nil
	m.host = nil
	m.event = nil
	m.item = nil
	m.history = nil
	m.scroll = 0
}

// ScrollUp scrolls the detail view up.
func (m *Model) ScrollUp() {
	if m.scroll > 0 {
		m.scroll--
	}
}

// ScrollDown scrolls the detail view down.
func (m *Model) ScrollDown() {
	m.scroll++
}

// Scroll scrolls the detail view by delta lines (positive = down, negative = up).
func (m *Model) Scroll(delta int) {
	m.scroll += delta
	if m.scroll < 0 {
		m.scroll = 0
	}
	// Note: maxScroll isn't bounded here as it depends on dynamic content
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
	// Handle zero-size case
	if m.width < 10 || m.height < 5 {
		return ""
	}

	switch m.mode {
	case ViewModeHost:
		return m.viewHost()
	case ViewModeEvent:
		return m.viewEvent()
	case ViewModeGraph:
		return m.viewGraph()
	default:
		return m.viewProblem()
	}
}

// viewProblem renders the problem detail view.
func (m Model) viewProblem() string {
	var b strings.Builder

	// Header
	b.WriteString(m.styles.PaneTitle.Render("ALERT DETAIL"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", max(0, m.width-4)))
	b.WriteString("\n")

	if m.problem == nil {
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render("  Select an alert to view details"))
		b.WriteString("\n")
	} else {
		p := m.problem

		// Build detail lines
		lines := []string{}

		// Host
		lines = append(lines, m.renderField("Host", p.HostName()))

		// IP Address
		if ip := p.HostIP(); ip != "" {
			lines = append(lines, m.renderField("IP", ip))
		}

		// Trigger/Problem name
		lines = append(lines, m.renderField("Trigger", p.Name))

		// Severity
		sevName := theme.SeverityName(p.SeverityInt())
		sevStyle := m.styles.AlertSeverity[p.SeverityInt()]
		lines = append(lines, m.renderFieldStyled("Severity", sevName, sevStyle))

		// Duration
		lines = append(lines, m.renderField("Duration", p.DurationString()))

		// Start time
		lines = append(lines, m.renderField("Started", p.StartTime().Format("2006-01-02 15:04:05")))

		// Acknowledged
		if p.IsAcknowledged() {
			lines = append(lines, m.renderFieldStyled("Status", "Acknowledged", m.styles.AlertAcked))
		} else {
			lines = append(lines, m.renderFieldStyled("Status", "Unacknowledged", m.styles.AlertSeverity[4]))
		}

		// Suppressed
		if p.IsSuppressed() {
			lines = append(lines, m.renderField("Suppressed", "Yes"))
		}

		// Event ID
		lines = append(lines, m.renderField("Event ID", p.EventID))

		// Tags
		if len(p.Tags) > 0 {
			lines = append(lines, "")
			lines = append(lines, m.styles.DetailLabel.Render("Tags:"))
			for _, tag := range p.Tags {
				tagStr := tag.Tag
				if tag.Value != "" {
					tagStr += "=" + tag.Value
				}
				lines = append(lines, "  "+m.styles.DetailTag.Render(tagStr))
			}
		}

		// Acknowledgments
		if len(p.Acknowledges) > 0 {
			lines = append(lines, "")
			lines = append(lines, m.styles.DetailLabel.Render("History:"))
			for _, ack := range p.Acknowledges {
				user := ack.Username
				if user == "" {
					user = "system"
				}
				msg := fmt.Sprintf("  %s: %s", user, ack.Message)
				if len(msg) > m.width-6 {
					msg = msg[:m.width-9] + "..."
				}
				lines = append(lines, m.styles.Subtle.Render(msg))
			}
		}

		// Actions hint
		lines = append(lines, "")
		lines = append(lines, strings.Repeat("─", max(0, m.width-4)))
		lines = append(lines, m.styles.Subtle.Render("[a]ck [A]ck+msg [t]riggers [m]acros [r]efresh"))

		b.WriteString(m.renderLines(lines))
	}

	return m.renderPane(b.String())
}

// viewEvent renders the event detail view.
func (m Model) viewEvent() string {
	var b strings.Builder

	// Header
	b.WriteString(m.styles.PaneTitle.Render("EVENT DETAIL"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", max(0, m.width-4)))
	b.WriteString("\n")

	if m.event == nil {
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render("  Select an event to view details"))
		b.WriteString("\n")
	} else {
		e := m.event

		// Build detail lines
		lines := []string{}

		// Event type (Problem or Recovery)
		if e.IsRecovery() {
			lines = append(lines, m.renderFieldStyled("Type", "Recovery (OK)", m.styles.StatusOK))
		} else {
			lines = append(lines, m.renderFieldStyled("Type", "Problem", m.styles.StatusProblem))
		}

		// Host
		lines = append(lines, m.renderField("Host", e.HostName()))

		// IP Address
		if ip := e.HostIP(); ip != "" {
			lines = append(lines, m.renderField("IP", ip))
		}

		// Trigger/Event name
		lines = append(lines, m.renderField("Trigger", e.Name))

		// Severity
		sevName := theme.SeverityName(e.SeverityInt())
		sevStyle := m.styles.AlertSeverity[e.SeverityInt()]
		lines = append(lines, m.renderFieldStyled("Severity", sevName, sevStyle))

		// Time
		lines = append(lines, m.renderField("Time", e.StartTime().Format("2006-01-02 15:04:05")))

		// Duration / Resolved info
		if e.IsRecovery() {
			lines = append(lines, m.renderField("Resolved", e.RecoveryTime().Format("2006-01-02 15:04:05")))
			lines = append(lines, m.renderField("Duration", e.ResolvedDurationString()))
		} else {
			lines = append(lines, m.renderField("Duration", e.DurationString()))
		}

		// Acknowledged
		if e.IsAcknowledged() {
			lines = append(lines, m.renderFieldStyled("Ack", "Yes", m.styles.AlertAcked))
		} else {
			lines = append(lines, m.renderFieldStyled("Ack", "No", m.styles.Subtle))
		}

		// Event ID
		lines = append(lines, m.renderField("Event ID", e.EventID))

		// Recovery Event ID (if resolved)
		if e.REventID != "" && e.REventID != "0" {
			lines = append(lines, m.renderField("Recovery ID", e.REventID))
		}

		// Tags
		if len(e.Tags) > 0 {
			lines = append(lines, "")
			lines = append(lines, m.styles.DetailLabel.Render("Tags:"))
			for _, tag := range e.Tags {
				tagStr := tag.Tag
				if tag.Value != "" {
					tagStr += "=" + tag.Value
				}
				lines = append(lines, "  "+m.styles.DetailTag.Render(tagStr))
			}
		}

		// Acknowledgments
		if len(e.Acknowledges) > 0 {
			lines = append(lines, "")
			lines = append(lines, m.styles.DetailLabel.Render("History:"))
			for _, ack := range e.Acknowledges {
				user := ack.Username
				if user == "" {
					user = "system"
				}
				msg := fmt.Sprintf("  %s: %s", user, ack.Message)
				if len(msg) > m.width-6 {
					msg = msg[:m.width-9] + "..."
				}
				lines = append(lines, m.styles.Subtle.Render(msg))
			}
		}

		// Actions hint
		lines = append(lines, "")
		lines = append(lines, strings.Repeat("─", max(0, m.width-4)))
		lines = append(lines, m.styles.Subtle.Render("[t]riggers [m]acros [r]efresh"))

		b.WriteString(m.renderLines(lines))
	}

	return m.renderPane(b.String())
}

// viewHost renders the host detail view.
func (m Model) viewHost() string {
	var b strings.Builder

	// Header
	b.WriteString(m.styles.PaneTitle.Render("HOST DETAIL"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", max(0, m.width-4)))
	b.WriteString("\n")

	if m.host == nil {
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render("  Select a host to view details"))
		b.WriteString("\n")
	} else {
		h := m.host

		// Build detail lines
		lines := []string{}

		// Host name
		lines = append(lines, m.renderField("Name", h.DisplayName()))

		// Technical name (if different)
		if h.Name != "" && h.Name != h.Host {
			lines = append(lines, m.renderField("Host", h.Host))
		}

		// Host ID
		lines = append(lines, m.renderField("Host ID", h.HostID))

		// Status
		if h.InMaintenance() {
			lines = append(lines, m.renderFieldStyled("Status", "In Maintenance", m.styles.StatusMaint))
		} else {
			switch h.IsAvailable() {
			case 1:
				lines = append(lines, m.renderFieldStyled("Status", "Available", m.styles.StatusOK))
			case 2:
				lines = append(lines, m.renderFieldStyled("Status", "Unavailable", m.styles.StatusProblem))
			default:
				lines = append(lines, m.renderFieldStyled("Status", "Unknown", m.styles.StatusUnknown))
			}
		}

		// Monitoring status
		if h.IsMonitored() {
			lines = append(lines, m.renderFieldStyled("Monitoring", "Enabled", m.styles.StatusOK))
		} else {
			lines = append(lines, m.renderFieldStyled("Monitoring", "Disabled", m.styles.StatusUnknown))
		}

		// Interfaces
		if len(h.Interfaces) > 0 {
			lines = append(lines, "")
			lines = append(lines, m.styles.DetailLabel.Render("Interfaces:"))
			for _, iface := range h.Interfaces {
				ifaceType := m.interfaceTypeName(iface.Type)
				addr := iface.IP
				if addr == "" {
					addr = iface.DNS
				}
				if iface.Port != "" && iface.Port != "0" {
					addr += ":" + iface.Port
				}

				mainStr := ""
				if iface.Main == "1" {
					mainStr = " (default)"
				}

				availStr := ""
				switch iface.Available {
				case "1":
					availStr = m.styles.StatusOK.Render(" [OK]")
				case "2":
					availStr = m.styles.StatusProblem.Render(" [FAIL]")
				default:
					availStr = m.styles.StatusUnknown.Render(" [?]")
				}

				line := fmt.Sprintf("  %s: %s%s%s", ifaceType, addr, mainStr, availStr)
				lines = append(lines, line)
			}
		}

		// Host groups
		if len(h.Groups) > 0 {
			lines = append(lines, "")
			lines = append(lines, m.styles.DetailLabel.Render("Groups:"))
			for _, group := range h.Groups {
				lines = append(lines, "  "+m.styles.DetailTag.Render(group.Name))
			}
		}

		// Actions hint
		lines = append(lines, "")
		lines = append(lines, strings.Repeat("─", max(0, m.width-4)))
		lines = append(lines, m.styles.Subtle.Render("[t]riggers [m]acros [e]nable/disable [r]efresh"))

		b.WriteString(m.renderLines(lines))
	}

	return m.renderPane(b.String())
}

// renderLines applies scrolling and renders lines.
func (m Model) renderLines(lines []string) string {
	var b strings.Builder

	visible := m.height - 4
	if visible < 1 {
		visible = 1
	}

	start := m.scroll
	if start >= len(lines) {
		start = max(0, len(lines)-1)
	}
	end := min(start+visible, len(lines))

	for i := start; i < end; i++ {
		b.WriteString(lines[i])
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderPane applies the pane style.
func (m Model) renderPane(content string) string {
	if m.focused {
		return m.styles.PaneFocused.Width(m.width).Height(m.height).Render(content)
	}
	return m.styles.PaneBlurred.Width(m.width).Height(m.height).Render(content)
}

// renderField renders a label-value pair.
func (m Model) renderField(label, value string) string {
	return m.styles.DetailLabel.Render(label+":") + " " + m.styles.DetailValue.Render(value)
}

// renderFieldStyled renders a label-value pair with a custom style for the value.
func (m Model) renderFieldStyled(label, value string, style lipgloss.Style) string {
	return m.styles.DetailLabel.Render(label+":") + " " + style.Render(value)
}

// interfaceTypeName returns a human-readable name for interface type.
func (m Model) interfaceTypeName(t string) string {
	switch t {
	case "1":
		return "Agent"
	case "2":
		return "SNMP"
	case "3":
		return "IPMI"
	case "4":
		return "JMX"
	default:
		return "Unknown"
	}
}

// viewGraph renders the item/graph detail view.
func (m Model) viewGraph() string {
	var b strings.Builder

	// Header
	b.WriteString(m.styles.PaneTitle.Render("GRAPH DETAIL"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", max(0, m.width-4)))
	b.WriteString("\n")

	if m.item == nil {
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render("  Select an item to view graph"))
		b.WriteString("\n")
	} else {
		item := m.item

		// Build detail lines
		lines := []string{}

		// Item name
		lines = append(lines, m.renderField("Item", item.Name))

		// Host
		lines = append(lines, m.renderField("Host", item.HostName()))

		// Key
		lines = append(lines, m.renderField("Key", item.Key))

		// Current value
		value := formatItemValue(item.LastValueFloat(), item.Units)
		lines = append(lines, m.renderField("Value", value))

		// Last update
		if !item.LastTime().IsZero() {
			lines = append(lines, m.renderField("Updated", item.LastTime().Format("15:04:05")))
		}

		// Units
		if item.Units != "" {
			lines = append(lines, m.renderField("Units", item.Units))
		}

		// Item ID
		lines = append(lines, m.renderField("Item ID", item.ItemID))

		// Chart section
		if len(m.history) > 0 {
			lines = append(lines, "")
			lines = append(lines, m.styles.DetailLabel.Render("History Chart:"))
			lines = append(lines, "")

			// Create time series chart
			chartWidth := m.width - 8
			chartHeight := m.height - len(lines) - 8
			if chartHeight < 5 {
				chartHeight = 5
			}
			if chartHeight > 15 {
				chartHeight = 15
			}

			chart := tslc.New(chartWidth, chartHeight,
				tslc.WithXLabelFormatter(tslc.HourTimeLabelFormatter()),
				tslc.WithYLabelFormatter(humanReadableYLabelFormatter(item.Units)),
			)

			// Push history data points
			for _, h := range m.history {
				t := h.Time()
				if !t.IsZero() {
					chart.Push(tslc.TimePoint{Time: t, Value: h.ValueFloat()})
				}
			}

			// Draw the chart using braille characters for better resolution
			chart.DrawBraille()

			// Add chart lines
			chartLines := strings.Split(chart.View(), "\n")
			lines = append(lines, chartLines...)

			// Add time range info
			if len(m.history) > 0 {
				first := m.history[0].Time()
				last := m.history[len(m.history)-1].Time()
				timeRange := fmt.Sprintf("%s - %s", first.Format("15:04"), last.Format("15:04"))
				lines = append(lines, m.styles.Subtle.Render(timeRange))
			}
		} else {
			lines = append(lines, "")
			lines = append(lines, m.styles.Subtle.Render("  No history data available"))
		}

		// Stats section
		if len(m.history) > 0 {
			lines = append(lines, "")
			lines = append(lines, strings.Repeat("─", max(0, m.width-4)))

			// Calculate stats
			minVal, maxVal, avgVal := calcStats(m.history)
			statsLine := fmt.Sprintf("Min: %s  Max: %s  Avg: %s",
				formatItemValue(minVal, item.Units),
				formatItemValue(maxVal, item.Units),
				formatItemValue(avgVal, item.Units))
			lines = append(lines, m.styles.Subtle.Render(statsLine))
		}

		b.WriteString(m.renderLines(lines))
	}

	return m.renderPane(b.String())
}

// formatItemValue formats a numeric value with appropriate units.
func formatItemValue(value float64, units string) string {
	// Handle percentage
	if units == "%" {
		return fmt.Sprintf("%.1f%%", value)
	}

	// Handle bytes
	if units == "B" || units == "Bps" {
		return formatBytesValue(value) + strings.TrimPrefix(units, "B")
	}

	// Handle time units
	if units == "s" {
		if value < 1 {
			return fmt.Sprintf("%.0fms", value*1000)
		}
		return fmt.Sprintf("%.2fs", value)
	}

	// Default formatting
	if value >= 1000000 {
		return fmt.Sprintf("%.1fM", value/1000000)
	}
	if value >= 1000 {
		return fmt.Sprintf("%.1fK", value/1000)
	}
	if value == float64(int64(value)) {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

// formatBytesValue formats bytes to human-readable form.
func formatBytesValue(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", bytes/div, "KMGTPE"[exp])
}

// calcStats calculates minVal, maxVal, avgVal for history data.
func calcStats(history []zabbix.History) (minVal, maxVal, avgVal float64) {
	if len(history) == 0 {
		return 0, 0, 0
	}

	minVal = history[0].ValueFloat()
	maxVal = minVal
	sum := 0.0

	for _, h := range history {
		v := h.ValueFloat()
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
		sum += v
	}

	avgVal = sum / float64(len(history))
	return minVal, maxVal, avgVal
}

// humanReadableYLabelFormatter returns a LabelFormatter that formats Y axis
// values in human-readable form (e.g., 16G instead of 16000000000).
func humanReadableYLabelFormatter(units string) linechart.LabelFormatter {
	return func(_ int, v float64) string {
		return formatYValue(v, units)
	}
}

// formatYValue formats a value for the Y axis label with appropriate SI suffixes.
func formatYValue(value float64, units string) string {
	// Handle bytes specially - use binary prefixes (Ki, Mi, Gi)
	if units == "B" || units == "Bps" {
		return formatBytesShort(value)
	}

	// Handle percentages
	if units == "%" {
		return fmt.Sprintf("%.0f%%", value)
	}

	// For other values, use SI prefixes
	absVal := value
	if absVal < 0 {
		absVal = -absVal
	}

	switch {
	case absVal >= 1e12:
		return fmt.Sprintf("%.1fT", value/1e12)
	case absVal >= 1e9:
		return fmt.Sprintf("%.1fG", value/1e9)
	case absVal >= 1e6:
		return fmt.Sprintf("%.1fM", value/1e6)
	case absVal >= 1e3:
		return fmt.Sprintf("%.1fK", value/1e3)
	case absVal >= 1:
		return fmt.Sprintf("%.0f", value)
	case absVal >= 0.01:
		return fmt.Sprintf("%.2f", value)
	default:
		return fmt.Sprintf("%.0f", value)
	}
}

// formatBytesShort formats bytes to short human-readable form for axis labels.
func formatBytesShort(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", bytes/div, "KMGTP"[exp])
}
