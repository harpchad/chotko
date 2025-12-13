package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/harpchad/chotko/internal/components/alerts"
	"github.com/harpchad/chotko/internal/components/command"
	"github.com/harpchad/chotko/internal/components/detail"
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
	errorModal modal.Model

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

// SetSize updates the window dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate pane sizes
	statusBarHeight := 1
	tabBarHeight := 1
	commandHeight := 1
	contentHeight := height - statusBarHeight - tabBarHeight - commandHeight - 4 // borders

	// Split width: 45% list, 55% detail
	// Account for borders: each pane has 2 chars (left+right border)
	availableWidth := width - 4 // 4 = 2 borders per pane * 2 panes
	listWidth := availableWidth * 45 / 100
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
}

// Shutdown performs cleanup.
func (m *Model) Shutdown() {
	m.cancel()
	if m.client != nil {
		_ = m.client.Logout(m.ctx)
	}
}
