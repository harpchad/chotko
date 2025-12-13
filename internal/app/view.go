package app

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the entire application UI.
func (m Model) View() string {
	// Show error modal if active
	if m.showError || m.showHelp {
		return m.errorModal.View()
	}

	// Render components
	statusBar := m.statusBar.View()
	tabBar := m.tabBar.View()
	commandBar := m.commandInput.View()

	// Render main content area (alerts + detail panes)
	alertsPane := m.alertList.View()
	detailPane := m.detailPane.View()

	// Join panes horizontally
	contentArea := lipgloss.JoinHorizontal(lipgloss.Top, alertsPane, detailPane)

	// Stack everything vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		tabBar,
		contentArea,
		commandBar,
	)
}
