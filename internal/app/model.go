package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/harpchad/chotko/internal/components/alerts"
	"github.com/harpchad/chotko/internal/components/command"
	"github.com/harpchad/chotko/internal/components/detail"
	"github.com/harpchad/chotko/internal/components/editor"
	"github.com/harpchad/chotko/internal/components/events"
	"github.com/harpchad/chotko/internal/components/graphs"
	"github.com/harpchad/chotko/internal/components/hosts"
	"github.com/harpchad/chotko/internal/components/modal"
	"github.com/harpchad/chotko/internal/components/statusbar"
	"github.com/harpchad/chotko/internal/components/tabs"
	"github.com/harpchad/chotko/internal/config"
	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// Pane represents a focusable pane in the UI.
type Pane int

// Pane constants for UI focus management.
const (
	PaneList Pane = iota // PaneList is the left pane (alerts or hosts list)
	PaneDetail
)

// Tab constants.
const (
	TabAlerts = 0
	TabHosts  = 1
	TabEvents = 2
	TabGraphs = 3
	TabCount  = 4
)

// Layout constants for UI rendering.
const (
	ListWidthPercent   = 45 // Percentage of width for list pane
	DetailWidthPercent = 55 // Percentage of width for detail pane
	LogoutTimeout      = 5  // Seconds to wait for logout on shutdown
)

// Zabbix object type constants.
const (
	ObjectTypeTrigger = "0" // Trigger-based problem
)

// Mode represents the current input mode.
type Mode int

// Mode constants for input handling.
const (
	ModeNormal Mode = iota
	ModeFilter
	ModeCommand
	ModeAckMessage
)

// Model is the main application model.
type Model struct {
	// Configuration
	config *config.Config

	// Zabbix client
	client *zabbix.Client

	// Theme and styles
	theme  *theme.Theme
	styles *theme.Styles

	// Key bindings
	keys KeyMap

	// Window dimensions
	width  int
	height int

	// Current state
	focused         Pane
	mode            Mode
	minSeverity     int
	textFilter      string
	refreshInterval time.Duration

	// Data
	problems   []zabbix.Problem
	hosts      []zabbix.Host
	events     []zabbix.Event
	items      []zabbix.Item
	hostCounts *zabbix.HostCounts

	// Components
	statusBar    statusbar.Model
	tabBar       tabs.Model
	alertList    alerts.Model
	hostList     hosts.Model
	eventList    events.Model
	graphList    graphs.Model
	detailPane   detail.Model
	commandInput command.Model

	// Modal state
	showHelp   bool
	showError  bool
	showEditor bool
	errorModal modal.Model
	editorPane editor.Model

	// Loading states
	loading     bool
	lastRefresh time.Time
	connected   bool
	version     string

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Mouse tracking - pane bounds for scroll detection
	listPaneX     int // X position where list pane starts (0)
	listPaneWidth int // Width of list pane including borders
	detailPaneX   int // X position where detail pane starts
	contentY      int // Y position where content panes start (after status bar and tab bar)
	contentHeight int // Height of content area
}

// New creates a new application model.
func New(cfg *config.Config, t *theme.Theme) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	styles := theme.NewStyles(t)

	m := &Model{
		config:          cfg,
		theme:           t,
		styles:          styles,
		keys:            DefaultKeyMap(),
		focused:         PaneList,
		mode:            ModeNormal,
		minSeverity:     cfg.Display.MinSeverity,
		refreshInterval: time.Duration(cfg.Display.RefreshInterval) * time.Second,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize components
	m.statusBar = statusbar.New(styles)
	m.tabBar = tabs.New(styles, []string{"Alerts", "Hosts", "Events", "Graphs"}, 0)
	m.alertList = alerts.New(styles)
	m.hostList = hosts.New(styles)
	m.eventList = events.New(styles)
	m.graphList = graphs.New(styles)
	m.detailPane = detail.New(styles)
	m.commandInput = command.New(styles)
	m.errorModal = modal.New(styles)
	m.editorPane = editor.New(styles)

	// Set initial focus to alerts list
	m.alertList.SetFocused(true)

	return m
}

// Init initializes the application.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.connect(),
		m.tickRefresh(),
	)
}

