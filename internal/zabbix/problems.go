package zabbix

import (
	"context"
	"fmt"
	"time"
)

// ProblemGetParams defines parameters for fetching problems.
type ProblemGetParams struct {
	Severities []int
	Limit      int
}

// DefaultProblemGetParams returns default parameters for fetching active problems.
func DefaultProblemGetParams() ProblemGetParams {
	return ProblemGetParams{}
}

// internalProblemGetParams defines parameters for problem.get API call.
// Used internally to get current active problems (problem.get only returns unresolved problems).
type internalProblemGetParams struct {
	Output             interface{} `json:"output,omitempty"`
	SelectTags         interface{} `json:"selectTags,omitempty"`
	SelectAcknowledges interface{} `json:"selectAcknowledges,omitempty"`
	Severities         []int       `json:"severities,omitempty"`
	SortField          []string    `json:"sortfield,omitempty"`
	SortOrder          string      `json:"sortorder,omitempty"`
	Limit              int         `json:"limit,omitempty"`
}

// EventGetParams defines parameters for event.get API call.
// We use event.get to get host information for problems.
// See: https://www.zabbix.com/documentation/7.0/en/manual/api/reference/event/get
type EventGetParams struct {
	Output              interface{} `json:"output,omitempty"`
	SelectHosts         interface{} `json:"selectHosts,omitempty"`
	SelectTags          interface{} `json:"selectTags,omitempty"`
	SelectAcknowledges  interface{} `json:"selectAcknowledges,omitempty"`
	SelectRelatedObject interface{} `json:"selectRelatedObject,omitempty"`
	EventIDs            []string    `json:"eventids,omitempty"`
	HostIDs             []string    `json:"hostids,omitempty"`
	GroupIDs            []string    `json:"groupids,omitempty"`
	ObjectIDs           []string    `json:"objectids,omitempty"`
	Severities          []int       `json:"severities,omitempty"`
	Value               []int       `json:"value,omitempty"`  // 0 = OK, 1 = problem (can be array)
	Source              *int        `json:"source,omitempty"` // 0 = trigger (single int, use pointer to omit when nil)
	Object              *int        `json:"object,omitempty"` // 0 = trigger (single int, use pointer to omit when nil)
	SortField           []string    `json:"sortfield,omitempty"`
	SortOrder           string      `json:"sortorder,omitempty"`
	Limit               int         `json:"limit,omitempty"`
	TimeFrom            int64       `json:"time_from,omitempty"`
	TimeTill            int64       `json:"time_till,omitempty"`
}

// GetProblems retrieves current active problems from Zabbix.
// Uses a two-step approach:
// 1. problem.get to get current active problem eventids (problem.get only returns unresolved problems)
// 2. event.get with those eventids to get host information (selectHosts not supported in problem.get)
func (c *Client) GetProblems(ctx context.Context, params ProblemGetParams) ([]Problem, error) {
	// Step 1: Get current active problems from problem.get
	problemParams := internalProblemGetParams{
		Output:             []string{"eventid"}, // Only need eventids for step 2
		SelectTags:         "extend",
		SelectAcknowledges: "extend",
		Severities:         params.Severities,
		SortField:          []string{"eventid"},
		SortOrder:          "DESC",
	}

	if params.Limit > 0 {
		problemParams.Limit = params.Limit
	}

	var problemResults []struct {
		EventID string `json:"eventid"`
	}
	if err := c.call(ctx, "problem.get", problemParams, &problemResults); err != nil {
		return nil, fmt.Errorf("failed to get active problems: %w", err)
	}

	// If no problems, return empty slice
	if len(problemResults) == 0 {
		return []Problem{}, nil
	}

	// Extract eventids
	eventIDs := make([]string, len(problemResults))
	for i, p := range problemResults {
		eventIDs[i] = p.EventID
	}

	// Step 2: Get full event details with hosts and trigger status
	eventParams := EventGetParams{
		Output:              "extend",
		SelectHosts:         []string{"hostid", "host", "name"},
		SelectTags:          "extend",
		SelectAcknowledges:  "extend",
		SelectRelatedObject: []string{"triggerid", "status"},
		EventIDs:            eventIDs,
		SortField:           []string{"eventid"},
		SortOrder:           "DESC",
	}

	var problems []Problem
	if err := c.call(ctx, "event.get", eventParams, &problems); err != nil {
		return nil, fmt.Errorf("failed to get problem details: %w", err)
	}

	// Filter out problems where the trigger is disabled (status=1)
	// This matches Zabbix web UI behavior which hides disabled trigger problems
	filtered := make([]Problem, 0, len(problems))
	for _, p := range problems {
		if p.RelatedObject.Status != "1" {
			filtered = append(filtered, p)
		}
	}

	return filtered, nil
}

// GetActiveProblems retrieves all active (unresolved) problems.
func (c *Client) GetActiveProblems(ctx context.Context) ([]Problem, error) {
	params := DefaultProblemGetParams()
	return c.GetProblems(ctx, params)
}

// GetProblemsWithMinSeverity retrieves problems with at least the given severity.
func (c *Client) GetProblemsWithMinSeverity(ctx context.Context, minSeverity int) ([]Problem, error) {
	params := DefaultProblemGetParams()

	// Build severity list from minSeverity to 5
	severities := make([]int, 0)
	for s := minSeverity; s <= 5; s++ {
		severities = append(severities, s)
	}
	params.Severities = severities

	return c.GetProblems(ctx, params)
}

// AcknowledgeParams defines parameters for event.acknowledge API call.
type AcknowledgeParams struct {
	EventIDs []string `json:"eventids"`
	Action   int      `json:"action"`
	Message  string   `json:"message,omitempty"`
	Severity int      `json:"severity,omitempty"`
}

