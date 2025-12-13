// Package detail provides the detail pane component for displaying problem and host information.
package detail

import (
	"fmt"
	"strings"

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
)

// Model represents the detail pane component.
type Model struct {
	styles  *theme.Styles
	mode    ViewMode
	problem *zabbix.Problem
	host    *zabbix.Host
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
	m.scroll = 0
}

// Clear clears the displayed content.
func (m *Model) Clear() {
	m.problem = nil
	m.host = nil
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
		lines = append(lines, m.styles.Subtle.Render("[a]ck [A]ck+msg [r]efresh"))

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
		lines = append(lines, m.styles.Subtle.Render("[r]efresh"))

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
