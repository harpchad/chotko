package zabbix

import (
	"context"
	"testing"
)

func TestClient_GetProblems(t *testing.T) {
	// GetProblems does a two-step fetch:
	// 1. problem.get to get active problem eventids
	// 2. event.get with those eventids to get full details
	server := newMockServer(t, map[string]mockResponse{
		"problem.get": {
			Result: []struct {
				EventID string `json:"eventid"`
			}{
				{EventID: "123"},
				{EventID: "456"},
			},
		},
		"event.get": {
			Result: []Problem{
				{
					EventID:  "123",
					Name:     "Problem 1",
					Severity: "4",
					Hosts:    []Host{{HostID: "1", Name: "Host 1"}},
					RelatedObject: RelatedObject{
						TriggerID: "100",
						Status:    "0", // Enabled trigger
					},
				},
				{
					EventID:  "456",
					Name:     "Problem 2",
					Severity: "3",
					Hosts:    []Host{{HostID: "2", Name: "Host 2"}},
					RelatedObject: RelatedObject{
						TriggerID: "101",
						Status:    "0", // Enabled trigger
					},
				},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	problems, err := client.GetProblems(context.Background(), DefaultProblemGetParams())
	if err != nil {
		t.Fatalf("GetProblems() error = %v", err)
	}

	if len(problems) != 2 {
		t.Errorf("len(problems) = %d, want 2", len(problems))
	}
}

func TestClient_GetProblems_FiltersDisabledTriggers(t *testing.T) {
	// This tests the critical bug fix: problems from disabled triggers should be filtered out
	server := newMockServer(t, map[string]mockResponse{
		"problem.get": {
			Result: []struct {
				EventID string `json:"eventid"`
			}{
				{EventID: "123"},
				{EventID: "456"},
				{EventID: "789"},
			},
		},
		"event.get": {
			Result: []Problem{
				{
					EventID:  "123",
					Name:     "Active Problem",
					Severity: "4",
					RelatedObject: RelatedObject{
						Status: "0", // Enabled trigger
					},
				},
				{
					EventID:  "456",
					Name:     "Disabled Trigger Problem",
					Severity: "3",
					RelatedObject: RelatedObject{
						Status: "1", // Disabled trigger - should be filtered
					},
				},
				{
					EventID:  "789",
					Name:     "Another Active Problem",
					Severity: "2",
					RelatedObject: RelatedObject{
						Status: "0", // Enabled trigger
					},
				},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	problems, err := client.GetProblems(context.Background(), DefaultProblemGetParams())
	if err != nil {
		t.Fatalf("GetProblems() error = %v", err)
	}

	// Should only have 2 problems (the one with disabled trigger filtered out)
	if len(problems) != 2 {
		t.Errorf("len(problems) = %d, want 2 (disabled trigger problem should be filtered)", len(problems))
	}

	// Verify the disabled trigger problem is not in the result
	for _, p := range problems {
		if p.Name == "Disabled Trigger Problem" {
			t.Error("Problem from disabled trigger should have been filtered out")
		}
	}
}

func TestClient_GetProblems_Empty(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"problem.get": {Result: []struct {
			EventID string `json:"eventid"`
		}{}},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	problems, err := client.GetProblems(context.Background(), DefaultProblemGetParams())
	if err != nil {
		t.Fatalf("GetProblems() error = %v", err)
	}

	if len(problems) != 0 {
		t.Errorf("len(problems) = %d, want 0", len(problems))
	}
}

func TestClient_GetActiveProblems(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"problem.get": {
			Result: []struct {
				EventID string `json:"eventid"`
			}{
				{EventID: "123"},
			},
		},
		"event.get": {
			Result: []Problem{
				{
					EventID:  "123",
					Name:     "Active Problem",
					Severity: "4",
					RelatedObject: RelatedObject{
						Status: "0",
					},
				},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	problems, err := client.GetActiveProblems(context.Background())
	if err != nil {
		t.Fatalf("GetActiveProblems() error = %v", err)
	}

	if len(problems) != 1 {
		t.Errorf("len(problems) = %d, want 1", len(problems))
	}
}

func TestClient_GetProblemsWithMinSeverity(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"problem.get": {
			Result: []struct {
				EventID string `json:"eventid"`
			}{
				{EventID: "123"},
			},
		},
		"event.get": {
			Result: []Problem{
				{
					EventID:  "123",
					Name:     "High Severity Problem",
					Severity: "4",
					RelatedObject: RelatedObject{
						Status: "0",
					},
				},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	problems, err := client.GetProblemsWithMinSeverity(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetProblemsWithMinSeverity() error = %v", err)
	}

	if len(problems) != 1 {
		t.Errorf("len(problems) = %d, want 1", len(problems))
	}
}

func TestClient_AcknowledgeProblem(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"event.acknowledge": {
			Result: map[string]any{"eventids": []string{"123"}},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.AcknowledgeProblem(context.Background(), "123", "Acknowledged via API")
	if err != nil {
		t.Fatalf("AcknowledgeProblem() error = %v", err)
	}
}

func TestClient_AcknowledgeProblem_NoMessage(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"event.acknowledge": {
			Result: map[string]any{"eventids": []string{"123"}},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	// With empty message, should use different action bitmask
	err := client.AcknowledgeProblem(context.Background(), "123", "")
	if err != nil {
		t.Fatalf("AcknowledgeProblem() with no message error = %v", err)
	}
}

func TestClient_AcknowledgeProblems(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"event.acknowledge": {
			Result: map[string]any{"eventids": []string{"123", "456"}},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.AcknowledgeProblems(context.Background(), []string{"123", "456"}, "Bulk ack")
	if err != nil {
		t.Fatalf("AcknowledgeProblems() error = %v", err)
	}
}

func TestClient_CloseProblem(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"event.acknowledge": {
			Result: map[string]any{"eventids": []string{"123"}},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.CloseProblem(context.Background(), "123", "Closing problem")
	if err != nil {
		t.Fatalf("CloseProblem() error = %v", err)
	}
}

func TestClient_SuppressProblem(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"event.acknowledge": {
			Result: map[string]any{"eventids": []string{"123"}},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.SuppressProblem(context.Background(), "123")
	if err != nil {
		t.Fatalf("SuppressProblem() error = %v", err)
	}
}

func TestClient_UnsuppressProblem(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"event.acknowledge": {
			Result: map[string]any{"eventids": []string{"123"}},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.UnsuppressProblem(context.Background(), "123")
	if err != nil {
		t.Fatalf("UnsuppressProblem() error = %v", err)
	}
}

func TestClient_AcknowledgeProblem_NumericEventIDs(t *testing.T) {
	// Some Zabbix versions return numeric event IDs instead of strings
	server := newMockServer(t, map[string]mockResponse{
		"event.acknowledge": {
			Result: map[string]any{"eventids": []int{123}},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	err := client.AcknowledgeProblem(context.Background(), "123", "Acknowledged via API")
	if err != nil {
		t.Fatalf("AcknowledgeProblem() with numeric eventids error = %v", err)
	}
}

func TestAcknowledgeActionConstants(t *testing.T) {
	// Verify the action bitmask constants are correct
	// These are documented in Zabbix API documentation
	if ActionClose != 1 {
		t.Errorf("ActionClose = %d, want 1", ActionClose)
	}
	if ActionAcknowledge != 2 {
		t.Errorf("ActionAcknowledge = %d, want 2", ActionAcknowledge)
	}
	if ActionAddMessage != 4 {
		t.Errorf("ActionAddMessage = %d, want 4", ActionAddMessage)
	}
	if ActionChangeSeverity != 8 {
		t.Errorf("ActionChangeSeverity = %d, want 8", ActionChangeSeverity)
	}
	if ActionUnacknowledge != 16 {
		t.Errorf("ActionUnacknowledge = %d, want 16", ActionUnacknowledge)
	}
	if ActionSuppress != 32 {
		t.Errorf("ActionSuppress = %d, want 32", ActionSuppress)
	}
	if ActionUnsuppress != 64 {
		t.Errorf("ActionUnsuppress = %d, want 64", ActionUnsuppress)
	}
}
