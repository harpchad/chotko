// Package modal provides modal dialog components for errors, help, and confirmations.
package modal

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harpchad/chotko/internal/theme"
)

// Type represents the type of modal.
type Type int

// Type constants for different modal dialogs.
const (
	TypeError Type = iota
	TypeHelp
)

// Model represents a modal dialog component.
type Model struct {
	styles       *theme.Styles
	visible      bool
	modalType    Type
	title        string
	message      string
	details      string
	width        int
	height       int
	screenWidth  int
	screenHeight int
}

// New creates a new modal model.
func New(styles *theme.Styles) Model {
	return Model{
		styles: styles,
		width:  60,
		height: 15,
	}
}

// SetScreenSize sets the screen dimensions for centering.
func (m *Model) SetScreenSize(width, height int) {
	m.screenWidth = width
	m.screenHeight = height
}

// Show displays the modal with the given content.
func (m *Model) Show(modalType Type, title, message string) {
	m.visible = true
	m.modalType = modalType
	m.title = title
	m.message = message
	m.details = ""
}

// ShowError displays an error modal.
func (m *Model) ShowError(title, message string, err error) {
	m.Show(TypeError, title, message)
	if err != nil {
		m.details = err.Error()
	}
}

// ShowHelp displays the help modal.
func (m *Model) ShowHelp() {
	m.visible = true
	m.modalType = TypeHelp
	m.title = "Keyboard Shortcuts"
	m.width = 50
	m.height = 24
}

// ShowMessage displays a simple message modal.
func (m *Model) ShowMessage(title, message string) {
	m.Show(TypeError, title, message)
}

// Hide hides the modal.
func (m *Model) Hide() {
	m.visible = false
}

// Visible returns true if the modal is visible.
func (m Model) Visible() bool {
	return m.visible
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc", "enter", "q"))) {
			m.Hide()
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(m.styles.ModalTitle.Render(m.title))
	content.WriteString("\n\n")

	if m.modalType == TypeHelp {
		content.WriteString(m.renderHelp())
	} else {
		// Message
		content.WriteString(m.styles.ModalText.Render(m.message))

		// Details (for errors)
		if m.details != "" {
			content.WriteString("\n\n")
			content.WriteString(m.styles.Subtle.Render(m.details))
		}

		// Button hint
		content.WriteString("\n\n")
		content.WriteString(m.styles.ModalButton.Render(" OK "))
		content.WriteString("  ")
		content.WriteString(m.styles.Subtle.Render("(Press Enter or Esc)"))
	}

	// Create modal box
	box := m.styles.ModalBox.
		Width(m.width).
		Render(content.String())

	// Center on screen
	return lipgloss.Place(
		m.screenWidth, m.screenHeight,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

// renderHelp renders the help content.
func (m Model) renderHelp() string {
	var b strings.Builder

	sections := []struct {
		title string
		keys  [][]string
	}{
		{
			title: "Navigation",
			keys: [][]string{
				{"↑/k", "Move up"},
				{"↓/j", "Move down"},
				{"PgUp/Ctrl+U", "Page up"},
				{"PgDn/Ctrl+D", "Page down"},
				{"Home/g", "Go to top"},
				{"End/G", "Go to bottom"},
			},
		},
		{
			title: "Tabs & Panes",
			keys: [][]string{
				{"]/L", "Next tab"},
				{"[/H", "Previous tab"},
				{"F1-F3", "Jump to tab"},
				{"Tab", "Next pane"},
				{"Shift+Tab", "Previous pane"},
			},
		},
		{
			title: "Actions",
			keys: [][]string{
				{"a", "Acknowledge problem"},
				{"A", "Acknowledge with message"},
				{"r", "Refresh data"},
				{"Enter", "Select/Confirm"},
			},
		},
		{
			title: "Host Editing (Hosts tab)",
			keys: [][]string{
				{"t", "Edit triggers"},
				{"m", "Edit macros"},
				{"e", "Enable/disable host"},
			},
		},
		{
			title: "Alert Ignoring (Alerts tab)",
			keys: [][]string{
				{"i", "Ignore alert locally"},
				{"I", "List ignored alerts"},
				{":unignore N", "Remove ignore rule"},
			},
		},
		{
			title: "Filtering",
			keys: [][]string{
				{"/", "Filter mode"},
				{"0-5", "Filter by severity"},
				{"Ctrl+L", "Clear filter"},
			},
		},
		{
			title: "General",
			keys: [][]string{
				{":", "Command mode"},
				{"?", "Show this help"},
				{"Esc", "Cancel/Close"},
				{"q", "Quit"},
			},
		},
	}

	for i, section := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(m.styles.Title.Render(section.title))
		b.WriteString("\n")

		for _, kv := range section.keys {
			key := m.styles.HelpKey.Width(14).Render(kv[0])
			desc := m.styles.HelpDesc.Render(kv[1])
			b.WriteString("  " + key + " " + desc + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Subtle.Render("Press Esc to close"))

	return b.String()
}