// AcknowledgeAction constants for the action bitmask.
const (
	ActionClose          = 1
	ActionAcknowledge    = 2
	ActionAddMessage     = 4
	ActionChangeSeverity = 8
	ActionUnacknowledge  = 16
	ActionSuppress       = 32
	ActionUnsuppress     = 64
)

// AcknowledgeProblem acknowledges a problem event.
func (c *Client) AcknowledgeProblem(ctx context.Context, eventID string, message string) error {
	action := ActionAcknowledge
	if message != "" {
		action |= ActionAddMessage
	}

	params := AcknowledgeParams{
		EventIDs: []string{eventID},
		Action:   action,
		Message:  message,
	}

	// Result contains eventids but the type varies by Zabbix version (string or number)
	// We don't need the result, just check if the call succeeded
	var result interface{}
	if err := c.call(ctx, "event.acknowledge", params, &result); err != nil {
		return fmt.Errorf("failed to acknowledge problem: %w", err)
	}

	return nil
}

// AcknowledgeProblems acknowledges multiple problem events.
func (c *Client) AcknowledgeProblems(ctx context.Context, eventIDs []string, message string) error {
	action := ActionAcknowledge
	if message != "" {
		action |= ActionAddMessage
	}

	params := AcknowledgeParams{
		EventIDs: eventIDs,
		Action:   action,
		Message:  message,
	}

	// Result contains eventids but the type varies by Zabbix version (string or number)
	// We don't need the result, just check if the call succeeded
	var result interface{}
	if err := c.call(ctx, "event.acknowledge", params, &result); err != nil {
		return fmt.Errorf("failed to acknowledge problems: %w", err)
	}

	return nil
}

// CloseProblem closes a problem (marks as resolved manually).
func (c *Client) CloseProblem(ctx context.Context, eventID string, message string) error {
	action := ActionClose
	if message != "" {
		action |= ActionAddMessage
	}

	params := AcknowledgeParams{
		EventIDs: []string{eventID},
		Action:   action,
		Message:  message,
	}

	// Result contains eventids but the type varies by Zabbix version (string or number)
	// We don't need the result, just check if the call succeeded
	var result interface{}
	if err := c.call(ctx, "event.acknowledge", params, &result); err != nil {
		return fmt.Errorf("failed to close problem: %w", err)
	}

	return nil
}

// SuppressProblem suppresses a problem event.
func (c *Client) SuppressProblem(ctx context.Context, eventID string) error {
	params := AcknowledgeParams{
		EventIDs: []string{eventID},
		Action:   ActionSuppress,
	}

	// Result contains eventids but the type varies by Zabbix version (string or number)
	// We don't need the result, just check if the call succeeded
	var result interface{}
	if err := c.call(ctx, "event.acknowledge", params, &result); err != nil {
		return fmt.Errorf("failed to suppress problem: %w", err)
	}

	return nil
}

// UnsuppressProblem unsuppresses a problem event.
func (c *Client) UnsuppressProblem(ctx context.Context, eventID string) error {
	params := AcknowledgeParams{
		EventIDs: []string{eventID},
		Action:   ActionUnsuppress,
	}

	// Result contains eventids but the type varies by Zabbix version (string or number)
	// We don't need the result, just check if the call succeeded
	var result interface{}
	if err := c.call(ctx, "event.acknowledge", params, &result); err != nil {
		return fmt.Errorf("failed to unsuppress problem: %w", err)
	}

	return nil
}

// EventHistoryParams defines parameters for fetching event history.
type EventHistoryParams struct {
	Limit    int   // Max events to return (default 100)
	TimeFrom int64 // Unix timestamp - events from this time
	TimeTill int64 // Unix timestamp - events until this time
	HostIDs  []string
}

// DefaultEventHistoryParams returns default parameters for event history.
func DefaultEventHistoryParams() EventHistoryParams {
	return EventHistoryParams{
		Limit: 100,
	}
}

// GetEventHistory retrieves recent events (both problem and recovery).
// This shows historical events, not just active problems.
func (c *Client) GetEventHistory(ctx context.Context, params EventHistoryParams) ([]Event, error) {
	source := 0 // 0 = trigger events
	object := 0 // 0 = trigger

	eventParams := EventGetParams{
		Output:             "extend",
		SelectHosts:        []string{"hostid", "host", "name"},
		SelectTags:         "extend",
		SelectAcknowledges: "extend",
		Source:             &source,
		Object:             &object,
		SortField:          []string{"clock", "eventid"},
		SortOrder:          "DESC",
		Limit:              params.Limit,
	}

	if params.Limit == 0 {
		eventParams.Limit = 100
	}

	if params.TimeFrom > 0 {
		eventParams.TimeFrom = params.TimeFrom
	}

	if params.TimeTill > 0 {
		eventParams.TimeTill = params.TimeTill
	}

	if len(params.HostIDs) > 0 {
		eventParams.HostIDs = params.HostIDs
	}

	var events []Event
	if err := c.call(ctx, "event.get", eventParams, &events); err != nil {
		return nil, fmt.Errorf("failed to get event history: %w", err)
	}

	return events, nil
}

// GetRecentEvents retrieves events from the last N hours.
func (c *Client) GetRecentEvents(ctx context.Context, hours int, limit int) ([]Event, error) {
	params := DefaultEventHistoryParams()
	params.TimeFrom = time.Now().Add(-time.Duration(hours) * time.Hour).Unix()
	if limit > 0 {
		params.Limit = limit
	}
	return c.GetEventHistory(ctx, params)
}
