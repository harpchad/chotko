package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/harpchad/chotko/internal/components/command"
	"github.com/harpchad/chotko/internal/components/editor"
	"github.com/harpchad/chotko/internal/components/graphs"
	"github.com/harpchad/chotko/internal/ignores"
)

// Update handles all incoming messages and updates the model accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always handle refresh tick to keep timer running, regardless of modal state.
	// This prevents the auto-refresh from stopping when modals are visible.
	if _, ok := msg.(RefreshTickMsg); ok {
		return m.handleRefreshTickMsg()
	}

	// Handle editor modal first if visible
	if m.showEditor {
		return m.handleEditorUpdate(msg)
	}

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
		return m, nil
	}

	// Route messages to appropriate handlers
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.errorModal.SetScreenSize(msg.Width, msg.Height)
		return m, nil
	case tea.MouseMsg:
		return m.handleMouseMsg(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case ConnectedMsg:
		return m.handleConnectedMsg(msg)
	case DisconnectedMsg:
		return m.handleDisconnectedMsg(msg)
	case ProblemsLoadedMsg:
		return m.handleProblemsLoadedMsg(msg)
	case HostsLoadedMsg:
		return m.handleHostsLoadedMsg(msg)
	case EventsLoadedMsg:
		return m.handleEventsLoadedMsg(msg)
	case ItemsLoadedMsg:
		return m.handleItemsLoadedMsg(msg)
	case graphs.HostExpandedMsg:
		return m, m.loadHostHistory(msg.HostID)
	case HostHistoryLoadedMsg:
		return m.handleHostHistoryLoadedMsg(msg)
	case HostCountsLoadedMsg:
		return m.handleHostCountsLoadedMsg(msg)
	case AcknowledgeResultMsg:
		return m.handleAcknowledgeResultMsg(msg)
	case ErrorMsg:
		m.showError = true
		m.errorModal.ShowError(msg.Title, msg.Message, msg.Err)
		return m, nil
	case ClearErrorMsg:
		m.showError = false
		m.errorModal.Hide()
		return m, nil
	case HostTriggersLoadedMsg:
		return m.handleHostTriggersLoadedMsg(msg)
	case HostMacrosLoadedMsg:
		return m.handleHostMacrosLoadedMsg(msg)
	case TriggerUpdateResultMsg:
		return m.handleTriggerUpdateResultMsg(msg)
	case MacroUpdateResultMsg:
		return m.handleMacroUpdateResultMsg(msg)
	case HostUpdateResultMsg:
		return m.handleHostUpdateResultMsg(msg)
	}

	return m.handleFocusedComponentUpdate(msg)
}

// handleConnectedMsg handles successful connection to Zabbix.
func (m Model) handleConnectedMsg(msg ConnectedMsg) (tea.Model, tea.Cmd) {
	m.connected = true
	m.version = msg.Version
	m.client = msg.Client
	m.statusBar.SetConnected(true, msg.Version)
	return m, tea.Batch(m.loadProblems(), m.loadHostCounts(), m.updateWindowTitle())
}

// handleDisconnectedMsg handles disconnection from Zabbix.
func (m Model) handleDisconnectedMsg(msg DisconnectedMsg) (tea.Model, tea.Cmd) {
	m.connected = false
	m.statusBar.SetConnected(false, "")
	if msg.Err != nil {
		m.showError = true
		m.errorModal.ShowError("Connection Lost", "Lost connection to Zabbix server", msg.Err)
	}
	return m, m.updateWindowTitle()
}

// handleProblemsLoadedMsg handles loaded problems data.
func (m Model) handleProblemsLoadedMsg(msg ProblemsLoadedMsg) (tea.Model, tea.Cmd) {
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

	if m.tabBar.Active() == TabAlerts {
		if selected := m.alertList.Selected(); selected != nil {
			m.detailPane.SetProblem(selected)
		}
	}
	return m, m.updateWindowTitle()
}

