// Package editor provides modal editing components for hosts, triggers, and macros.
package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// Type represents the type of editor.
type Type int

// Type constants.
const (
	TypeNone Type = iota
	TypeHost
	TypeTrigger
	TypeMacro
	TypeHostTriggers // List of triggers for a host
	TypeHostMacros   // List of macros for a host
)

// Field represents an editable field.
type Field struct {
	Label    string
	Value    string
	ReadOnly bool
	Input    textinput.Model
}

// TriggerItem represents a trigger in the trigger list.
type TriggerItem struct {
	Trigger  zabbix.Trigger
	Selected bool
}

// MacroItem represents a macro in the macro list.
type MacroItem struct {
	Macro    zabbix.HostMacro
	Selected bool
	Editing  bool
}

// Model represents the editor modal component.
type Model struct {
	styles       *theme.Styles
	visible      bool
	editorType   Type
	title        string
	width        int
	height       int
	screenWidth  int
	screenHeight int

	// Host being edited
	host *zabbix.Host

	// Trigger list
	triggers      []TriggerItem
	triggerCursor int
	triggerOffset int

	// Macro list
	macros      []MacroItem
	macroCursor int
	macroOffset int

	// Edit mode for macros
	editingMacroValue textinput.Model
	editingMacroIdx   int

	// Confirmation state
	confirmAction string
	confirmTarget string
}

// New creates a new editor model.
func New(styles *theme.Styles) Model {
	return Model{
		styles:          styles,
		editingMacroIdx: -1,
	}
}

// SetScreenSize sets the screen dimensions for centering.
func (m *Model) SetScreenSize(width, height int) {
	m.screenWidth = width
	m.screenHeight = height
	// Set modal size based on screen size (80% width, 70% height)
	m.width = width * 80 / 100
	if m.width < 60 {
		m.width = 60
	}
	if m.width > 120 {
		m.width = 120
	}
	m.height = height * 70 / 100
	if m.height < 15 {
		m.height = 15
	}
	if m.height > 40 {
		m.height = 40
	}
}

// ShowHostTriggers opens the host triggers editor.
// If selectTriggerID is non-empty, the cursor will be positioned on that trigger.
func (m *Model) ShowHostTriggers(host *zabbix.Host, triggers []zabbix.Trigger, selectTriggerID string) {
	m.visible = true
	m.editorType = TypeHostTriggers
	m.title = fmt.Sprintf("Triggers: %s", host.DisplayName())
	m.host = host
	m.triggerCursor = 0
	m.triggerOffset = 0
	m.confirmAction = ""

	// Convert to trigger items
	m.triggers = make([]TriggerItem, len(triggers))
	for i, t := range triggers {
		m.triggers[i] = TriggerItem{Trigger: t}
		// Position cursor on the selected trigger
		if selectTriggerID != "" && t.TriggerID == selectTriggerID {
			m.triggerCursor = i
		}
	}

	// Adjust offset to ensure selected trigger is visible
	if m.triggerCursor > 0 {
		maxVisible := m.height - 12
		if maxVisible < 3 {
			maxVisible = 3
		}
		if m.triggerCursor >= maxVisible {
			m.triggerOffset = m.triggerCursor - maxVisible/2
		}
	}
}

// ShowHostMacros opens the host macros editor.
func (m *Model) ShowHostMacros(host *zabbix.Host, macros []zabbix.HostMacro) {
	m.visible = true
	m.editorType = TypeHostMacros
	m.title = fmt.Sprintf("Macros: %s", host.DisplayName())
	m.host = host
	m.macroCursor = 0
	m.macroOffset = 0
	m.confirmAction = ""
	m.editingMacroIdx = -1

	// Convert to macro items
	m.macros = make([]MacroItem, len(macros))
	for i, macro := range macros {
		m.macros[i] = MacroItem{Macro: macro}
	}
}

