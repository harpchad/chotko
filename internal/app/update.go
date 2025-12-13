package app

import (
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/harpchad/chotko/internal/components/command"
)

// Update handles all incoming messages and updates the model accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle modals (help or error) first if visible
	if m.showHelp || m.showError {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if key.Matches(keyMsg, key.NewBinding(key.WithKeys("esc", "enter", "q", "?"))) {
				m.showHelp = false
				m.showError = false
				m.errorModal.Hide()
				return m, nil
			}
		}
		// Don't process other messages while modal is open
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.errorModal.SetScreenSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case ConnectedMsg:
		m.connected = true
		m.version = msg.Version
		m.client = msg.Client
		m.statusBar.SetConnected(true, msg.Version)
		// Load initial data
		return m, tea.Batch(m.loadProblems(), m.loadHostCounts())

	case DisconnectedMsg:
		m.connected = false
		m.statusBar.SetConnected(false, "")
		if msg.Err != nil {
			m.showError = true
			m.errorModal.ShowError("Connection Lost", "Lost connection to Zabbix server", msg.Err)
		}
		return m, nil

	case ProblemsLoadedMsg:
		m.loading = false
		m.statusBar.SetLoading(false)
		m.lastRefresh = time.Now()
		m.statusBar.SetLastUpdate(m.lastRefresh.Format("15:04:05"))

		if msg.Err != nil {
			m.showError = true
			m.errorModal.ShowError("Failed to Load Problems", "Could not retrieve problems from Zabbix", msg.Err)
			return m, nil
		}

		m.problems = msg.Problems
		m.alertList.SetProblems(msg.Problems)

		// Update detail pane with selected problem if on alerts tab
		if m.tabBar.Active() == TabAlerts {
			if selected := m.alertList.Selected(); selected != nil {
				m.detailPane.SetProblem(selected)
			}
		}
		return m, nil

	case HostsLoadedMsg:
		m.loading = false
		m.statusBar.SetLoading(false)
		m.lastRefresh = time.Now()
		m.statusBar.SetLastUpdate(m.lastRefresh.Format("15:04:05"))

		if msg.Err != nil {
			m.showError = true
			m.errorModal.ShowError("Failed to Load Hosts", "Could not retrieve hosts from Zabbix", msg.Err)
			return m, nil
		}

		m.hosts = msg.Hosts
		m.hostList.SetHosts(msg.Hosts)

		// Update detail pane with selected host if on hosts tab
		if m.tabBar.Active() == TabHosts {
			if selected := m.hostList.Selected(); selected != nil {
				m.detailPane.SetHost(selected)
			}
		}
		return m, nil

	case EventsLoadedMsg:
		m.loading = false
		m.statusBar.SetLoading(false)
		m.lastRefresh = time.Now()
		m.statusBar.SetLastUpdate(m.lastRefresh.Format("15:04:05"))

		if msg.Err != nil {
			m.showError = true
			m.errorModal.ShowError("Failed to Load Events", "Could not retrieve events from Zabbix", msg.Err)
			return m, nil
		}

		m.events = msg.Events
		m.eventList.SetEvents(msg.Events)

		// Update detail pane with selected event if on events tab
		if m.tabBar.Active() == TabEvents {
			if selected := m.eventList.Selected(); selected != nil {
				m.detailPane.SetEvent(selected)
			}
		}
		return m, nil

	case ItemsLoadedMsg:
		m.loading = false
		m.statusBar.SetLoading(false)
		m.lastRefresh = time.Now()
		m.statusBar.SetLastUpdate(m.lastRefresh.Format("15:04:05"))

		if msg.Err != nil {
			m.showError = true
			m.errorModal.ShowError("Failed to Load Items", "Could not retrieve items from Zabbix", msg.Err)
			return m, nil
		}

		m.items = msg.Items
		m.graphList.SetItems(msg.Items, m.config.GetGraphCategories())

		// Load history data for items and update detail pane
		if m.tabBar.Active() == TabGraphs {
			// Update detail pane with selected item (history will be empty until it loads)
			if selected := m.graphList.SelectedItem(); selected != nil {
				history := m.graphList.GetHistory(selected.ItemID)
				m.detailPane.SetItem(selected, history)
			}
			return m, m.loadItemHistory()
		}
		return m, nil

	case HistoryLoadedMsg:
		if msg.Err != nil {
			// Silent fail for history - not critical
			return m, nil
		}

		// Update graph list with history data
		m.graphList.SetHistory(msg.History)

		// Update detail pane with selected item if on graphs tab
		if m.tabBar.Active() == TabGraphs {
			if selected := m.graphList.SelectedItem(); selected != nil {
				history := m.graphList.GetHistory(selected.ItemID)
				m.detailPane.SetItem(selected, history)
			}
		}
		return m, nil

	case HostCountsLoadedMsg:
		if msg.Err != nil {
			// Silent fail for host counts
			return m, nil
		}
		m.hostCounts = msg.Counts
		m.statusBar.SetCounts(msg.Counts)
		return m, nil

	case AcknowledgeResultMsg:
		if msg.Err != nil {
			m.showError = true
			m.errorModal.ShowError("Acknowledge Failed", "Could not acknowledge problem", msg.Err)
			return m, nil
		}
		// Refresh data after successful ack
		return m, m.loadProblems()

	case RefreshTickMsg:
		if !m.loading && m.connected {
			m.loading = true
			m.statusBar.SetLoading(true)
			cmds = append(cmds, m.loadDataForCurrentTab()...)
		}
		cmds = append(cmds, m.tickRefresh())
		return m, tea.Batch(cmds...)

	case ErrorMsg:
		m.showError = true
		m.errorModal.ShowError(msg.Title, msg.Message, msg.Err)
		return m, nil

	case ClearErrorMsg:
		m.showError = false
		m.errorModal.Hide()
		return m, nil
	}

	// Update focused component based on current tab
	switch m.focused {
	case PaneList:
		return m.updateListPane(msg)
	case PaneDetail:
		var cmd tea.Cmd
		m.detailPane, cmd = m.detailPane.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update command input if active
	if m.commandInput.IsActive() {
		var cmd tea.Cmd
		m.commandInput, cmd = m.commandInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// updateListPane updates the appropriate list component based on current tab.
func (m Model) updateListPane(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch m.tabBar.Active() {
	case TabAlerts:
		var cmd tea.Cmd
		m.alertList, cmd = m.alertList.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// Update detail when selection changes
		if selected := m.alertList.Selected(); selected != nil {
			m.detailPane.SetProblem(selected)
		}
	case TabHosts:
		var cmd tea.Cmd
		m.hostList, cmd = m.hostList.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// Update detail when selection changes
		if selected := m.hostList.Selected(); selected != nil {
			m.detailPane.SetHost(selected)
		}
	case TabEvents:
		var cmd tea.Cmd
		m.eventList, cmd = m.eventList.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// Update detail when selection changes
		if selected := m.eventList.Selected(); selected != nil {
			m.detailPane.SetEvent(selected)
		}
	case TabGraphs:
		var cmd tea.Cmd
		m.graphList, cmd = m.graphList.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// Update detail when selection changes
		if selected := m.graphList.SelectedItem(); selected != nil {
			history := m.graphList.GetHistory(selected.ItemID)
			m.detailPane.SetItem(selected, history)
		}
	}

	return m, tea.Batch(cmds...)
}

// loadDataForCurrentTab returns commands to load data for the current tab.
func (m *Model) loadDataForCurrentTab() []tea.Cmd {
	cmds := []tea.Cmd{m.loadHostCounts()}

	switch m.tabBar.Active() {
	case TabAlerts:
		cmds = append(cmds, m.loadProblems())
	case TabHosts:
		cmds = append(cmds, m.loadHosts())
	case TabEvents:
		cmds = append(cmds, m.loadEvents())
	case TabGraphs:
		// Load both items and history for graphs tab
		// History will be loaded after ItemsLoadedMsg is received
		cmds = append(cmds, m.loadItems())
	default:
		// For other tabs, load problems by default
		cmds = append(cmds, m.loadProblems())
	}

	return cmds
}

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle command input mode
	if m.commandInput.IsActive() {
		return m.handleCommandInput(msg)
	}

	switch {
	// Quit
	case key.Matches(msg, m.keys.Quit):
		m.Shutdown()
		return m, tea.Quit

	// Help
	case key.Matches(msg, m.keys.Help):
		m.showHelp = true
		m.errorModal.ShowHelp()
		return m, nil

	// Refresh
	case key.Matches(msg, m.keys.Refresh):
		if !m.loading && m.connected {
			m.loading = true
			m.statusBar.SetLoading(true)
			return m, tea.Batch(m.loadDataForCurrentTab()...)
		}
		return m, nil

	// Tab navigation
	case key.Matches(msg, m.keys.NextTab):
		return m.switchTab(m.tabBar.Active() + 1)
	case key.Matches(msg, m.keys.PrevTab):
		return m.switchTab(m.tabBar.Active() - 1)
	case key.Matches(msg, m.keys.Tab1):
		return m.switchTab(0)
	case key.Matches(msg, m.keys.Tab2):
		return m.switchTab(1)
	case key.Matches(msg, m.keys.Tab3):
		return m.switchTab(2)
	case key.Matches(msg, m.keys.Tab4):
		return m.switchTab(3)

	// Pane navigation
	case key.Matches(msg, m.keys.NextPane):
		m.cycleFocus(1)
		return m, nil
	case key.Matches(msg, m.keys.PrevPane):
		m.cycleFocus(-1)
		return m, nil

	// Acknowledge (only on alerts tab)
	case key.Matches(msg, m.keys.Acknowledge):
		if m.tabBar.Active() == TabAlerts && m.alertList.Selected() != nil {
			return m, m.acknowledgeProblem("")
		}
		return m, nil

	case key.Matches(msg, m.keys.AckMessage):
		if m.tabBar.Active() == TabAlerts && m.alertList.Selected() != nil {
			m.mode = ModeAckMessage
			m.commandInput.SetMode(command.ModeAckMessage)
		}
		return m, nil

	// Filter mode
	case key.Matches(msg, m.keys.Filter):
		m.mode = ModeFilter
		m.commandInput.SetMode(command.ModeFilter)
		return m, nil

	// Command mode
	case key.Matches(msg, m.keys.Command):
		m.mode = ModeCommand
		m.commandInput.SetMode(command.ModeCommand)
		return m, nil

	// Severity filter (0-5) - only on alerts tab
	case key.Matches(msg, m.keys.SeverityFilter):
		if m.tabBar.Active() == TabAlerts {
			if severity, err := strconv.Atoi(msg.String()); err == nil {
				m.minSeverity = severity
				m.alertList.SetMinSeverity(severity)
				m.statusBar.SetFilter(m.minSeverity, m.textFilter)
			}
		}
		return m, nil

	// Clear filter
	case key.Matches(msg, m.keys.ClearFilter):
		m.minSeverity = 0
		m.textFilter = ""
		m.alertList.SetMinSeverity(0)
		m.alertList.SetTextFilter("")
		m.hostList.SetTextFilter("")
		m.eventList.SetTextFilter("")
		m.statusBar.SetFilter(0, "")
		return m, nil
	}

	// Forward to focused component
	switch m.focused {
	case PaneList:
		return m.updateListPane(msg)
	case PaneDetail:
		var cmd tea.Cmd
		m.detailPane, cmd = m.detailPane.Update(msg)
		return m, cmd
	}

	return m, nil
}

// switchTab switches to a new tab and loads appropriate data.
func (m Model) switchTab(newTab int) (tea.Model, tea.Cmd) {
	// Wrap around
	tabCount := 4 // Alerts, Hosts, Events, Graphs
	if newTab < 0 {
		newTab = tabCount - 1
	} else if newTab >= tabCount {
		newTab = 0
	}

	oldTab := m.tabBar.Active()
	m.tabBar.SetActive(newTab)

	// Update focus when switching tabs
	m.updateListFocus()

	// Update detail pane for new tab
	m.updateDetailForCurrentTab()

	// Load data if switching to a different tab type
	var cmds []tea.Cmd
	if oldTab != newTab && m.connected && !m.loading {
		switch newTab {
		case TabAlerts:
			if len(m.problems) == 0 {
				m.loading = true
				m.statusBar.SetLoading(true)
				cmds = append(cmds, m.loadProblems())
			}
		case TabHosts:
			if len(m.hosts) == 0 {
				m.loading = true
				m.statusBar.SetLoading(true)
				cmds = append(cmds, m.loadHosts())
			}
		case TabEvents:
			if len(m.events) == 0 {
				m.loading = true
				m.statusBar.SetLoading(true)
				cmds = append(cmds, m.loadEvents())
			}
		case TabGraphs:
			if len(m.items) == 0 {
				m.loading = true
				m.statusBar.SetLoading(true)
				cmds = append(cmds, m.loadItems())
			} else {
				// Items already loaded, but load history for them
				cmds = append(cmds, m.loadItemHistory())
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// updateListFocus updates the focus state for list components based on current tab.
func (m *Model) updateListFocus() {
	isFocused := m.focused == PaneList

	// Clear all focuses first
	m.alertList.SetFocused(false)
	m.hostList.SetFocused(false)
	m.eventList.SetFocused(false)
	m.graphList.SetFocused(false)

	// Set focus for the active tab's list
	switch m.tabBar.Active() {
	case TabAlerts:
		m.alertList.SetFocused(isFocused)
	case TabHosts:
		m.hostList.SetFocused(isFocused)
	case TabEvents:
		m.eventList.SetFocused(isFocused)
	case TabGraphs:
		m.graphList.SetFocused(isFocused)
	}
}

// updateDetailForCurrentTab updates the detail pane content for the current tab.
func (m *Model) updateDetailForCurrentTab() {
	switch m.tabBar.Active() {
	case TabAlerts:
		if selected := m.alertList.Selected(); selected != nil {
			m.detailPane.SetProblem(selected)
		} else {
			m.detailPane.SetProblem(nil)
		}
	case TabHosts:
		if selected := m.hostList.Selected(); selected != nil {
			m.detailPane.SetHost(selected)
		} else {
			m.detailPane.SetHost(nil)
		}
	case TabEvents:
		if selected := m.eventList.Selected(); selected != nil {
			m.detailPane.SetEvent(selected)
		} else {
			m.detailPane.SetEvent(nil)
		}
	case TabGraphs:
		if selected := m.graphList.SelectedItem(); selected != nil {
			history := m.graphList.GetHistory(selected.ItemID)
			m.detailPane.SetItem(selected, history)
		} else {
			m.detailPane.SetItem(nil, nil)
		}
	default:
		m.detailPane.Clear()
	}
}

// handleCommandInput processes input when in command/filter/ack mode.
func (m Model) handleCommandInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.commandInput.Hide()
		return m, nil

	case "enter":
		value := m.commandInput.Value()
		mode := m.commandInput.Mode()
		m.mode = ModeNormal
		m.commandInput.Hide()

		switch mode {
		case command.ModeFilter:
			m.textFilter = value
			// Apply filter to current tab's list
			switch m.tabBar.Active() {
			case TabAlerts:
				m.alertList.SetTextFilter(value)
			case TabHosts:
				m.hostList.SetTextFilter(value)
			case TabEvents:
				m.eventList.SetTextFilter(value)
			}
			m.statusBar.SetFilter(m.minSeverity, m.textFilter)
		case command.ModeAckMessage:
			if m.tabBar.Active() == TabAlerts && m.alertList.Selected() != nil {
				return m, m.acknowledgeProblem(value)
			}
		case command.ModeCommand:
			return m.executeCommand(value)
		}
		return m, nil
	}

	// Forward to command input
	var cmd tea.Cmd
	m.commandInput, cmd = m.commandInput.Update(msg)
	return m, cmd
}

// executeCommand processes a command entered in command mode.
func (m Model) executeCommand(cmd string) (tea.Model, tea.Cmd) {
	switch cmd {
	case "q", "quit", "exit":
		m.Shutdown()
		return m, tea.Quit
	case "r", "refresh":
		if !m.loading && m.connected {
			m.loading = true
			m.statusBar.SetLoading(true)
			return m, tea.Batch(m.loadDataForCurrentTab()...)
		}
	case "help":
		m.showHelp = true
		m.errorModal.ShowHelp()
	}
	return m, nil
}

// cycleFocus moves focus to the next or previous pane.
func (m *Model) cycleFocus(direction int) {
	// Update current pane focus
	switch m.focused {
	case PaneList:
		m.alertList.SetFocused(false)
		m.hostList.SetFocused(false)
		m.eventList.SetFocused(false)
		m.graphList.SetFocused(false)
	case PaneDetail:
		m.detailPane.SetFocused(false)
	}

	// Move to next/prev pane
	if direction > 0 {
		m.focused = (m.focused + 1) % 2
	} else {
		m.focused = (m.focused - 1 + 2) % 2
	}

	// Update new pane focus
	switch m.focused {
	case PaneList:
		m.updateListFocus()
	case PaneDetail:
		m.detailPane.SetFocused(true)
	}
}
