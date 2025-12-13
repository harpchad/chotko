package zabbix

import (
	"context"
	"testing"
)

func TestDefaultTriggerGetParams(t *testing.T) {
	params := DefaultTriggerGetParams()

	if params.Output == nil {
		t.Error("Output not set in default params")
	}

	if params.SelectHosts == nil {
		t.Error("SelectHosts not set in default params")
	}

	if params.SelectTags == nil {
		t.Error("SelectTags not set in default params")
	}

	if len(params.SortField) == 0 {
		t.Error("SortField not set in default params")
	}
}

func TestClient_GetTriggers(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.get": {
			Result: []Trigger{
				{TriggerID: "1", Description: "Trigger 1", Status: "0", Priority: "3"},
				{TriggerID: "2", Description: "Trigger 2", Status: "1", Priority: "4"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	triggers, err := client.GetTriggers(context.Background(), DefaultTriggerGetParams())
	if err != nil {
		t.Fatalf("GetTriggers() error = %v", err)
	}

	if len(triggers) != 2 {
		t.Errorf("len(triggers) = %d, want 2", len(triggers))
	}

	if triggers[0].Description != "Trigger 1" {
		t.Errorf("triggers[0].Description = %q, want %q", triggers[0].Description, "Trigger 1")
	}
}

func TestClient_GetTriggers_Empty(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.get": {Result: []Trigger{}},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	triggers, err := client.GetTriggers(context.Background(), DefaultTriggerGetParams())
	if err != nil {
		t.Fatalf("GetTriggers() error = %v", err)
	}

	if len(triggers) != 0 {
		t.Errorf("len(triggers) = %d, want 0", len(triggers))
	}
}

func TestClient_GetTriggers_Error(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.get": {
			Error: &APIError{
				Code:    -32602,
				Message: "Invalid params",
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	_, err := client.GetTriggers(context.Background(), DefaultTriggerGetParams())
	if err == nil {
		t.Fatal("GetTriggers() expected error")
	}
}

func TestClient_GetHostTriggers(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.get": {
			Result: []Trigger{
				{TriggerID: "1", Description: "CPU high", Status: "0", Priority: "4"},
				{TriggerID: "2", Description: "Memory low", Status: "0", Priority: "3"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	triggers, err := client.GetHostTriggers(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetHostTriggers() error = %v", err)
	}

	if len(triggers) != 2 {
		t.Errorf("len(triggers) = %d, want 2", len(triggers))
	}
}

func TestClient_GetTrigger(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.get": {
			Result: []Trigger{
				{TriggerID: "456", Description: "Test Trigger", Status: "0"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	trigger, err := client.GetTrigger(context.Background(), "456")
	if err != nil {
		t.Fatalf("GetTrigger() error = %v", err)
	}

	if trigger == nil {
		t.Fatal("GetTrigger() returned nil")
	}

	if trigger.TriggerID != "456" {
		t.Errorf("trigger.TriggerID = %q, want %q", trigger.TriggerID, "456")
	}
}

func TestClient_GetTrigger_NotFound(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.get": {Result: []Trigger{}},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	_, err := client.GetTrigger(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetTrigger() expected error for non-existent trigger")
	}
}

func TestClient_UpdateTrigger(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.update": {
			Result: map[string]interface{}{
				"triggerids": []string{"123"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.UpdateTrigger(context.Background(), TriggerUpdateParams{
		TriggerID: "123",
		Status:    TriggerStatusDisabled,
	})
	if err != nil {
		t.Fatalf("UpdateTrigger() error = %v", err)
	}
}

func TestClient_EnableTrigger(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.update": {
			Result: map[string]interface{}{
				"triggerids": []string{"123"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.EnableTrigger(context.Background(), "123")
	if err != nil {
		t.Fatalf("EnableTrigger() error = %v", err)
	}
}

func TestClient_DisableTrigger(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"trigger.update": {
			Result: map[string]interface{}{
				"triggerids": []string{"123"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.DisableTrigger(context.Background(), "123")
	if err != nil {
		t.Fatalf("DisableTrigger() error = %v", err)
	}
}

func TestTrigger_IsEnabled(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"enabled", TriggerStatusEnabled, true},
		{"disabled", TriggerStatusDisabled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := &Trigger{Status: tt.status}
			if got := trigger.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrigger_IsDisabled(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"enabled", TriggerStatusEnabled, false},
		{"disabled", TriggerStatusDisabled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := &Trigger{Status: tt.status}
			if got := trigger.IsDisabled(); got != tt.want {
				t.Errorf("IsDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrigger_IsProblem(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"ok", "0", false},
		{"problem", "1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := &Trigger{Value: tt.value}
			if got := trigger.IsProblem(); got != tt.want {
				t.Errorf("IsProblem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrigger_PriorityInt(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		want     int
	}{
		{"not classified", "0", 0},
		{"information", "1", 1},
		{"warning", "2", 2},
		{"average", "3", 3},
		{"high", "4", 4},
		{"disaster", "5", 5},
		{"invalid", "invalid", 0},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := &Trigger{Priority: tt.priority}
			if got := trigger.PriorityInt(); got != tt.want {
				t.Errorf("PriorityInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
