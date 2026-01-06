package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/harpchad/chotko/internal/theme"
)

// View renders the entire application UI.
func (m Model) View() string {
	// Show "terminal too small" message if below minimum size
	if m.tooSmall {
		msg := fmt.Sprintf("Terminal too small\nMinimum size: %dx%d\nCurrent: %dx%d",
			MinTerminalWidth, MinTerminalHeight, m.width, m.height)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.ModalText.Render(msg))
	}

	// Show loading indicator during initial connection
	if !m.connected && !m.showError && m.client == nil {
		msg := m.styles.Title.Render("Chotko") + "\n\n" +
			m.styles.ModalText.Render("⟳ Connecting to Zabbix...")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
	}

	// Show editor modal if active
	if m.showEditor {
		return m.editorPane.View()
	}

	// Show error modal if active
	if m.showError || m.showHelp {
		return m.errorModal.View()
	}

	// Render components
	statusBar := m.statusBar.View()
	tabBar := m.tabBar.View()
	commandBar := m.commandInput.View()

	// Render main content area based on active tab
	var listPane string
	switch m.tabBar.Active() {
	case TabAlerts:
		listPane = m.alertList.View()
	case TabHosts:
		listPane = m.hostList.View()
	case TabEvents:
		listPane = m.eventList.View()
	case TabGraphs:
		listPane = m.graphList.View()
	default:
		// For unimplemented tabs, show alerts as fallback
		listPane = m.alertList.View()
	}

	// When theme picker is active, show it in place of detail pane
	var detailPane string
	if m.showThemePicker {
		detailPane = m.renderThemePicker()
	} else {
		detailPane = m.detailPane.View()
	}

	// Join panes horizontally
	contentArea := lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)

	// Stack everything vertically
	mainUI := lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		tabBar,
		contentArea,
		commandBar,
	)

	return zone.Scan(mainUI)
}

// renderThemePicker renders the theme picker in place of the detail pane.
func (m Model) renderThemePicker() string {
	var content strings.Builder

	// Title
	content.WriteString(m.styles.Title.Render("Select Theme"))
	content.WriteString("\n\n")

	// Theme list with descriptions
	themes := theme.BuiltinThemes()
	for i, name := range m.themeNames {
		var line string
		desc := ""
		if t, ok := themes[name]; ok {
			desc = t.Description
		}

		if i == m.themePickerIndex {
			line = m.styles.AlertSelected.Render(fmt.Sprintf(" ▶ %-14s", name))
			if desc != "" {
				line += " " + m.styles.Subtle.Render(desc)
			}
		} else {
			line = m.styles.AlertNormal.Render(fmt.Sprintf("   %-14s", name))
			if desc != "" {
				line += " " + m.styles.Subtle.Render(desc)
			}
		}
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\n")
	content.WriteString(m.styles.Subtle.Render("↑/↓ Preview • Enter Confirm • Esc Cancel"))

	// Use the same pane style as detail pane for consistency
	// Calculate width based on detail pane size
	detailWidth := m.width - (m.width * 45 / 100) - 4 // Approximate detail pane width

	return m.styles.PaneFocused.
		Width(detailWidth).
		Height(m.contentHeight - 2).
		Render(content.String())
}
