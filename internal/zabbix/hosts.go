package zabbix

import (
	"context"
	"fmt"
)

// HostGetParams defines parameters for host.get API call.
type HostGetParams struct {
	// Output fields to return
	Output interface{} `json:"output,omitempty"`
	// Select interfaces
	SelectInterfaces interface{} `json:"selectInterfaces,omitempty"`
	// Select host groups
	SelectHostGroups interface{} `json:"selectHostGroups,omitempty"`
	// Select macros
	SelectMacros interface{} `json:"selectMacros,omitempty"`
	// Select triggers
	SelectTriggers interface{} `json:"selectTriggers,omitempty"`
	// Filter by host IDs
	HostIDs []string `json:"hostids,omitempty"`
	// Filter by group IDs
	GroupIDs []string `json:"groupids,omitempty"`
	// Filter by monitored hosts only
	MonitoredHosts bool `json:"monitored_hosts,omitempty"`
	// Filter to hosts with problems
	WithProblemsSuppressed *bool `json:"withProblemsSuppressed,omitempty"`
	// Sort field
	SortField []string `json:"sortfield,omitempty"`
	// Sort order
	SortOrder string `json:"sortorder,omitempty"`
	// Limit results
	Limit int `json:"limit,omitempty"`
	// Search filter
	Search map[string]string `json:"search,omitempty"`
	// Enable wildcard search
	SearchWildcardsEnabled bool `json:"searchWildcardsEnabled,omitempty"`
}

// DefaultHostGetParams returns default parameters for fetching hosts.
func DefaultHostGetParams() HostGetParams {
	return HostGetParams{
		Output:           []string{"hostid", "host", "name", "status", "maintenance_status", "active_available"},
		SelectInterfaces: []string{"interfaceid", "ip", "dns", "port", "type", "main", "available"},
		SelectHostGroups: []string{"groupid", "name"},
		MonitoredHosts:   true,
		SortField:        []string{"name"},
		SortOrder:        "ASC",
	}
}

// GetHosts retrieves hosts from Zabbix.
func (c *Client) GetHosts(ctx context.Context, params HostGetParams) ([]Host, error) {
	var hosts []Host
	if err := c.call(ctx, "host.get", params, &hosts); err != nil {
		return nil, fmt.Errorf("failed to get hosts: %w", err)
	}
	return hosts, nil
}

// GetAllHosts retrieves all monitored hosts.
func (c *Client) GetAllHosts(ctx context.Context) ([]Host, error) {
	params := DefaultHostGetParams()
	return c.GetHosts(ctx, params)
}

// GetHostCounts retrieves aggregated host status counts.
func (c *Client) GetHostCounts(ctx context.Context) (*HostCounts, error) {
	hosts, err := c.GetAllHosts(ctx)
	if err != nil {
		return nil, err
	}

	counts := &HostCounts{
		Total: len(hosts),
	}

	for _, h := range hosts {
		if h.InMaintenance() {
			counts.Maintenance++
			continue
		}

		// active_available values in Zabbix 7.x:
		// 0 = Unknown (no active agent data yet)
		// 1 = Available
		// 2 = Unavailable (agent not responding)
		switch h.IsAvailable() {
		case 1: // Available
			counts.OK++
		case 2: // Unavailable
			counts.Problem++
		case 0: // Unknown
			counts.Unknown++
		default:
			counts.Unknown++
		}
	}

	return counts, nil
}

// GetHost retrieves a single host by ID.
func (c *Client) GetHost(ctx context.Context, hostID string) (*Host, error) {
	params := DefaultHostGetParams()
	params.HostIDs = []string{hostID}

	hosts, err := c.GetHosts(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("host not found: %s", hostID)
	}

	return &hosts[0], nil
}

// SearchHosts searches for hosts by name pattern.
func (c *Client) SearchHosts(ctx context.Context, pattern string) ([]Host, error) {
	params := DefaultHostGetParams()
	params.Search = map[string]string{"name": pattern}
	params.SearchWildcardsEnabled = true

	return c.GetHosts(ctx, params)
}

// GetHostWithDetails retrieves a host with extended details including macros and triggers.
func (c *Client) GetHostWithDetails(ctx context.Context, hostID string) (*Host, error) {
	params := HostGetParams{
		Output:           "extend",
		SelectInterfaces: []string{"interfaceid", "ip", "dns", "port", "type", "main", "available"},
		SelectHostGroups: []string{"groupid", "name"},
		SelectMacros:     "extend",
		SelectTriggers:   []string{"triggerid", "description", "priority", "status", "value"},
		HostIDs:          []string{hostID},
	}

	hosts, err := c.GetHosts(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("host not found: %s", hostID)
	}

	return &hosts[0], nil
}

