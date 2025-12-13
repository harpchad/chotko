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

		// Update detail pane with selected problem
		if selected := m.alertList.Selected(); selected != nil {
			m.detailPane.SetProblem(selected)
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
			cmds = append(cmds, m.loadProblems(), m.loadHostCounts())
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

	// Update focused component
	switch m.focused {
	case PaneAlerts:
		var cmd tea.Cmd
		m.alertList, cmd = m.alertList.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// Update detail when selection changes
		if selected := m.alertList.Selected(); selected != nil {
			m.detailPane.SetProblem(selected)
		}
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
			return m, tea.Batch(m.loadProblems(), m.loadHostCounts())
		}
		return m, nil

	// Tab navigation
	case key.Matches(msg, m.keys.NextTab):
		m.tabBar.Next()
		return m, nil
	case key.Matches(msg, m.keys.PrevTab):
		m.tabBar.Prev()
		return m, nil
	case key.Matches(msg, m.keys.Tab1):
		m.tabBar.SetActive(0)
		return m, nil
	case key.Matches(msg, m.keys.Tab2):
		m.tabBar.SetActive(1)
		return m, nil
	case key.Matches(msg, m.keys.Tab3):
		m.tabBar.SetActive(2)
		return m, nil

	// Pane navigation
	case key.Matches(msg, m.keys.NextPane):
		m.cycleFocus(1)
		return m, nil
	case key.Matches(msg, m.keys.PrevPane):
		m.cycleFocus(-1)
		return m, nil

	// Acknowledge
	case key.Matches(msg, m.keys.Acknowledge):
		if m.alertList.Selected() != nil {
			return m, m.acknowledgeProblem("")
		}
		return m, nil

	case key.Matches(msg, m.keys.AckMessage):
		if m.alertList.Selected() != nil {
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

	// Severity filter (0-5)
	case key.Matches(msg, m.keys.SeverityFilter):
		if severity, err := strconv.Atoi(msg.String()); err == nil {
			m.minSeverity = severity
			m.alertList.SetMinSeverity(severity)
			m.statusBar.SetFilter(m.minSeverity, m.textFilter)
		}
		return m, nil

	// Clear filter
	case key.Matches(msg, m.keys.ClearFilter):
		m.minSeverity = 0
		m.textFilter = ""
		m.alertList.SetMinSeverity(0)
		m.alertList.SetTextFilter("")
		m.statusBar.SetFilter(0, "")
		return m, nil
	}

	// Forward to focused component
	switch m.focused {
	case PaneAlerts:
		var cmd tea.Cmd
		m.alertList, cmd = m.alertList.Update(msg)
		// Update detail when selection changes
		if selected := m.alertList.Selected(); selected != nil {
			m.detailPane.SetProblem(selected)
		}
		return m, cmd
	case PaneDetail:
		var cmd tea.Cmd
		m.detailPane, cmd = m.detailPane.Update(msg)
		return m, cmd
	}

	return m, nil
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
			m.alertList.SetTextFilter(value)
			m.statusBar.SetFilter(m.minSeverity, m.textFilter)
		case command.ModeAckMessage:
			if m.alertList.Selected() != nil {
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
			return m, tea.Batch(m.loadProblems(), m.loadHostCounts())
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
	case PaneAlerts:
		m.alertList.SetFocused(false)
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
	case PaneAlerts:
		m.alertList.SetFocused(true)
	case PaneDetail:
		m.detailPane.SetFocused(true)
	}
}
