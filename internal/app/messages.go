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

// ConfigLoadedMsg is sent when configuration is loaded.
type ConfigLoadedMsg struct {
	Success bool
	Err     error
}

// ConnectedMsg is sent when successfully connected to Zabbix.
type ConnectedMsg struct {
	Version string
	Client  *zabbix.Client
}

// DisconnectedMsg is sent when disconnected from Zabbix.
type DisconnectedMsg struct {
	Err error
}

// FilterChangedMsg is sent when the filter changes.
type FilterChangedMsg struct {
	MinSeverity int
	TextFilter  string
}

// CommandExecutedMsg is sent after a command is executed.
type CommandExecutedMsg struct {
	Command string
	Result  string
	Err     error
}

// WizardCompleteMsg is sent when the setup wizard completes.
type WizardCompleteMsg struct {
	Success bool
}