// handleHostsLoadedMsg handles loaded hosts data.
func (m Model) handleHostsLoadedMsg(msg HostsLoadedMsg) (tea.Model, tea.Cmd) {
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

	if m.tabBar.Active() == TabHosts {
		if selected := m.hostList.Selected(); selected != nil {
			m.detailPane.SetHost(selected)
		}
	}
	return m, nil
}

// handleEventsLoadedMsg handles loaded events data.
func (m Model) handleEventsLoadedMsg(msg EventsLoadedMsg) (tea.Model, tea.Cmd) {
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

	if m.tabBar.Active() == TabEvents {
		if selected := m.eventList.Selected(); selected != nil {
			m.detailPane.SetEvent(selected)
		}
	}
	return m, nil
}

// handleItemsLoadedMsg handles loaded items data.
func (m Model) handleItemsLoadedMsg(msg ItemsLoadedMsg) (tea.Model, tea.Cmd) {
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

	if m.tabBar.Active() == TabGraphs {
		if selected := m.graphList.SelectedItem(); selected != nil {
			history := m.graphList.GetHistory(selected.ItemID)
			m.detailPane.SetItem(selected, history)
		}
	}
	return m, nil
}

// handleHostHistoryLoadedMsg handles loaded history data.
func (m Model) handleHostHistoryLoadedMsg(msg HostHistoryLoadedMsg) (tea.Model, tea.Cmd) {
	m.graphList.SetHostLoading(msg.HostID, false)

	if msg.Err != nil {
		return m, nil
	}

	m.graphList.MergeHistory(msg.History)

	if m.tabBar.Active() == TabGraphs {
		if selected := m.graphList.SelectedItem(); selected != nil {
			history := m.graphList.GetHistory(selected.ItemID)
			m.detailPane.SetItem(selected, history)
		}
	}
	return m, nil
}

// handleHostCountsLoadedMsg handles loaded host counts.
func (m Model) handleHostCountsLoadedMsg(msg HostCountsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		return m, nil
	}
	m.hostCounts = msg.Counts
	m.statusBar.SetCounts(msg.Counts)
	return m, nil
}

// handleAcknowledgeResultMsg handles acknowledge result.
func (m Model) handleAcknowledgeResultMsg(msg AcknowledgeResultMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.showError = true
		m.errorModal.ShowError("Acknowledge Failed", "Could not acknowledge problem", msg.Err)
		return m, nil
	}
	return m, m.loadProblems()
}

// handleRefreshTickMsg handles periodic refresh.
func (m Model) handleRefreshTickMsg() (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if !m.loading && m.connected {
		m.loading = true
		m.statusBar.SetLoading(true)
		cmds = append(cmds, m.loadDataForCurrentTab()...)
	}
	cmds = append(cmds, m.tickRefresh())
	return m, tea.Batch(cmds...)
}

// handleHostTriggersLoadedMsg handles loaded triggers for editor.
func (m Model) handleHostTriggersLoadedMsg(msg HostTriggersLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.showError = true
		m.errorModal.ShowError("Failed to Load Triggers", "Could not retrieve triggers from Zabbix", msg.Err)
		return m, nil
	}
	host := m.findHostByID(msg.HostID)
	if host != nil {
		m.editorPane.ShowHostTriggers(host, msg.Triggers, msg.SelectTriggerID)
		m.showEditor = true
	}
	return m, nil
}

// handleHostMacrosLoadedMsg handles loaded macros for editor.
func (m Model) handleHostMacrosLoadedMsg(msg HostMacrosLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.showError = true
		m.errorModal.ShowError("Failed to Load Macros", "Could not retrieve macros from Zabbix", msg.Err)
		return m, nil
	}
	host := m.findHostByID(msg.HostID)
	if host != nil {
		m.editorPane.ShowHostMacros(host, msg.Macros)
		m.showEditor = true
	}
	return m, nil
}

// handleTriggerUpdateResultMsg handles trigger update result.
func (m Model) handleTriggerUpdateResultMsg(msg TriggerUpdateResultMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.showError = true
		m.errorModal.ShowError("Trigger Update Failed", "Could not update trigger", msg.Err)
		return m, nil
	}
	hostID := m.getSelectedHostID()
	if hostID != "" {
		return m, tea.Batch(m.loadHosts(), m.loadHostTriggers(hostID, msg.TriggerID))
	}
	return m, m.loadHosts()
}