// Hide hides the editor.
func (m *Model) Hide() {
	m.visible = false
	m.editorType = TypeNone
	m.confirmAction = ""
	m.editingMacroIdx = -1
}

// Visible returns true if the editor is visible.
func (m Model) Visible() bool {
	return m.visible
}

// Type returns the current editor type.
func (m Model) Type() Type {
	return m.editorType
}

// Host returns the host being edited.
func (m Model) Host() *zabbix.Host {
	return m.host
}

// SelectedTrigger returns the currently selected trigger.
func (m Model) SelectedTrigger() *zabbix.Trigger {
	if m.editorType != TypeHostTriggers || len(m.triggers) == 0 {
		return nil
	}
	return &m.triggers[m.triggerCursor].Trigger
}

// SelectedMacro returns the currently selected macro.
func (m Model) SelectedMacro() *zabbix.HostMacro {
	if m.editorType != TypeHostMacros || len(m.macros) == 0 {
		return nil
	}
	return &m.macros[m.macroCursor].Macro
}

// IsConfirming returns true if a confirmation is pending.
func (m Model) IsConfirming() bool {
	return m.confirmAction != ""
}

// ConfirmAction returns the pending confirmation action.
func (m Model) ConfirmAction() string {
	return m.confirmAction
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

	// Handle editing macro value
	if m.editingMacroIdx >= 0 {
		return m.updateMacroEdit(msg)
	}

	// Handle confirmation
	if m.confirmAction != "" {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.updateConfirm(keyMsg)
		}
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch m.editorType {
		case TypeHostTriggers:
			return m.updateTriggerList(keyMsg)
		case TypeHostMacros:
			return m.updateMacroList(keyMsg)
		}
	}

	return m, nil
}

// updateTriggerList handles key input for the trigger list.
func (m Model) updateTriggerList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Hide()
		return m, nil

	case "up", "k":
		if m.triggerCursor > 0 {
			m.triggerCursor--
			if m.triggerCursor < m.triggerOffset {
				m.triggerOffset = m.triggerCursor
			}
		}

	case "down", "j":
		if m.triggerCursor < len(m.triggers)-1 {
			m.triggerCursor++
			maxVisible := m.height - 10
			if m.triggerCursor >= m.triggerOffset+maxVisible {
				m.triggerOffset = m.triggerCursor - maxVisible + 1
			}
		}

	case " ":
		// Toggle enable/disable for selected trigger
		if len(m.triggers) > 0 {
			t := &m.triggers[m.triggerCursor].Trigger
			if t.IsEnabled() {
				m.confirmAction = "disable"
				m.confirmTarget = t.Description
			} else {
				m.confirmAction = "enable"
				m.confirmTarget = t.Description
			}
		}
	}

	return m, nil
}

// updateMacroList handles key input for the macro list.
func (m Model) updateMacroList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Hide()
		return m, nil

	case "up", "k":
		if m.macroCursor > 0 {
			m.macroCursor--
			if m.macroCursor < m.macroOffset {
				m.macroOffset = m.macroCursor
			}
		}

	case "down", "j":
		if m.macroCursor < len(m.macros)-1 {
			m.macroCursor++
			maxVisible := m.height - 10
			if m.macroCursor >= m.macroOffset+maxVisible {
				m.macroOffset = m.macroCursor - maxVisible + 1
			}
		}

	case "e", "enter":
		// Edit macro value
		if len(m.macros) > 0 {
			m.editingMacroIdx = m.macroCursor
			macro := m.macros[m.macroCursor].Macro

			ti := textinput.New()
			ti.SetValue(macro.Value)
			ti.Focus()
			ti.CharLimit = 2048
			ti.Width = m.width - 20
			m.editingMacroValue = ti
		}

	case "d":
		// Delete macro
		if len(m.macros) > 0 {
			m.confirmAction = "delete"
			m.confirmTarget = m.macros[m.macroCursor].Macro.Macro
		}
	}

	return m, nil
}

