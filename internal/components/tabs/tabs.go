// Package tabs provides a tab bar component for navigation.
package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/harpchad/chotko/internal/theme"
)

// Model represents the tab bar component.
type Model struct {
	styles *theme.Styles
	tabs   []string
	active int
	width  int
}

// New creates a new tab bar model.
func New(styles *theme.Styles, tabs []string, active int) Model {
	return Model{
		styles: styles,
		tabs:   tabs,
		active: active,
	}
}

// SetWidth sets the tab bar width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetActive sets the active tab index.
func (m *Model) SetActive(index int) {
	if index >= 0 && index < len(m.tabs) {
		m.active = index
	}
}

// Active returns the current active tab index.
func (m Model) Active() int {
	return m.active
}

// ActiveTab returns the name of the active tab.
func (m Model) ActiveTab() string {
	if m.active < len(m.tabs) {
		return m.tabs[m.active]
	}
	return ""
}

// Next moves to the next tab.
func (m *Model) Next() {
	m.active = (m.active + 1) % len(m.tabs)
}

// Prev moves to the previous tab.
func (m *Model) Prev() {
	m.active = (m.active - 1 + len(m.tabs)) % len(m.tabs)
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
	var tabs []string

	for i, t := range m.tabs {
		var tab string
		tabID := fmt.Sprintf("tab_%d", i)
		if i == m.active {
			tab = m.styles.TabActive.Render(zone.Mark(tabID, "["+t+"]"))
		} else {
			tab = m.styles.TabInactive.Render(zone.Mark(tabID, " "+t+" "))
		}
		tabs = append(tabs, tab)
	}

	row := strings.Join(tabs, " ")
	return m.styles.TabBar.Width(m.width).Render(row)
}