// handleMacroUpdateResultMsg handles macro update result.
func (m Model) handleMacroUpdateResultMsg(msg MacroUpdateResultMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.showError = true
		m.errorModal.ShowError("Macro Update Failed", "Could not update macro", msg.Err)
		return m, nil
	}
	if selected := m.hostList.Selected(); selected != nil {
		return m, m.loadHostMacros(selected.HostID)
	}
	return m, nil
}

// handleHostUpdateResultMsg handles host update result.
func (m Model) handleHostUpdateResultMsg(msg HostUpdateResultMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.showError = true
		m.errorModal.ShowError("Host Update Failed", "Could not update host", msg.Err)
		return m, nil
	}
	return m, m.loadHosts()
}

// handleFocusedComponentUpdate updates the focused component.
func (m Model) handleFocusedComponentUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
	// Handle ignore confirmation mode first
	if m.awaitingIgnoreConfirm {
		return m.handleIgnoreConfirm(msg)
	}

	if m.commandInput.IsActive() {
		return m.handleCommandInput(msg)
	}

	// Global keys
	if model, cmd, handled := m.handleGlobalKeys(msg); handled {
		return model, cmd
	}

	// Navigation keys
	if model, cmd, handled := m.handleNavigationKeys(msg); handled {
		return model, cmd
	}

	// Action keys
	if model, cmd, handled := m.handleActionKeys(msg); handled {
		return model, cmd
	}

	// Forward to focused component
	return m.forwardToFocusedComponent(msg)
}

// handleGlobalKeys handles quit, help, and refresh.
func (m Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.Shutdown()
		return m, tea.Quit, true
	case key.Matches(msg, m.keys.Help):
		m.showHelp = true
		m.errorModal.ShowHelp()
		return m, nil, true
	case key.Matches(msg, m.keys.Refresh):
		if !m.loading && m.connected {
			m.loading = true
			m.statusBar.SetLoading(true)
			return m, tea.Batch(m.loadDataForCurrentTab()...), true
		}
		return m, nil, true
	}
	return m, nil, false
}

// handleNavigationKeys handles tab and pane navigation.
func (m Model) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.NextTab):
		model, cmd := m.switchTab(m.tabBar.Active() + 1)
		return model, cmd, true
	case key.Matches(msg, m.keys.PrevTab):
		model, cmd := m.switchTab(m.tabBar.Active() - 1)
		return model, cmd, true
	case key.Matches(msg, m.keys.Tab1):
		model, cmd := m.switchTab(0)
		return model, cmd, true
	case key.Matches(msg, m.keys.Tab2):
		model, cmd := m.switchTab(1)
		return model, cmd, true
	case key.Matches(msg, m.keys.Tab3):
		model, cmd := m.switchTab(2)
		return model, cmd, true
	case key.Matches(msg, m.keys.Tab4):
		model, cmd := m.switchTab(3)
		return model, cmd, true
	case key.Matches(msg, m.keys.NextPane):
		m.cycleFocus(1)
		return m, nil, true
	case key.Matches(msg, m.keys.PrevPane):
		m.cycleFocus(-1)
		return m, nil, true
	}
	return m, nil, false
}

