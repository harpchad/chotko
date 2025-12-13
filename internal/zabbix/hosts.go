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
