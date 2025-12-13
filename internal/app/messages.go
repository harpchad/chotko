package app

import (
	"github.com/harpchad/chotko/internal/zabbix"
)

// Custom message types for the application

// ProblemsLoadedMsg is sent when problems are loaded from Zabbix.
type ProblemsLoadedMsg struct {
	Problems []zabbix.Problem
	Err      error
}

// HostCountsLoadedMsg is sent when host counts are loaded from Zabbix.
type HostCountsLoadedMsg struct {
	Counts *zabbix.HostCounts
	Err    error
}

// HostsLoadedMsg is sent when hosts are loaded from Zabbix.
type HostsLoadedMsg struct {
	Hosts []zabbix.Host
	Err   error
}

// EventsLoadedMsg is sent when events are loaded from Zabbix.
type EventsLoadedMsg struct {
	Events []zabbix.Event
	Err    error
}

// AcknowledgeResultMsg is sent after acknowledging a problem.
type AcknowledgeResultMsg struct {
	EventID string
	Success bool
	Err     error
}

// RefreshTickMsg is sent periodically to trigger data refresh.
type RefreshTickMsg struct{}

// ErrorMsg represents an error to be displayed to the user.
type ErrorMsg struct {
	Title   string
	Message string
	Err     error
}

// ClearErrorMsg clears the current error display.
type ClearErrorMsg struct{}

// ConnectedMsg is sent when successfully connected to Zabbix.
type ConnectedMsg struct {
	Version string
	Client  *zabbix.Client
}

// DisconnectedMsg is sent when disconnected from Zabbix.
type DisconnectedMsg struct {
	Err error
}

// ItemsLoadedMsg is sent when items are loaded from Zabbix.
type ItemsLoadedMsg struct {
	Items []zabbix.Item
	Err   error
}

// HostHistoryLoadedMsg is sent when history for a specific host is loaded.
type HostHistoryLoadedMsg struct {
	HostID  string
	History map[string][]zabbix.History
	Err     error
}

// HostTriggersLoadedMsg is sent when triggers for a host are loaded.
type HostTriggersLoadedMsg struct {
	HostID          string
	Triggers        []zabbix.Trigger
	SelectTriggerID string // Optional: pre-select this trigger
	Err             error
}

// HostMacrosLoadedMsg is sent when macros for a host are loaded.
type HostMacrosLoadedMsg struct {
	HostID string
	Macros []zabbix.HostMacro
	Err    error
}

// HostUpdateResultMsg is sent after a host update operation.
type HostUpdateResultMsg struct {
	HostID  string
	Action  string // "enable", "disable", "update"
	Success bool
	Err     error
}

// TriggerUpdateResultMsg is sent after a trigger update operation.
type TriggerUpdateResultMsg struct {
	TriggerID string
	Action    string // "enable", "disable", "update"
	Success   bool
	Err       error
}

// MacroUpdateResultMsg is sent after a macro update operation.
type MacroUpdateResultMsg struct {
	MacroID string
	Action  string // "create", "update", "delete"
	Success bool
	Err     error
}
