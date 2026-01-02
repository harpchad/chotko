package zabbix

import (
	"context"
	"fmt"
	"time"
)

// ItemGetParams defines parameters for item.get API call.
type ItemGetParams struct {
	// Output fields to return
	Output interface{} `json:"output,omitempty"`
	// Select hosts
	SelectHosts interface{} `json:"selectHosts,omitempty"`
	// Filter by host IDs
	HostIDs []string `json:"hostids,omitempty"`
	// Filter by item IDs
	ItemIDs []string `json:"itemids,omitempty"`
	// Filter by group IDs
	GroupIDs []string `json:"groupids,omitempty"`
	// Filter by key pattern (supports wildcards with searchWildcardsEnabled)
	Search map[string]string `json:"search,omitempty"`
	// Enable wildcard search
	SearchWildcardsEnabled bool `json:"searchWildcardsEnabled,omitempty"`
	// Filter to monitored items only
	Monitored bool `json:"monitored,omitempty"`
	// Filter by value types (0=float, 3=unsigned for numeric)
	Filter map[string]interface{} `json:"filter,omitempty"`
	// Sort field
	SortField []string `json:"sortfield,omitempty"`
	// Sort order
	SortOrder string `json:"sortorder,omitempty"`
	// Limit results
	Limit int `json:"limit,omitempty"`
	// Include suppressed items
	WebtItems bool `json:"webitems,omitempty"`
}

// DefaultItemGetParams returns default parameters for fetching items.
func DefaultItemGetParams() ItemGetParams {
	return ItemGetParams{
		Output:      []string{"itemid", "hostid", "name", "key_", "value_type", "units", "lastvalue", "lastclock", "state", "status"},
		SelectHosts: []string{"hostid", "host", "name"},
		Monitored:   true,
		SortField:   []string{"name"},
		SortOrder:   "ASC",
	}
}

// GetItems retrieves items from Zabbix.
func (c *Client) GetItems(ctx context.Context, params ItemGetParams) ([]Item, error) {
	var items []Item
	if err := c.call(ctx, "item.get", params, &items); err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	return items, nil
}

// GetNumericItems retrieves numeric items (float and unsigned int) for the given hosts.
// This is used for the graphs tab to fetch items that can be charted.
func (c *Client) GetNumericItems(ctx context.Context, hostIDs, keyPrefixes []string) ([]Item, error) {
	params := DefaultItemGetParams()
	params.HostIDs = hostIDs
	// Filter to numeric value types only (0=float, 3=unsigned int)
	params.Filter = map[string]interface{}{
		"value_type": []string{ItemValueTypeFloat, ItemValueTypeUnsigned},
		"status":     "0", // enabled items only
	}

	items, err := c.GetItems(ctx, params)
	if err != nil {
		return nil, err
	}

	// If key prefixes specified, filter items by key prefix
	if len(keyPrefixes) > 0 {
		filtered := make([]Item, 0, len(items))
		for _, item := range items {
			for _, prefix := range keyPrefixes {
				if matchKeyPrefix(item.Key, prefix) {
					filtered = append(filtered, item)
					break
				}
			}
		}
		return filtered, nil
	}

	return items, nil
}

// GetAllNumericItems retrieves all numeric items from all hosts.
// Used when loading items for the graphs tab.
func (c *Client) GetAllNumericItems(ctx context.Context, keyPrefixes []string) ([]Item, error) {
	return c.GetNumericItems(ctx, nil, keyPrefixes)
}

// matchKeyPrefix checks if an item key matches a prefix pattern.
// Supports basic prefix matching (e.g., "system.cpu" matches "system.cpu.util")
func matchKeyPrefix(key, prefix string) bool {
	if len(key) < len(prefix) {
		return false
	}
	return key[:len(prefix)] == prefix
}

// HistoryGetParams defines parameters for history.get API call.
type HistoryGetParams struct {
	// History type (0=float, 3=unsigned int, matches item value_type)
	History int `json:"history"`
	// Item IDs to get history for
	ItemIDs []string `json:"itemids"`
	// Unix timestamp - get history from this time
	TimeFrom int64 `json:"time_from,omitempty"`
	// Unix timestamp - get history until this time
	TimeTill int64 `json:"time_till,omitempty"`
	// Output fields
	Output interface{} `json:"output,omitempty"`
	// Sort field
	SortField []string `json:"sortfield,omitempty"`
	// Sort order
	SortOrder string `json:"sortorder,omitempty"`
	// Limit results
	Limit int `json:"limit,omitempty"`
}

// GetHistory retrieves history data for items.
func (c *Client) GetHistory(ctx context.Context, params HistoryGetParams) ([]History, error) {
	if params.Output == nil {
		params.Output = "extend"
	}

	var history []History
	if err := c.call(ctx, "history.get", params, &history); err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	return history, nil
}

// GetItemHistory retrieves history for a single item over a time range.
func (c *Client) GetItemHistory(ctx context.Context, itemID, valueType string, hours int) ([]History, error) {
	// Determine history type based on value_type
	historyType := 0 // float by default
	if valueType == ItemValueTypeUnsigned {
		historyType = 3
	}

	now := time.Now()
	params := HistoryGetParams{
		History:   historyType,
		ItemIDs:   []string{itemID},
		TimeFrom:  now.Add(-time.Duration(hours) * time.Hour).Unix(),
		TimeTill:  now.Unix(),
		Output:    "extend",
		SortField: []string{"clock"},
		SortOrder: "ASC",
	}

	return c.GetHistory(ctx, params)
}

// GetItemsHistory retrieves history for multiple items over a time range.
// Returns a map of itemID -> []History
func (c *Client) GetItemsHistory(ctx context.Context, items []Item, hours int) (map[string][]History, error) {
	result := make(map[string][]History)

	// Group items by value type to make efficient API calls
	floatItems := make([]string, 0)
	unsignedItems := make([]string, 0)

	for _, item := range items {
		switch item.ValueType {
		case ItemValueTypeFloat:
			floatItems = append(floatItems, item.ItemID)
		case ItemValueTypeUnsigned:
			unsignedItems = append(unsignedItems, item.ItemID)
		}
	}

	now := time.Now()
	timeFrom := now.Add(-time.Duration(hours) * time.Hour).Unix()
	timeTill := now.Unix()

	// Fetch float history
	if len(floatItems) > 0 {
		params := HistoryGetParams{
			History:   0, // float
			ItemIDs:   floatItems,
			TimeFrom:  timeFrom,
			TimeTill:  timeTill,
			Output:    "extend",
			SortField: []string{"clock"},
			SortOrder: "ASC",
		}

		history, err := c.GetHistory(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get float history: %w", err)
		}

		for _, h := range history {
			result[h.ItemID] = append(result[h.ItemID], h)
		}
	}

	// Fetch unsigned history
	if len(unsignedItems) > 0 {
		params := HistoryGetParams{
			History:   3, // unsigned
			ItemIDs:   unsignedItems,
			TimeFrom:  timeFrom,
			TimeTill:  timeTill,
			Output:    "extend",
			SortField: []string{"clock"},
			SortOrder: "ASC",
		}

		history, err := c.GetHistory(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get unsigned history: %w", err)
		}

		for _, h := range history {
			result[h.ItemID] = append(result[h.ItemID], h)
		}
	}

	return result, nil
}