// updateMacroEdit handles key input when editing a macro value.
func (m Model) updateMacroEdit(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.editingMacroIdx = -1
			return m, nil
		case "enter":
			// Save the macro value
			if m.editingMacroIdx >= 0 && m.editingMacroIdx < len(m.macros) {
				newValue := m.editingMacroValue.Value()
				macro := m.macros[m.editingMacroIdx].Macro

				m.editingMacroIdx = -1

				// Return command to update macro
				return m, func() tea.Msg {
					return MacroEditedMsg{
						MacroID:  macro.HostMacroID,
						Macro:    macro.Macro,
						NewValue: newValue,
						HostID:   m.host.HostID,
					}
				}
			}
			m.editingMacroIdx = -1
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.editingMacroValue, cmd = m.editingMacroValue.Update(msg)
	return m, cmd
}

// updateConfirm handles key input during confirmation.
func (m Model) updateConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		action := m.confirmAction
		m.confirmAction = ""

		switch m.editorType {
		case TypeHostTriggers:
			if len(m.triggers) > 0 {
				t := m.triggers[m.triggerCursor].Trigger
				return m, func() tea.Msg {
					return TriggerToggleMsg{
						TriggerID: t.TriggerID,
						Enable:    action == "enable",
						HostID:    m.host.HostID,
					}
				}
			}
		case TypeHostMacros:
			if action == "delete" && len(m.macros) > 0 {
				macro := m.macros[m.macroCursor].Macro
				return m, func() tea.Msg {
					return MacroDeleteMsg{
						MacroID: macro.HostMacroID,
						Macro:   macro.Macro,
						HostID:  m.host.HostID,
					}
				}
			}
		}

	case "n", "N", "esc":
		m.confirmAction = ""
	}

	return m, nil
}

// TriggerToggleMsg is sent when a trigger should be enabled/disabled.
type TriggerToggleMsg struct {
	TriggerID string
	Enable    bool
	HostID    string
}

// MacroEditedMsg is sent when a macro value is edited.
type MacroEditedMsg struct {
	MacroID  string
	Macro    string
	NewValue string
	HostID   string
}

// MacroDeleteMsg is sent when a macro should be deleted.
type MacroDeleteMsg struct {
	MacroID string
	Macro   string
	HostID  string
}