// connect establishes connection to Zabbix.
func (m *Model) connect() tea.Cmd {
	// Capture config values for the goroutine
	serverURL := m.config.Server.URL
	useToken := m.config.UseToken()
	token := m.config.Auth.Token
	username := m.config.Auth.Username
	password := m.config.Auth.Password
	ctx := m.ctx

	return func() tea.Msg {
		// Create client
		client := zabbix.NewClient(serverURL)

		// Authenticate
		if useToken {
			client.SetToken(token)
		} else {
			if err := client.Login(ctx, username, password); err != nil {
				return ErrorMsg{
					Title:   "Authentication Failed",
					Message: "Failed to connect to Zabbix server",
					Err:     err,
				}
			}
		}

		// Get version to verify connection
		version, err := client.Version(ctx)
		if err != nil {
			return ErrorMsg{
				Title:   "Connection Failed",
				Message: "Failed to connect to Zabbix server",
				Err:     err,
			}
		}

		return ConnectedMsg{Version: version, Client: client}
	}
}

// tickRefresh returns a command that triggers periodic refresh.
func (m *Model) tickRefresh() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(_ time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// loadProblems fetches problems from Zabbix.
func (m *Model) loadProblems() tea.Cmd {
	// Capture values for the goroutine
	client := m.client
	ctx := m.ctx
	minSeverity := m.minSeverity

	return func() tea.Msg {
		if client == nil {
			return ProblemsLoadedMsg{Err: nil}
		}

		var problems []zabbix.Problem
		var err error

		if minSeverity > 0 {
			problems, err = client.GetProblemsWithMinSeverity(ctx, minSeverity)
		} else {
			problems, err = client.GetActiveProblems(ctx)
		}

		return ProblemsLoadedMsg{
			Problems: problems,
			Err:      err,
		}
	}
}

// loadHostCounts fetches host status counts from Zabbix.
func (m *Model) loadHostCounts() tea.Cmd {
	// Capture values for the goroutine
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return HostCountsLoadedMsg{Err: nil}
		}

		counts, err := client.GetHostCounts(ctx)
		return HostCountsLoadedMsg{
			Counts: counts,
			Err:    err,
		}
	}
}

// loadHosts fetches all hosts from Zabbix.
func (m *Model) loadHosts() tea.Cmd {
	// Capture values for the goroutine
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return HostsLoadedMsg{Err: nil}
		}

		hosts, err := client.GetAllHosts(ctx)
		return HostsLoadedMsg{
			Hosts: hosts,
			Err:   err,
		}
	}
}

// loadEvents fetches recent events from Zabbix.
func (m *Model) loadEvents() tea.Cmd {
	// Capture values for the goroutine
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return EventsLoadedMsg{Err: nil}
		}

		// Get events from the last 24 hours, limit to 500
		events, err := client.GetRecentEvents(ctx, 24, 500)
		return EventsLoadedMsg{
			Events: events,
			Err:    err,
		}
	}
}

// loadItems fetches numeric items from Zabbix for the graphs tab.
func (m *Model) loadItems() tea.Cmd {
	// Capture values for the goroutine
	client := m.client
	ctx := m.ctx
	categories := m.config.GetGraphCategories()

	return func() tea.Msg {
		if client == nil {
			return ItemsLoadedMsg{Err: nil}
		}

		items, err := client.GetAllNumericItems(ctx, categories)
		return ItemsLoadedMsg{
			Items: items,
			Err:   err,
		}
	}
}

// loadHostHistory fetches history data for items belonging to a specific host.
func (m *Model) loadHostHistory(hostID string) tea.Cmd {
	// Capture values for the goroutine
	client := m.client
	ctx := m.ctx
	hours := m.config.GetHistoryHours()

	// Get items for this host from the graph list
	hostItems := m.graphList.GetHostItems(hostID)

	return func() tea.Msg {
		if client == nil || len(hostItems) == 0 {
			return HostHistoryLoadedMsg{HostID: hostID, History: nil, Err: nil}
		}

		history, err := client.GetItemsHistory(ctx, hostItems, hours)
		return HostHistoryLoadedMsg{
			HostID:  hostID,
			History: history,
			Err:     err,
		}
	}
}