// handleActionKeys handles acknowledge, filter, edit, and other actions.
func (m Model) handleActionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Acknowledge):
		if m.tabBar.Active() == TabAlerts && m.alertList.Selected() != nil {
			return m, m.acknowledgeProblem(""), true
		}
		return m, nil, true
	case key.Matches(msg, m.keys.AckMessage):
		if m.tabBar.Active() == TabAlerts && m.alertList.Selected() != nil {
			m.mode = ModeAckMessage
			m.commandInput.SetMode(command.ModeAckMessage)
		}
		return m, nil, true
	case key.Matches(msg, m.keys.Filter):
		m.mode = ModeFilter
		m.commandInput.SetMode(command.ModeFilter)
		return m, nil, true
	case key.Matches(msg, m.keys.Command):
		m.mode = ModeCommand
		m.commandInput.SetMode(command.ModeCommand)
		return m, nil, true
	case key.Matches(msg, m.keys.SeverityFilter):
		return m.handleSeverityFilter(msg)
	case key.Matches(msg, m.keys.EditTriggers):
		return m.handleEditTriggers()
	case key.Matches(msg, m.keys.EditMacros):
		return m.handleEditMacros()
	case key.Matches(msg, m.keys.ToggleMonitor):
		return m.handleToggleMonitor()
	case key.Matches(msg, m.keys.ClearFilter):
		return m.handleClearFilter()
	case key.Matches(msg, m.keys.Ignore):
		return m.handleIgnore()
	case key.Matches(msg, m.keys.ListIgnores):
		return m.handleListIgnores()
	}
	return m, nil, false
}

// handleSeverityFilter handles severity filter keys (0-5).
func (m Model) handleSeverityFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.tabBar.Active() == TabAlerts {
		if severity, err := strconv.Atoi(msg.String()); err == nil {
			m.minSeverity = severity
			m.alertList.SetMinSeverity(severity)
			m.statusBar.SetFilter(m.minSeverity, m.textFilter)
		}
	}
	return m, nil, true
}

// handleEditTriggers opens the trigger editor.
func (m Model) handleEditTriggers() (tea.Model, tea.Cmd, bool) {
	hostID, triggerID := m.getSelectedHostAndTriggerID()
	if hostID != "" {
		return m, m.loadHostTriggers(hostID, triggerID), true
	}
	return m, nil, true
}

// handleEditMacros opens the macro editor.
func (m Model) handleEditMacros() (tea.Model, tea.Cmd, bool) {
	hostID := m.getSelectedHostID()
	if hostID != "" {
		return m, m.loadHostMacros(hostID), true
	}
	return m, nil, true
}

// handleToggleMonitor toggles host monitoring.
func (m Model) handleToggleMonitor() (tea.Model, tea.Cmd, bool) {
	if m.tabBar.Active() == TabHosts && m.hostList.Selected() != nil {
		host := m.hostList.Selected()
		if host.IsMonitored() {
			return m, m.disableHost(host.HostID), true
		}
		return m, m.enableHost(host.HostID), true
	}
	return m, nil, true
}

// handleClearFilter clears all filters.
func (m Model) handleClearFilter() (tea.Model, tea.Cmd, bool) {
	m.minSeverity = 0
	m.textFilter = ""
	m.alertList.SetMinSeverity(0)
	m.alertList.SetTextFilter("")
	m.hostList.SetTextFilter("")
	m.eventList.SetTextFilter("")
	m.statusBar.SetFilter(0, "")
	return m, nil, true
}

// handleIgnore initiates the ignore flow for the selected alert.
func (m Model) handleIgnore() (tea.Model, tea.Cmd, bool) {
	// Only works on Alerts tab
	if m.tabBar.Active() != TabAlerts {
		return m, nil, true
	}

	selected := m.alertList.Selected()
	if selected == nil {
		return m, nil, true
	}

	// Get host and trigger info
	hostID, triggerID := m.getSelectedHostAndTriggerID()
	if hostID == "" || triggerID == "" {
		m.statusBar.SetStatus("Cannot ignore: no trigger associated")
		return m, nil, true
	}

	// Get human-readable names
	hostName := selected.HostName()
	triggerName := selected.Name

	// Check if already ignored
	if m.ignoreList != nil && m.ignoreList.IsIgnored(hostID, triggerID) {
		m.statusBar.SetStatus("Already ignored")
		return m, nil, true
	}

	// Set up pending ignore and await confirmation
	m.pendingIgnore = &ignores.Rule{
		HostID:      hostID,
		HostName:    hostName,
		TriggerID:   triggerID,
		TriggerName: triggerName,
	}
	m.awaitingIgnoreConfirm = true
	m.statusBar.SetStatus(fmt.Sprintf("Ignore %s / %s? (y/n)", hostName, truncate(triggerName, 30)))

	return m, nil, true
}

