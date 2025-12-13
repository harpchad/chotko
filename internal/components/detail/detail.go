// Package detail provides the detail pane component for displaying problem information.
package detail

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// Model represents the detail pane component.
type Model struct {
	styles  *theme.Styles
	problem *zabbix.Problem
	width   int
	height  int
	focused bool
	scroll  int
}

// New creates a new detail pane model.
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

// SetProblem sets the problem to display.
func (m *Model) SetProblem(p *zabbix.Problem) {
	m.problem = p
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

		// Apply scroll and render
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
	}

	content := b.String()
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