// acknowledgeProblem sends an acknowledgment for the selected problem.
func (m *Model) acknowledgeProblem(message string) tea.Cmd {
	// Capture values for the goroutine
	client := m.client
	ctx := m.ctx
	selected := m.alertList.Selected()

	return func() tea.Msg {
		if client == nil || selected == nil {
			return AcknowledgeResultMsg{Success: false}
		}

		err := client.AcknowledgeProblem(ctx, selected.EventID, message)

		return AcknowledgeResultMsg{
			EventID: selected.EventID,
			Success: err == nil,
			Err:     err,
		}
	}
}

// loadHostTriggers fetches triggers for a specific host.
// If selectTriggerID is non-empty, that trigger will be pre-selected in the editor.
func (m *Model) loadHostTriggers(hostID, selectTriggerID string) tea.Cmd {
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return HostTriggersLoadedMsg{HostID: hostID, Err: nil}
		}

		triggers, err := client.GetHostTriggers(ctx, hostID)
		return HostTriggersLoadedMsg{
			HostID:          hostID,
			Triggers:        triggers,
			SelectTriggerID: selectTriggerID,
			Err:             err,
		}
	}
}

// loadHostMacros fetches macros for a specific host.
func (m *Model) loadHostMacros(hostID string) tea.Cmd {
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return HostMacrosLoadedMsg{HostID: hostID, Err: nil}
		}

		macros, err := client.GetHostMacros(ctx, hostID)
		return HostMacrosLoadedMsg{
			HostID: hostID,
			Macros: macros,
			Err:    err,
		}
	}
}

// toggleTrigger enables or disables a trigger.
func (m *Model) toggleTrigger(triggerID string, enable bool, _ string) tea.Cmd {
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return TriggerUpdateResultMsg{TriggerID: triggerID, Err: nil}
		}

		var err error
		var action string
		if enable {
			err = client.EnableTrigger(ctx, triggerID)
			action = "enable"
		} else {
			err = client.DisableTrigger(ctx, triggerID)
			action = "disable"
		}

		return TriggerUpdateResultMsg{
			TriggerID: triggerID,
			Action:    action,
			Success:   err == nil,
			Err:       err,
		}
	}
}

// updateHostMacro updates a macro value.
func (m *Model) updateHostMacro(macroID, value, _ string) tea.Cmd {
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return MacroUpdateResultMsg{MacroID: macroID, Err: nil}
		}

		err := client.UpdateHostMacro(ctx, macroID, value)
		return MacroUpdateResultMsg{
			MacroID: macroID,
			Action:  "update",
			Success: err == nil,
			Err:     err,
		}
	}
}

// deleteHostMacro deletes a macro.
func (m *Model) deleteHostMacro(macroID, _ string) tea.Cmd {
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return MacroUpdateResultMsg{MacroID: macroID, Err: nil}
		}

		err := client.DeleteHostMacro(ctx, macroID)
		return MacroUpdateResultMsg{
			MacroID: macroID,
			Action:  "delete",
			Success: err == nil,
			Err:     err,
		}
	}
}

// enableHost enables monitoring for a host.
func (m *Model) enableHost(hostID string) tea.Cmd {
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return HostUpdateResultMsg{HostID: hostID, Err: nil}
		}

		err := client.EnableHost(ctx, hostID)
		return HostUpdateResultMsg{
			HostID:  hostID,
			Action:  "enable",
			Success: err == nil,
			Err:     err,
		}
	}
}

// disableHost disables monitoring for a host.
func (m *Model) disableHost(hostID string) tea.Cmd {
	client := m.client
	ctx := m.ctx

	return func() tea.Msg {
		if client == nil {
			return HostUpdateResultMsg{HostID: hostID, Err: nil}
		}

		err := client.DisableHost(ctx, hostID)
		return HostUpdateResultMsg{
			HostID:  hostID,
			Action:  "disable",
			Success: err == nil,
			Err:     err,
		}
	}
}