// handleIgnoreConfirm handles y/n/esc during ignore confirmation.
func (m Model) handleIgnoreConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Add the ignore rule
		if m.ignoreList != nil && m.pendingIgnore != nil {
			if err := m.ignoreList.Add(*m.pendingIgnore); err != nil {
				m.statusBar.SetStatus(err.Error())
			} else {
				// Save to disk
				if err := m.ignoreList.Save(); err != nil {
					m.statusBar.SetStatus(fmt.Sprintf("Ignored (save failed: %v)", err))
				} else {
					m.statusBar.SetStatus(fmt.Sprintf("Ignored: %s / %s", m.pendingIgnore.HostName, truncate(m.pendingIgnore.TriggerName, 20)))
				}
				// Refresh alerts to hide the ignored one
				m.alertList.SetIgnoreChecker(m.ignoreList.IsIgnored)
			}
		}
		m.pendingIgnore = nil
		m.awaitingIgnoreConfirm = false
		return m, nil

	case "n", "N", "esc":
		// Cancel
		m.statusBar.SetStatus("Canceled")
		m.pendingIgnore = nil
		m.awaitingIgnoreConfirm = false
		return m, nil
	}

	// Ignore other keys while awaiting confirmation
	return m, nil
}

// handleListIgnores shows the list of ignored alerts.
func (m Model) handleListIgnores() (tea.Model, tea.Cmd, bool) {
	m.showIgnoresModal()
	return m, nil, true
}

// showIgnoresModal displays the ignores list in a modal.
func (m *Model) showIgnoresModal() {
	if m.ignoreList == nil || m.ignoreList.Len() == 0 {
		m.showError = true
		m.errorModal.ShowMessage("Ignored Alerts", "No ignored alerts.\n\nPress 'i' on an alert to ignore it.")
		return
	}

	var sb strings.Builder
	sb.WriteString("Ignored host/trigger pairs:\n\n")

	rules := m.ignoreList.Rules()
	for i, rule := range rules {
		sb.WriteString(fmt.Sprintf("%2d. %s / %s\n", i+1, rule.HostName, truncate(rule.TriggerName, 40)))
	}

	sb.WriteString("\nUse :unignore N to remove a rule.")

	m.showError = true
	m.errorModal.ShowMessage("Ignored Alerts", sb.String())
}

