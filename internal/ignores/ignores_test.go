package ignores

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_NewFile(t *testing.T) {
	tmpDir := t.TempDir()

	list, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if list.Len() != 0 {
		t.Errorf("Len() = %d, want 0", list.Len())
	}

	expectedPath := filepath.Join(tmpDir, "ignores.yaml")
	if list.Path() != expectedPath {
		t.Errorf("Path() = %q, want %q", list.Path(), expectedPath)
	}
}

func TestLoad_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ignores.yaml")

	content := `ignores:
  - host_id: "10084"
    host_name: "webserver01"
    trigger_id: "15234"
    trigger_name: "High CPU"
    created: 2024-01-15T10:30:00Z
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	list, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if list.Len() != 1 {
		t.Errorf("Len() = %d, want 1", list.Len())
	}

	rules := list.Rules()
	if rules[0].HostID != "10084" {
		t.Errorf("HostID = %q, want %q", rules[0].HostID, "10084")
	}
	if rules[0].TriggerID != "15234" {
		t.Errorf("TriggerID = %q, want %q", rules[0].TriggerID, "15234")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ignores.yaml")

	if err := os.WriteFile(path, []byte("invalid: [yaml: content"), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := Load(tmpDir)
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()

	list, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	rule := Rule{
		HostID:      "10084",
		HostName:    "webserver01",
		TriggerID:   "15234",
		TriggerName: "High CPU",
		Created:     time.Now(),
	}

	if err = list.Add(rule); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if err = list.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists and contains expected content
	data, err := os.ReadFile(list.Path())
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	content := string(data)
	if !contains(content, "host_id: \"10084\"") {
		t.Error("Saved file should contain host_id")
	}
	if !contains(content, "trigger_id: \"15234\"") {
		t.Error("Saved file should contain trigger_id")
	}
}

func TestAdd(t *testing.T) {
	tmpDir := t.TempDir()
	list, _ := Load(tmpDir)

	rule := Rule{
		HostID:      "10084",
		HostName:    "webserver01",
		TriggerID:   "15234",
		TriggerName: "High CPU",
	}

	// First add should succeed
	if err := list.Add(rule); err != nil {
		t.Errorf("Add() error = %v", err)
	}

	if list.Len() != 1 {
		t.Errorf("Len() = %d, want 1", list.Len())
	}

	// Verify created time was set
	rules := list.Rules()
	if rules[0].Created.IsZero() {
		t.Error("Created time should be set automatically")
	}
}

func TestAdd_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	list, _ := Load(tmpDir)

	rule := Rule{
		HostID:    "10084",
		TriggerID: "15234",
	}

	if err := list.Add(rule); err != nil {
		t.Fatalf("First Add() error = %v", err)
	}

	// Second add with same host+trigger should fail
	err := list.Add(rule)
	if err == nil {
		t.Error("Add() should return error for duplicate")
	}
	if !contains(err.Error(), "already ignored") {
		t.Errorf("Error message should contain 'already ignored', got: %v", err)
	}

	if list.Len() != 1 {
		t.Errorf("Len() = %d, want 1 (duplicate should not be added)", list.Len())
	}
}

func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()
	list, _ := Load(tmpDir)

	// Add some rules
	_ = list.Add(Rule{HostID: "1", TriggerID: "100", HostName: "host1", TriggerName: "trigger1"})
	_ = list.Add(Rule{HostID: "2", TriggerID: "200", HostName: "host2", TriggerName: "trigger2"})
	_ = list.Add(Rule{HostID: "3", TriggerID: "300", HostName: "host3", TriggerName: "trigger3"})

	// Remove middle item (index 2, 1-based)
	removed, ok := list.Remove(2)
	if !ok {
		t.Error("Remove(2) should succeed")
	}
	if removed.HostID != "2" {
		t.Errorf("Removed HostID = %q, want %q", removed.HostID, "2")
	}

	if list.Len() != 2 {
		t.Errorf("Len() = %d, want 2", list.Len())
	}

	// Verify remaining items
	rules := list.Rules()
	if rules[0].HostID != "1" || rules[1].HostID != "3" {
		t.Error("Remaining rules should be host1 and host3")
	}
}

func TestRemove_InvalidIndex(t *testing.T) {
	tmpDir := t.TempDir()
	list, _ := Load(tmpDir)

	_ = list.Add(Rule{HostID: "1", TriggerID: "100"})

	// Invalid indices
	tests := []int{0, -1, 2, 100}
	for _, idx := range tests {
		_, ok := list.Remove(idx)
		if ok {
			t.Errorf("Remove(%d) should fail", idx)
		}
	}
}

func TestIsIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	list, _ := Load(tmpDir)

	_ = list.Add(Rule{HostID: "10084", TriggerID: "15234"})
	_ = list.Add(Rule{HostID: "10085", TriggerID: "15240"})

	tests := []struct {
		hostID    string
		triggerID string
		want      bool
	}{
		{"10084", "15234", true},  // exact match
		{"10085", "15240", true},  // exact match
		{"10084", "15240", false}, // wrong trigger
		{"10085", "15234", false}, // wrong host
		{"99999", "99999", false}, // no match
	}

	for _, tt := range tests {
		got := list.IsIgnored(tt.hostID, tt.triggerID)
		if got != tt.want {
			t.Errorf("IsIgnored(%q, %q) = %v, want %v", tt.hostID, tt.triggerID, got, tt.want)
		}
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	tmpDir := t.TempDir()
	list, _ := Load(tmpDir)

	_ = list.Add(Rule{HostID: "1", TriggerID: "100"})

	rules := list.Rules()
	rules[0].HostID = "modified"

	// Original should be unchanged
	original := list.Rules()
	if original[0].HostID != "1" {
		t.Error("Rules() should return a copy, not the original slice")
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