// HostUpdateParams defines parameters for host.update API call.
type HostUpdateParams struct {
	HostID      string      `json:"hostid"`
	Host        string      `json:"host,omitempty"`
	Name        string      `json:"name,omitempty"`
	Status      string      `json:"status,omitempty"`
	Description string      `json:"description,omitempty"`
	Macros      []HostMacro `json:"macros,omitempty"`
}

// HostUpdateResult represents the result of a host.update API call.
type HostUpdateResult struct {
	HostIDs []string `json:"hostids"`
}

// UpdateHost updates a host with the given parameters.
func (c *Client) UpdateHost(ctx context.Context, params HostUpdateParams) error {
	var result HostUpdateResult
	if err := c.call(ctx, "host.update", params, &result); err != nil {
		return fmt.Errorf("failed to update host: %w", err)
	}
	return nil
}

// EnableHost enables monitoring for a host.
func (c *Client) EnableHost(ctx context.Context, hostID string) error {
	params := HostUpdateParams{
		HostID: hostID,
		Status: HostStatusMonitored,
	}
	return c.UpdateHost(ctx, params)
}

// DisableHost disables monitoring for a host.
func (c *Client) DisableHost(ctx context.Context, hostID string) error {
	params := HostUpdateParams{
		HostID: hostID,
		Status: HostStatusUnmonitored,
	}
	return c.UpdateHost(ctx, params)
}

// SetHostDescription updates the host description.
func (c *Client) SetHostDescription(ctx context.Context, hostID, description string) error {
	params := HostUpdateParams{
		HostID:      hostID,
		Description: description,
	}
	return c.UpdateHost(ctx, params)
}

// SetHostMacros updates the macros for a host.
// Note: This replaces ALL macros on the host. To add/update a single macro,
// first get existing macros, modify the list, then call this.
func (c *Client) SetHostMacros(ctx context.Context, hostID string, macros []HostMacro) error {
	params := HostUpdateParams{
		HostID: hostID,
		Macros: macros,
	}
	return c.UpdateHost(ctx, params)
}

// UserMacroGetParams defines parameters for usermacro.get API call.
type UserMacroGetParams struct {
	Output  interface{} `json:"output,omitempty"`
	HostIDs []string    `json:"hostids,omitempty"`
}

// GetHostMacros retrieves all macros for a specific host.
func (c *Client) GetHostMacros(ctx context.Context, hostID string) ([]HostMacro, error) {
	params := UserMacroGetParams{
		Output:  "extend",
		HostIDs: []string{hostID},
	}

	var macros []HostMacro
	if err := c.call(ctx, "usermacro.get", params, &macros); err != nil {
		return nil, fmt.Errorf("failed to get host macros: %w", err)
	}
	return macros, nil
}

// UserMacroCreateParams defines parameters for usermacro.create API call.
type UserMacroCreateParams struct {
	HostID      string `json:"hostid"`
	Macro       string `json:"macro"`
	Value       string `json:"value"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

// UserMacroCreateResult represents the result of usermacro.create.
type UserMacroCreateResult struct {
	HostMacroIDs []string `json:"hostmacroids"`
}

// CreateHostMacro creates a new macro on a host.
func (c *Client) CreateHostMacro(ctx context.Context, hostID, macro, value string) error {
	params := UserMacroCreateParams{
		HostID: hostID,
		Macro:  macro,
		Value:  value,
	}

	var result UserMacroCreateResult
	if err := c.call(ctx, "usermacro.create", params, &result); err != nil {
		return fmt.Errorf("failed to create host macro: %w", err)
	}
	return nil
}

// UserMacroUpdateParams defines parameters for usermacro.update API call.
type UserMacroUpdateParams struct {
	HostMacroID string `json:"hostmacroid"`
	Macro       string `json:"macro,omitempty"`
	Value       string `json:"value,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

// UserMacroUpdateResult represents the result of usermacro.update.
type UserMacroUpdateResult struct {
	HostMacroIDs []string `json:"hostmacroids"`
}

// UpdateHostMacro updates an existing macro.
func (c *Client) UpdateHostMacro(ctx context.Context, macroID, value string) error {
	params := UserMacroUpdateParams{
		HostMacroID: macroID,
		Value:       value,
	}

	var result UserMacroUpdateResult
	if err := c.call(ctx, "usermacro.update", params, &result); err != nil {
		return fmt.Errorf("failed to update host macro: %w", err)
	}
	return nil
}

// DeleteHostMacro deletes a macro.
func (c *Client) DeleteHostMacro(ctx context.Context, macroID string) error {
	// usermacro.delete takes an array of macro IDs
	params := []string{macroID}

	var result UserMacroUpdateResult
	if err := c.call(ctx, "usermacro.delete", params, &result); err != nil {
		return fmt.Errorf("failed to delete host macro: %w", err)
	}
	return nil
}
