// Package command provides the command input component for the TUI.
package command

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/harpchad/chotko/internal/theme"
)

// Mode represents the command input mode.
type Mode int

// Mode constants for command input states.
const (
	ModeHidden Mode = iota
	ModeCommand
	ModeFilter
	ModeAckMessage
)

// Model represents the command input component.
type Model struct {
	styles *theme.Styles
	input  textinput.Model
	mode   Mode
	width  int
	hint   string
}

// New creates a new command input model.
func New(styles *theme.Styles) Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 256

	return Model{
		styles: styles,
		input:  ti,
		mode:   ModeHidden,
	}
}

// SetWidth sets the component width.
func (m *Model) SetWidth(width int) {
	m.width = width
	m.input.Width = width - 4
}

// SetMode activates a specific input mode.
func (m *Model) SetMode(mode Mode) {
	m.mode = mode
	m.input.Reset()

	switch mode {
	case ModeCommand:
		m.input.Prompt = ": "
		m.input.Placeholder = "command"
		m.hint = "Enter command (help for list)"
		m.input.Focus()
	case ModeFilter:
		m.input.Prompt = "/ "
		m.input.Placeholder = "filter"
		m.hint = "Type to filter alerts"
		m.input.Focus()
	case ModeAckMessage:
		m.input.Prompt = "Message: "
		m.input.Placeholder = "acknowledgment message"
		m.hint = "Enter message and press Enter"
		m.input.Focus()
	default:
		m.input.Blur()
		m.hint = ""
	}
}

// Hide hides the command input.
func (m *Model) Hide() {
	m.mode = ModeHidden
	m.input.Blur()
	m.input.Reset()
}

// Value returns the current input value.
func (m Model) Value() string {
	return m.input.Value()
}

// Mode returns the current mode.
func (m Model) Mode() Mode {
	return m.mode
}

// IsActive returns true if the command input is active.
func (m Model) IsActive() bool {
	return m.mode != ModeHidden
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.mode == ModeHidden {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	if m.mode == ModeHidden {
		// Show hint bar when hidden
		return m.styles.CommandHint.Width(m.width).Render(
			"Press : for commands, / to filter, ? for help",
		)
	}

	// Show active input (textinput.View() already includes the prompt)
	input := m.input.View()
	hint := ""
	if m.hint != "" {
		hint = "  " + m.styles.CommandHint.Render(m.hint)
	}

	return input + hint
}