// truncate truncates a string to maxLen characters with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// forwardToFocusedComponent forwards key events to the focused component.
func (m Model) forwardToFocusedComponent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	if newTab < 0 {
		newTab = TabCount - 1
	} else if newTab >= TabCount {
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
	if oldTab != newTab && m.connected {
		switch newTab {
		case TabAlerts:
			if len(m.problems) == 0 {
				m.statusBar.SetLoading(true)
				cmds = append(cmds, m.loadProblems())
			}
		case TabHosts:
			if len(m.hosts) == 0 {
				m.statusBar.SetLoading(true)
				cmds = append(cmds, m.loadHosts())
			}
		case TabEvents:
			if len(m.events) == 0 {
				m.statusBar.SetLoading(true)
				cmds = append(cmds, m.loadEvents())
			}
		case TabGraphs:
			if len(m.items) == 0 {
				m.statusBar.SetLoading(true)
				cmds = append(cmds, m.loadItems())
			}
			// History is now lazy-loaded when hosts are expanded
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
		default:
			// Other modes don't need special handling
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
	switch {
	case cmd == "q" || cmd == "quit" || cmd == "exit":
		m.Shutdown()
		return m, tea.Quit
	case cmd == "r" || cmd == "refresh":
		if !m.loading && m.connected {
			m.loading = true
			m.statusBar.SetLoading(true)
			return m, tea.Batch(m.loadDataForCurrentTab()...)
		}
	case cmd == "help":
		m.showHelp = true
		m.errorModal.ShowHelp()
	case cmd == "ignores":
		m.showIgnoresModal()
	case strings.HasPrefix(cmd, "unignore "):
		return m.handleUnignoreCommand(cmd)
	}
	return m, nil
}

// handleUnignoreCommand removes an ignore rule by index.
func (m Model) handleUnignoreCommand(cmd string) (tea.Model, tea.Cmd) {
	// Parse the index from "unignore N"
	parts := strings.Fields(cmd)
	if len(parts) != 2 {
		m.statusBar.SetStatus("Usage: :unignore N")
		return m, nil
	}

	index, err := strconv.Atoi(parts[1])
	if err != nil || index < 1 {
		m.statusBar.SetStatus(fmt.Sprintf("Invalid index: %s", parts[1]))
		return m, nil
	}

	if m.ignoreList == nil {
		m.statusBar.SetStatus("No ignore list loaded")
		return m, nil
	}

	removed, ok := m.ignoreList.Remove(index)
	if !ok {
		m.statusBar.SetStatus(fmt.Sprintf("Invalid index: %d", index))
		return m, nil
	}

	// Save to disk
	if err := m.ignoreList.Save(); err != nil {
		m.statusBar.SetStatus(fmt.Sprintf("Removed (save failed: %v)", err))
	} else {
		m.statusBar.SetStatus(fmt.Sprintf("Removed: %s / %s", removed.HostName, truncate(removed.TriggerName, 20)))
	}

	// Refresh alerts to show the previously ignored one
	m.alertList.SetIgnoreChecker(m.ignoreList.IsIgnored)

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

// handleMouseMsg processes mouse input.
func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Scroll speed: 3 lines per wheel tick
	const scrollSpeed = 3

	// Handle scroll wheel - scrolls pane under mouse, not focused pane
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return m.handleScroll(-scrollSpeed, msg.X, msg.Y)
	case tea.MouseButtonWheelDown:
		return m.handleScroll(scrollSpeed, msg.X, msg.Y)
	default:
		// Other mouse buttons handled below or ignored
	}

	// Handle left click release (not press, to avoid double-firing)
	if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
		return m.handleClick(msg.X, msg.Y)
	}

	return m, nil
}

// handleScroll handles scroll wheel events, scrolling the pane under the mouse.
func (m Model) handleScroll(delta, mouseX, mouseY int) (tea.Model, tea.Cmd) {
	// Check if mouse is in content area (not in status bar, tab bar, or command bar)
	if mouseY < m.contentY || mouseY >= m.contentY+m.contentHeight {
		return m, nil
	}

	// Determine which pane the mouse is over and scroll it
	if mouseX < m.listPaneWidth {
		// Mouse is over list pane - scroll the current tab's list
		switch m.tabBar.Active() {
		case TabAlerts:
			m.alertList.Scroll(delta)
		case TabHosts:
			m.hostList.Scroll(delta)
		case TabEvents:
			m.eventList.Scroll(delta)
		case TabGraphs:
			m.graphList.Scroll(delta)
		}
	} else {
		// Mouse is over detail pane
		m.detailPane.Scroll(delta)
	}

	return m, nil
}

// handleClick handles left mouse button clicks.
func (m Model) handleClick(mouseX, mouseY int) (tea.Model, tea.Cmd) {
	// Check for tab clicks using zone detection
	for i := 0; i < 4; i++ {
		tabID := fmt.Sprintf("tab_%d", i)
		if zone.Get(tabID).InBounds(tea.MouseMsg{X: mouseX, Y: mouseY}) {
			return m.switchTab(i)
		}
	}

	// Check if click is in content area
	if mouseY >= m.contentY && mouseY < m.contentY+m.contentHeight {
		// Determine which pane was clicked and set focus
		if mouseX < m.listPaneWidth {
			// Clicked on list pane
			if m.focused != PaneList {
				m.setFocus(PaneList)
			}

			// Check for list item clicks
			return m.handleListClick(mouseX, mouseY)
		}

		// Clicked on detail pane
		if m.focused != PaneDetail {
			m.setFocus(PaneDetail)
		}
	}

	return m, nil
}

