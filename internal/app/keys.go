// Package app contains the main application logic and UI model.
package app

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application.
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Tab navigation
	NextTab key.Binding
	PrevTab key.Binding
	Tab1    key.Binding
	Tab2    key.Binding
	Tab3    key.Binding
	Tab4    key.Binding

	// Pane navigation
	NextPane key.Binding
	PrevPane key.Binding

	// Actions
	Select      key.Binding
	Acknowledge key.Binding
	AckMessage  key.Binding
	Refresh     key.Binding

	// Filtering
	Filter         key.Binding
	ClearFilter    key.Binding
	SeverityFilter key.Binding

	// Modes
	Command key.Binding
	Help    key.Binding
	Escape  key.Binding
	Quit    key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("PgUp", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("PgDn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("Home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("End/G", "go to bottom"),
		),

		// Tab navigation
		NextTab: key.NewBinding(
			key.WithKeys("]", "L"),
			key.WithHelp("]/L", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("[", "H"),
			key.WithHelp("[/H", "prev tab"),
		),
		Tab1: key.NewBinding(
			key.WithKeys("F1"),
			key.WithHelp("F1", "Alerts tab"),
		),
		Tab2: key.NewBinding(
			key.WithKeys("F2"),
			key.WithHelp("F2", "Hosts tab"),
		),
		Tab3: key.NewBinding(
			key.WithKeys("F3"),
			key.WithHelp("F3", "Events tab"),
		),
		Tab4: key.NewBinding(
			key.WithKeys("F4"),
			key.WithHelp("F4", "Graphs tab"),
		),

		// Pane navigation
		NextPane: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("Tab", "next pane"),
		),
		PrevPane: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("Shift+Tab", "prev pane"),
		),

		// Actions
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "select/confirm"),
		),
		Acknowledge: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "acknowledge"),
		),
		AckMessage: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "ack with message"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),

		// Filtering
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("Ctrl+L", "clear filter"),
		),
		SeverityFilter: key.NewBinding(
			key.WithKeys("0", "1", "2", "3", "4", "5"),
			key.WithHelp("0-5", "severity filter"),
		),

		// Modes
		Command: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "command mode"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "cancel/close"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns key bindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.NextTab, k.NextPane, k.Acknowledge, k.Refresh, k.Help, k.Quit,
	}
}

// FullHelp returns key bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Home, k.End},
		// Tabs
		{k.NextTab, k.PrevTab, k.Tab1, k.Tab2, k.Tab3, k.Tab4},
		// Panes
		{k.NextPane, k.PrevPane, k.Select},
		// Actions
		{k.Acknowledge, k.AckMessage, k.Refresh},
		// Filtering & Modes
		{k.Filter, k.ClearFilter, k.Command, k.Help, k.Quit},
	}
}
