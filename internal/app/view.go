package app

import (
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

// View renders the entire application UI.
func (m Model) View() string {
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

	detailPane := m.detailPane.View()

	// Join panes horizontally
	contentArea := lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)

	// Stack everything vertically and scan for mouse zones
	return zone.Scan(lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		tabBar,
		contentArea,
		commandBar,
	))
}
