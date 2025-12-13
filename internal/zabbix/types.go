package zabbix

import (
	"fmt"
	"strconv"
	"time"
)

// Problem represents a Zabbix problem/alert.
type Problem struct {
	EventID       string        `json:"eventid"`
	Source        string        `json:"source"`
	Object        string        `json:"object"`
	ObjectID      string        `json:"objectid"`
	Clock         string        `json:"clock"`
	NS            string        `json:"ns"`
	REventID      string        `json:"r_eventid"`
	RClock        string        `json:"r_clock"`
	Name          string        `json:"name"`
	Acknowledged  string        `json:"acknowledged"`
	Severity      string        `json:"severity"`
	Suppressed    string        `json:"suppressed"`
	OpData        string        `json:"opdata"`
	URLs          []URL         `json:"urls,omitempty"`
	Tags          []Tag         `json:"tags,omitempty"`
	Acknowledges  []Ack         `json:"acknowledges,omitempty"`
	Hosts         []Host        `json:"hosts,omitempty"`
	Triggers      []Trigger     `json:"triggers,omitempty"`
	RelatedObject RelatedObject `json:"relatedObject,omitempty"`
}

// RelatedObject represents the trigger/item that caused the event.
type RelatedObject struct {
	TriggerID string `json:"triggerid,omitempty"`
	Status    string `json:"status,omitempty"` // 0=enabled, 1=disabled
}

