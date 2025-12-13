package zabbix

import (
	"context"
	"fmt"
)

// TriggerStatus constants.
const (
	TriggerStatusEnabled  = "0" // Trigger is enabled
	TriggerStatusDisabled = "1" // Trigger is disabled
)

// TriggerGetParams defines parameters for trigger.get API call.
type TriggerGetParams struct {
	// Output fields to return
	Output interface{} `json:"output,omitempty"`
	// Select hosts
	SelectHosts interface{} `json:"selectHosts,omitempty"`
	// Select tags
	SelectTags interface{} `json:"selectTags,omitempty"`
	// Filter by trigger IDs
	TriggerIDs []string `json:"triggerids,omitempty"`
	// Filter by host IDs
	HostIDs []string `json:"hostids,omitempty"`
	// Filter by group IDs
	GroupIDs []string `json:"groupids,omitempty"`
	// Only enabled triggers
	Active bool `json:"active,omitempty"`
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
	// Filter parameters
	Filter map[string]interface{} `json:"filter,omitempty"`
}

// DefaultTriggerGetParams returns default parameters for fetching triggers.
func DefaultTriggerGetParams() TriggerGetParams {
	return TriggerGetParams{
		Output:      "extend",
		SelectHosts: []string{"hostid", "host", "name"},
		SelectTags:  "extend",
		SortField:   []string{"description"},
		SortOrder:   "ASC",
	}
}

// GetTriggers retrieves triggers from Zabbix.
func (c *Client) GetTriggers(ctx context.Context, params TriggerGetParams) ([]Trigger, error) {
	var triggers []Trigger
	if err := c.call(ctx, "trigger.get", params, &triggers); err != nil {
		return nil, fmt.Errorf("failed to get triggers: %w", err)
	}
	return triggers, nil
}

// GetTrigger retrieves a single trigger by ID.
func (c *Client) GetTrigger(ctx context.Context, triggerID string) (*Trigger, error) {
	params := DefaultTriggerGetParams()
	params.TriggerIDs = []string{triggerID}

	triggers, err := c.GetTriggers(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(triggers) == 0 {
		return nil, fmt.Errorf("trigger not found: %s", triggerID)
	}

	return &triggers[0], nil
}

// GetHostTriggers retrieves all triggers for a specific host.
func (c *Client) GetHostTriggers(ctx context.Context, hostID string) ([]Trigger, error) {
	params := DefaultTriggerGetParams()
	params.HostIDs = []string{hostID}

	return c.GetTriggers(ctx, params)
}

// TriggerUpdateParams defines parameters for trigger.update API call.
type TriggerUpdateParams struct {
	TriggerID   string `json:"triggerid"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority,omitempty"`
	Comments    string `json:"comments,omitempty"`
	URL         string `json:"url,omitempty"`
}

// TriggerUpdateResult represents the result of a trigger.update API call.
type TriggerUpdateResult struct {
	TriggerIDs []string `json:"triggerids"`
}

// UpdateTrigger updates a trigger with the given parameters.
func (c *Client) UpdateTrigger(ctx context.Context, params TriggerUpdateParams) error {
	var result TriggerUpdateResult
	if err := c.call(ctx, "trigger.update", params, &result); err != nil {
		return fmt.Errorf("failed to update trigger: %w", err)
	}
	return nil
}

// EnableTrigger enables a trigger.
func (c *Client) EnableTrigger(ctx context.Context, triggerID string) error {
	params := TriggerUpdateParams{
		TriggerID: triggerID,
		Status:    TriggerStatusEnabled,
	}
	return c.UpdateTrigger(ctx, params)
}

// DisableTrigger disables a trigger.
func (c *Client) DisableTrigger(ctx context.Context, triggerID string) error {
	params := TriggerUpdateParams{
		TriggerID: triggerID,
		Status:    TriggerStatusDisabled,
	}
	return c.UpdateTrigger(ctx, params)
}

// EnableTriggers enables multiple triggers.
func (c *Client) EnableTriggers(ctx context.Context, triggerIDs []string) error {
	for _, id := range triggerIDs {
		if err := c.EnableTrigger(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// DisableTriggers disables multiple triggers.
func (c *Client) DisableTriggers(ctx context.Context, triggerIDs []string) error {
	for _, id := range triggerIDs {
		if err := c.DisableTrigger(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// SetTriggerPriority sets the priority (severity) of a trigger.
// Priority values: 0=Not classified, 1=Information, 2=Warning, 3=Average, 4=High, 5=Disaster
func (c *Client) SetTriggerPriority(ctx context.Context, triggerID string, priority int) error {
	params := TriggerUpdateParams{
		TriggerID: triggerID,
		Priority:  fmt.Sprintf("%d", priority),
	}
	return c.UpdateTrigger(ctx, params)
}

// Helper methods for Trigger

// IsEnabled returns true if the trigger is enabled.
func (t *Trigger) IsEnabled() bool {
	return t.Status == TriggerStatusEnabled
}

// IsDisabled returns true if the trigger is disabled.
func (t *Trigger) IsDisabled() bool {
	return t.Status == TriggerStatusDisabled
}

// IsProblem returns true if the trigger is in problem state.
func (t *Trigger) IsProblem() bool {
	return t.Value == "1"
}

// PriorityInt returns the priority as an integer.
func (t *Trigger) PriorityInt() int {
	p := 0
	_, _ = fmt.Sscanf(t.Priority, "%d", &p)
	return p
}