// View implements tea.Model.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(m.styles.ModalTitle.Render(m.title))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("─", m.width-4))
	content.WriteString("\n\n")

	switch m.editorType {
	case TypeHostTriggers:
		content.WriteString(m.viewTriggerList())
	case TypeHostMacros:
		content.WriteString(m.viewMacroList())
	}

	// Confirmation overlay
	if m.confirmAction != "" {
		content.WriteString("\n")
		content.WriteString(strings.Repeat("─", m.width-4))
		content.WriteString("\n")
		confirmMsg := fmt.Sprintf("Are you sure you want to %s '%s'? (y/n)",
			m.confirmAction, truncate(m.confirmTarget, 30))
		content.WriteString(m.styles.AlertSeverity[4].Render(confirmMsg))
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

// viewTriggerList renders the trigger list view.
func (m Model) viewTriggerList() string {
	var b strings.Builder

	if len(m.triggers) == 0 {
		b.WriteString(m.styles.Subtle.Render("  No triggers found for this host"))
		b.WriteString("\n")
	} else {
		maxVisible := m.height - 12
		if maxVisible < 3 {
			maxVisible = 3
		}

		end := m.triggerOffset + maxVisible
		if end > len(m.triggers) {
			end = len(m.triggers)
		}

		for i := m.triggerOffset; i < end; i++ {
			t := m.triggers[i]
			isSelected := i == m.triggerCursor

			cursor := "  "
			if isSelected {
				cursor = "> "
			}

			// Status indicator
			statusText := "[ON] "
			if t.Trigger.IsDisabled() {
				statusText = "[OFF]"
			}

			// Problem indicator
			problemText := ""
			if t.Trigger.IsProblem() {
				problemText = " PROBLEM"
			}

			// Priority/severity
			priority := theme.SeverityName(t.Trigger.PriorityInt())
			desc := truncate(t.Trigger.Description, m.width-35)

			// Build the line - apply styles only if not selected
			var line string
			if isSelected {
				// Plain text for selected row, will be styled as a whole
				line = fmt.Sprintf("%s%-5s [%-4s] %s%s", cursor, statusText, priority, desc, problemText)
				// Pad to full width for consistent highlight
				if len(line) < m.width-6 {
					line += strings.Repeat(" ", m.width-6-len(line))
				}
				b.WriteString(m.styles.AlertSelected.Render(line))
			} else {
				// Apply individual styles for non-selected rows
				var status string
				if t.Trigger.IsDisabled() {
					status = m.styles.StatusUnknown.Render(statusText)
				} else {
					status = m.styles.StatusOK.Render(statusText)
				}

				priorityStyle := m.styles.AlertSeverity[t.Trigger.PriorityInt()]
				var problem string
				if problemText != "" {
					problem = m.styles.StatusProblem.Render(problemText)
				}

				line = fmt.Sprintf("%s%s %s %s%s", cursor, status,
					priorityStyle.Render(fmt.Sprintf("[%s]", priority)), desc, problem)
				b.WriteString(line)
			}
			b.WriteString("\n")
		}

		// Scroll indicator
		if len(m.triggers) > maxVisible {
			b.WriteString(m.styles.Subtle.Render(
				fmt.Sprintf("\n  (%d/%d triggers)", m.triggerCursor+1, len(m.triggers))))
		}
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width-4))
	b.WriteString("\n")
	b.WriteString(m.styles.Subtle.Render("[Space] toggle  [Esc] close"))

	return b.String()
}

// viewMacroList renders the macro list view.
func (m Model) viewMacroList() string {
	var b strings.Builder

	if len(m.macros) == 0 {
		b.WriteString(m.styles.Subtle.Render("  No macros found for this host"))
		b.WriteString("\n")
	} else {
		maxVisible := m.height - 12
		if maxVisible < 3 {
			maxVisible = 3
		}

		end := m.macroOffset + maxVisible
		if end > len(m.macros) {
			end = len(m.macros)
		}

		for i := m.macroOffset; i < end; i++ {
			macro := m.macros[i]
			cursor := "  "
			if i == m.macroCursor {
				cursor = "> "
			}

			// Editing indicator
			if i == m.editingMacroIdx {
				// Show edit input
				line := fmt.Sprintf("%s%s = ", cursor, macro.Macro.Macro)
				b.WriteString(line)
				b.WriteString(m.editingMacroValue.View())
				b.WriteString("\n")
				continue
			}

			// Show value (mask if secret)
			value := macro.Macro.Value
			if macro.Macro.Type == zabbix.MacroTypeSecret {
				value = "******"
			}

			line := fmt.Sprintf("%s%s = %s", cursor, macro.Macro.Macro, truncate(value, m.width-len(macro.Macro.Macro)-10))

			if i == m.macroCursor {
				// Pad to full width for consistent highlight
				if len(line) < m.width-6 {
					line += strings.Repeat(" ", m.width-6-len(line))
				}
				b.WriteString(m.styles.AlertSelected.Render(line))
			} else {
				b.WriteString(line)
			}
			b.WriteString("\n")
		}

		// Scroll indicator
		if len(m.macros) > maxVisible {
			b.WriteString(m.styles.Subtle.Render(
				fmt.Sprintf("\n  (%d/%d macros)", m.macroCursor+1, len(m.macros))))
		}
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width-4))
	b.WriteString("\n")
	b.WriteString(m.styles.Subtle.Render("[e]dit value  [d]elete  [Esc] close"))

	return b.String()
}

// truncate truncates a string to the given length with ellipsis.
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}