// handleListClick handles clicks on list items.
func (m Model) handleListClick(mouseX, mouseY int) (tea.Model, tea.Cmd) {
	switch m.tabBar.Active() {
	case TabAlerts:
		// Check for alert row clicks
		for i := 0; i < m.alertList.FilteredCount(); i++ {
			rowID := fmt.Sprintf("alert_%d", i)
			if zone.Get(rowID).InBounds(tea.MouseMsg{X: mouseX, Y: mouseY}) {
				m.alertList.SetCursor(i)
				if selected := m.alertList.Selected(); selected != nil {
					m.detailPane.SetProblem(selected)
				}
				return m, nil
			}
		}
	case TabHosts:
		// Check for host row clicks
		for i := 0; i < m.hostList.FilteredCount(); i++ {
			rowID := fmt.Sprintf("host_%d", i)
			if zone.Get(rowID).InBounds(tea.MouseMsg{X: mouseX, Y: mouseY}) {
				m.hostList.SetCursor(i)
				if selected := m.hostList.Selected(); selected != nil {
					m.detailPane.SetHost(selected)
				}
				return m, nil
			}
		}
	case TabEvents:
		// Check for event row clicks
		for i := 0; i < m.eventList.FilteredCount(); i++ {
			rowID := fmt.Sprintf("event_%d", i)
			if zone.Get(rowID).InBounds(tea.MouseMsg{X: mouseX, Y: mouseY}) {
				m.eventList.SetCursor(i)
				if selected := m.eventList.Selected(); selected != nil {
					m.detailPane.SetEvent(selected)
				}
				return m, nil
			}
		}
	case TabGraphs:
		// Check for graph tree node clicks
		return m.handleGraphsClick(mouseX, mouseY)
	}

	return m, nil
}

// handleGraphsClick handles clicks on the graphs tree view.
func (m Model) handleGraphsClick(mouseX, mouseY int) (tea.Model, tea.Cmd) {
	// Check for tree node clicks
	nodeCount := m.graphList.VisibleNodeCount()
	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("graph_node_%d", i)
		if zone.Get(nodeID).InBounds(tea.MouseMsg{X: mouseX, Y: mouseY}) {
			// Select and potentially toggle the node
			cmd := m.graphList.ClickNode(i)
			if selected := m.graphList.SelectedItem(); selected != nil {
				history := m.graphList.GetHistory(selected.ItemID)
				m.detailPane.SetItem(selected, history)
			}
			return m, cmd
		}
	}
	return m, nil
}

// setFocus sets focus to the specified pane.
func (m *Model) setFocus(pane Pane) {
	// Clear all focuses
	m.alertList.SetFocused(false)
	m.hostList.SetFocused(false)
	m.eventList.SetFocused(false)
	m.graphList.SetFocused(false)
	m.detailPane.SetFocused(false)

	m.focused = pane

	// Set new focus
	switch pane {
	case PaneList:
		m.updateListFocus()
	case PaneDetail:
		m.detailPane.SetFocused(true)
	}
}

// handleEditorUpdate handles messages when the editor modal is visible.
func (m Model) handleEditorUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle editor-specific messages
	switch msg := msg.(type) {
	case editor.TriggerToggleMsg:
		// Trigger enable/disable request
		m.editorPane.Hide()
		m.showEditor = false
		return m, m.toggleTrigger(msg.TriggerID, msg.Enable, msg.HostID)

	case editor.MacroEditedMsg:
		// Macro value changed
		return m, m.updateHostMacro(msg.MacroID, msg.NewValue, msg.HostID)

	case editor.MacroDeleteMsg:
		// Macro delete request
		m.editorPane.Hide()
		m.showEditor = false
		return m, m.deleteHostMacro(msg.MacroID, msg.HostID)

	case tea.KeyMsg:
		// Forward to editor
		var cmd tea.Cmd
		m.editorPane, cmd = m.editorPane.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check if editor was closed
		if !m.editorPane.Visible() {
			m.showEditor = false
		}
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	}

	// Forward other messages to editor
	var cmd tea.Cmd
	m.editorPane, cmd = m.editorPane.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