// SetSize updates the window dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate pane sizes
	statusBarHeight := 1
	tabBarHeight := 1
	commandHeight := 1
	contentHeight := height - statusBarHeight - tabBarHeight - commandHeight - 4 // borders

	// Split width based on defined percentages
	// Account for borders: each pane has 2 chars (left+right border)
	availableWidth := width - 4 // 4 = 2 borders per pane * 2 panes
	listWidth := availableWidth * ListWidthPercent / 100
	detailWidth := availableWidth - listWidth

	// Store pane bounds for mouse tracking
	m.listPaneX = 0
	m.listPaneWidth = listWidth + 2 // Include left+right border
	m.detailPaneX = m.listPaneWidth
	m.contentY = statusBarHeight + tabBarHeight // Y position after status bar and tab bar
	m.contentHeight = contentHeight + 2         // Include top+bottom border

	m.alertList.SetSize(listWidth, contentHeight)
	m.hostList.SetSize(listWidth, contentHeight)
	m.eventList.SetSize(listWidth, contentHeight)
	m.graphList.SetSize(listWidth, contentHeight)
	m.detailPane.SetSize(detailWidth, contentHeight)
	m.statusBar.SetWidth(width)
	m.tabBar.SetWidth(width)
	m.commandInput.SetWidth(width)
	m.editorPane.SetScreenSize(width, height)
}

// Shutdown performs cleanup.
func (m *Model) Shutdown() {
	// Logout before cancelling context so the request can complete
	if m.client != nil {
		// Use a short timeout context for logout, not the cancelled one
		ctx, cancel := context.WithTimeout(context.Background(), LogoutTimeout*time.Second)
		defer cancel()
		_ = m.client.Logout(ctx)
	}
	m.cancel()
}

// getSelectedHostID returns the host ID for the currently selected item.
// Works across Alerts tab (returns the alert's host) and Hosts tab.
func (m *Model) getSelectedHostID() string {
	hostID, _ := m.getSelectedHostAndTriggerID()
	return hostID
}

// getSelectedHostAndTriggerID returns the host ID and trigger ID for the currently selected item.
// For alerts/events, returns the trigger that caused the problem.
// For hosts, returns just the host ID (trigger ID will be empty).
func (m *Model) getSelectedHostAndTriggerID() (hostID, triggerID string) {
	switch m.tabBar.Active() {
	case TabAlerts:
		if selected := m.alertList.Selected(); selected != nil && len(selected.Hosts) > 0 {
			hostID = selected.Hosts[0].HostID
			// ObjectID is the trigger ID for trigger-based problems
			if selected.Object == ObjectTypeTrigger {
				triggerID = selected.ObjectID
			}
			// Also check RelatedObject
			if triggerID == "" && selected.RelatedObject.TriggerID != "" {
				triggerID = selected.RelatedObject.TriggerID
			}
		}
	case TabHosts:
		if selected := m.hostList.Selected(); selected != nil {
			hostID = selected.HostID
		}
	case TabEvents:
		if selected := m.eventList.Selected(); selected != nil && len(selected.Hosts) > 0 {
			hostID = selected.Hosts[0].HostID
			// ObjectID is the trigger ID for trigger-based events
			if selected.Object == "0" { // 0 = trigger
				triggerID = selected.ObjectID
			}
			if triggerID == "" && selected.RelatedObject.TriggerID != "" {
				triggerID = selected.RelatedObject.TriggerID
			}
		}
	}
	return
}

// findHostByID finds a host by ID from various sources.
// Checks the hosts list first, then looks for embedded host info in alerts/events.
func (m *Model) findHostByID(hostID string) *zabbix.Host {
	// Check hosts list first (most complete data)
	for i := range m.hosts {
		if m.hosts[i].HostID == hostID {
			return &m.hosts[i]
		}
	}

	// Check current alert's embedded host
	if selected := m.alertList.Selected(); selected != nil {
		for i := range selected.Hosts {
			if selected.Hosts[i].HostID == hostID {
				return &selected.Hosts[i]
			}
		}
	}

	// Check current event's embedded host
	if selected := m.eventList.Selected(); selected != nil {
		for i := range selected.Hosts {
			if selected.Hosts[i].HostID == hostID {
				return &selected.Hosts[i]
			}
		}
	}

	return nil
}