// URL represents a Zabbix URL associated with a problem.
type URL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Tag represents a Zabbix tag.
type Tag struct {
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

// Ack represents an acknowledgment on a problem.
type Ack struct {
	AckID       string `json:"acknowledgeid"`
	UserID      string `json:"userid"`
	EventID     string `json:"eventid"`
	Clock       string `json:"clock"`
	Message     string `json:"message"`
	Action      string `json:"action"`
	OldSeverity string `json:"old_severity"`
	NewSeverity string `json:"new_severity"`
	Username    string `json:"username,omitempty"`
	Name        string `json:"name,omitempty"`
	Surname     string `json:"surname,omitempty"`
}

// Host represents a Zabbix host.
type Host struct {
	HostID            string      `json:"hostid"`
	Host              string      `json:"host"`
	Name              string      `json:"name"`
	Status            string      `json:"status"`
	ProxyID           string      `json:"proxyid,omitempty"`
	MaintenanceStatus string      `json:"maintenance_status,omitempty"`
	MaintenanceType   string      `json:"maintenance_type,omitempty"`
	ActiveAvailable   string      `json:"active_available,omitempty"`
	Description       string      `json:"description,omitempty"`
	Interfaces        []Interface `json:"interfaces,omitempty"`
	Groups            []HostGroup `json:"groups,omitempty"`
	Macros            []HostMacro `json:"macros,omitempty"`
	Triggers          []Trigger   `json:"triggers,omitempty"`
}

// HostStatus constants.
const (
	HostStatusMonitored   = "0" // Host is monitored
	HostStatusUnmonitored = "1" // Host is not monitored
)

// HostMacro represents a user macro defined on a host.
type HostMacro struct {
	HostMacroID string `json:"hostmacroid,omitempty"`
	HostID      string `json:"hostid,omitempty"`
	Macro       string `json:"macro"`
	Value       string `json:"value"`
	Type        string `json:"type,omitempty"` // 0=text, 1=secret, 2=vault
	Description string `json:"description,omitempty"`
}

// MacroType constants.
const (
	MacroTypeText   = "0" // Plain text macro
	MacroTypeSecret = "1" // Secret macro (value hidden in UI)
	MacroTypeVault  = "2" // Vault macro
)

// Interface represents a host interface.
type Interface struct {
	InterfaceID string `json:"interfaceid"`
	IP          string `json:"ip"`
	DNS         string `json:"dns"`
	Port        string `json:"port"`
	Type        string `json:"type"`
	Main        string `json:"main"`
	Available   string `json:"available"`
}

// HostGroup represents a Zabbix host group.
type HostGroup struct {
	GroupID string `json:"groupid"`
	Name    string `json:"name"`
}

// Trigger represents a Zabbix trigger.
type Trigger struct {
	TriggerID   string `json:"triggerid"`
	Description string `json:"description"`
	Expression  string `json:"expression"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	Value       string `json:"value"`
	URL         string `json:"url,omitempty"`
	Comments    string `json:"comments,omitempty"`
}

// HostCounts represents aggregated host status counts.
type HostCounts struct {
	OK          int
	Problem     int
	Unknown     int
	Maintenance int
	Total       int
}

// Event represents a Zabbix event (problem or recovery).
// Uses the same structure as Problem since event.get returns the same fields.
type Event = Problem

// EventValue constants.
const (
	EventValueOK      = "0" // Recovery/OK event
	EventValueProblem = "1" // Problem event
)

// Helper methods

// SeverityInt returns the severity as an integer.
func (p *Problem) SeverityInt() int {
	s, _ := strconv.Atoi(p.Severity)
	return s
}

// IsAcknowledged returns true if the problem has been acknowledged.
func (p *Problem) IsAcknowledged() bool {
	return p.Acknowledged == "1"
}

// IsSuppressed returns true if the problem is suppressed.
func (p *Problem) IsSuppressed() bool {
	return p.Suppressed == "1"
}

// StartTime returns the problem start time.
func (p *Problem) StartTime() time.Time {
	ts, _ := strconv.ParseInt(p.Clock, 10, 64)
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

// Duration returns how long the problem has been active.
// Returns 0 if start time is invalid.
func (p *Problem) Duration() time.Duration {
	start := p.StartTime()
	if start.IsZero() {
		return 0
	}
	d := time.Since(start)
	// Sanity check - if duration is negative or unreasonably large, return 0
	if d < 0 {
		return 0
	}
	return d
}

// DurationString returns a human-readable duration string.
func (p *Problem) DurationString() string {
	d := p.Duration()
	if d <= 0 {
		return "-"
	}
	return formatDuration(d)
}

// formatDuration formats a duration as a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", h, m)
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

// HostName returns the first host name associated with this problem.
func (p *Problem) HostName() string {
	if len(p.Hosts) > 0 {
		if p.Hosts[0].Name != "" {
			return p.Hosts[0].Name
		}
		return p.Hosts[0].Host
	}
	return "Unknown"
}

// HostIP returns the first host IP associated with this problem.
func (p *Problem) HostIP() string {
	if len(p.Hosts) > 0 && len(p.Hosts[0].Interfaces) > 0 {
		return p.Hosts[0].Interfaces[0].IP
	}
	return ""
}

// IsMonitored returns true if the host is being monitored.
func (h *Host) IsMonitored() bool {
	return h.Status == "0"
}

// InMaintenance returns true if the host is in maintenance.
func (h *Host) InMaintenance() bool {
	return h.MaintenanceStatus == "1"
}

// IsAvailable returns the availability status based on all interfaces.
// Zabbix interface.available values: 0=unknown, 1=available, 2=unavailable
// Returns: 1=available (any interface OK), 2=unavailable (all interfaces down), 0=unknown (no data)
func (h *Host) IsAvailable() int {
	// First check interfaces - this is the most reliable method
	// as it covers all interface types (agent, SNMP, IPMI, JMX)
	if len(h.Interfaces) > 0 {
		hasAvailable := false
		hasUnavailable := false
		hasUnknown := false

		for _, iface := range h.Interfaces {
			switch iface.Available {
			case "1":
				hasAvailable = true
			case "2":
				hasUnavailable = true
			default:
				hasUnknown = true
			}
		}

		// If any interface is available, host is available
		if hasAvailable {
			return 1
		}
		// If any interface is unavailable (and none available), host is unavailable
		if hasUnavailable {
			return 2
		}
		// All interfaces are unknown
		if hasUnknown {
			return 0
		}
	}

	// Fallback to active_available for hosts without interface data
	// (shouldn't happen with proper selectInterfaces, but just in case)
	a, _ := strconv.Atoi(h.ActiveAvailable)
	return a
}

// DisplayName returns the visible name or falls back to technical name.
func (h *Host) DisplayName() string {
	if h.Name != "" {
		return h.Name
	}
	return h.Host
}

// IsRecovery returns true if this is a recovery (OK) event.
func (p *Problem) IsRecovery() bool {
	return p.REventID != "" && p.REventID != "0"
}

// RecoveryTime returns the recovery time if the problem was resolved.
func (p *Problem) RecoveryTime() time.Time {
	if p.RClock == "" || p.RClock == "0" {
		return time.Time{}
	}
	ts, _ := strconv.ParseInt(p.RClock, 10, 64)
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

// ResolvedDuration returns how long the problem lasted before being resolved.
// Returns 0 if not resolved or if times are invalid.
func (p *Problem) ResolvedDuration() time.Duration {
	if !p.IsRecovery() {
		return 0
	}
	start := p.StartTime()
	end := p.RecoveryTime()
	if start.IsZero() || end.IsZero() {
		return 0
	}
	d := end.Sub(start)
	// Sanity check - duration should be positive
	if d < 0 {
		return 0
	}
	return d
}

// ResolvedDurationString returns human-readable duration for resolved problems.
func (p *Problem) ResolvedDurationString() string {
	d := p.ResolvedDuration()
	if d <= 0 {
		return "-"
	}
	return formatDuration(d)
}

// Item represents a Zabbix item (metric).
type Item struct {
	ItemID    string `json:"itemid"`
	HostID    string `json:"hostid"`
	Name      string `json:"name"`
	Key       string `json:"key_"`
	ValueType string `json:"value_type"` // 0=float, 3=unsigned int (numeric)
	Units     string `json:"units"`
	LastValue string `json:"lastvalue"`
	LastClock string `json:"lastclock"`
	State     string `json:"state"`  // 0=normal, 1=not supported
	Status    string `json:"status"` // 0=enabled, 1=disabled
	Hosts     []Host `json:"hosts,omitempty"`
}

// ItemValueType constants for numeric items.
const (
	ItemValueTypeFloat    = "0" // Numeric (float)
	ItemValueTypeUnsigned = "3" // Numeric (unsigned)
)

// History represents a single history data point.
type History struct {
	ItemID string `json:"itemid"`
	Clock  string `json:"clock"`
	Value  string `json:"value"`
	NS     string `json:"ns"`
}

// Helper methods for Item

// IsNumeric returns true if the item has a numeric value type.
func (i *Item) IsNumeric() bool {
	return i.ValueType == ItemValueTypeFloat || i.ValueType == ItemValueTypeUnsigned
}

// IsEnabled returns true if the item is enabled.
func (i *Item) IsEnabled() bool {
	return i.Status == "0"
}

// IsSupported returns true if the item is in normal state (not unsupported).
func (i *Item) IsSupported() bool {
	return i.State == "0"
}

// LastValueFloat returns the last value as a float64.
func (i *Item) LastValueFloat() float64 {
	v, _ := strconv.ParseFloat(i.LastValue, 64)
	return v
}

// LastTime returns the last data collection time.
func (i *Item) LastTime() time.Time {
	ts, _ := strconv.ParseInt(i.LastClock, 10, 64)
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

// HostName returns the first host name associated with this item.
func (i *Item) HostName() string {
	if len(i.Hosts) > 0 {
		if i.Hosts[0].Name != "" {
			return i.Hosts[0].Name
		}
		return i.Hosts[0].Host
	}
	return "Unknown"
}

// GetHostID returns the host ID for this item.
func (i *Item) GetHostID() string {
	if len(i.Hosts) > 0 {
		return i.Hosts[0].HostID
	}
	return i.HostID
}

// Helper methods for History

// Time returns the history point timestamp.
func (h *History) Time() time.Time {
	ts, _ := strconv.ParseInt(h.Clock, 10, 64)
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

// ValueFloat returns the history value as a float64.
func (h *History) ValueFloat() float64 {
	v, _ := strconv.ParseFloat(h.Value, 64)
	return v
}
