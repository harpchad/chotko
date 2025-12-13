package tabs

import (
	"os"
	"testing"

	zone "github.com/lrstanley/bubblezone"

	"github.com/harpchad/chotko/internal/theme"
)

// TestMain initializes the zone manager for tests that call View().
func TestMain(m *testing.M) {
	zone.NewGlobal()
	os.Exit(m.Run())
}

// testStyles returns a theme.Styles instance for testing.
func testStyles() *theme.Styles {
	return theme.NewStyles(theme.DefaultTheme())
}

func TestNew(t *testing.T) {
	styles := testStyles()
	tabs := []string{"Alerts", "Hosts", "Events", "Graphs"}
	m := New(styles, tabs, 0)

	if m.styles != styles {
		t.Error("Expected styles to be set")
	}
	if len(m.tabs) != 4 {
		t.Errorf("Expected 4 tabs, got %d", len(m.tabs))
	}
	if m.active != 0 {
		t.Errorf("Expected active tab 0, got %d", m.active)
	}
}

func TestSetWidth(t *testing.T) {
	m := New(testStyles(), []string{"Tab1", "Tab2"}, 0)
	m.SetWidth(100)

	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
}

func TestSetActive(t *testing.T) {
	m := New(testStyles(), []string{"Tab1", "Tab2", "Tab3"}, 0)

	m.SetActive(1)
	if m.active != 1 {
		t.Errorf("Expected active 1, got %d", m.active)
	}

	m.SetActive(2)
	if m.active != 2 {
		t.Errorf("Expected active 2, got %d", m.active)
	}

	// Out of bounds should not change
	m.SetActive(10)
	if m.active != 2 {
		t.Errorf("Expected active to remain 2, got %d", m.active)
	}

	m.SetActive(-1)
	if m.active != 2 {
		t.Errorf("Expected active to remain 2, got %d", m.active)
	}
}

func TestActive(t *testing.T) {
	m := New(testStyles(), []string{"Tab1", "Tab2"}, 1)

	if m.Active() != 1 {
		t.Errorf("Expected Active() to return 1, got %d", m.Active())
	}
}

func TestActiveTab(t *testing.T) {
	m := New(testStyles(), []string{"Alerts", "Hosts", "Events"}, 0)

	if m.ActiveTab() != "Alerts" {
		t.Errorf("Expected ActiveTab() to return 'Alerts', got %s", m.ActiveTab())
	}

	m.SetActive(1)
	if m.ActiveTab() != "Hosts" {
		t.Errorf("Expected ActiveTab() to return 'Hosts', got %s", m.ActiveTab())
	}

	m.SetActive(2)
	if m.ActiveTab() != "Events" {
		t.Errorf("Expected ActiveTab() to return 'Events', got %s", m.ActiveTab())
	}
}

func TestNext(t *testing.T) {
	m := New(testStyles(), []string{"Tab1", "Tab2", "Tab3"}, 0)

	m.Next()
	if m.active != 1 {
		t.Errorf("Expected active 1 after Next, got %d", m.active)
	}

	m.Next()
	if m.active != 2 {
		t.Errorf("Expected active 2 after Next, got %d", m.active)
	}

	// Should wrap around
	m.Next()
	if m.active != 0 {
		t.Errorf("Expected active 0 after wrap, got %d", m.active)
	}
}

func TestPrev(t *testing.T) {
	m := New(testStyles(), []string{"Tab1", "Tab2", "Tab3"}, 0)

	// Should wrap around to end
	m.Prev()
	if m.active != 2 {
		t.Errorf("Expected active 2 after Prev from 0, got %d", m.active)
	}

	m.Prev()
	if m.active != 1 {
		t.Errorf("Expected active 1 after Prev, got %d", m.active)
	}

	m.Prev()
	if m.active != 0 {
		t.Errorf("Expected active 0 after Prev, got %d", m.active)
	}
}

func TestInit(t *testing.T) {
	m := New(testStyles(), []string{"Tab1"}, 0)
	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil")
	}
}

func TestUpdate(t *testing.T) {
	m := New(testStyles(), []string{"Tab1"}, 0)
	newModel, cmd := m.Update(nil)

	if cmd != nil {
		t.Error("Expected Update to return nil cmd")
	}
	if newModel.active != m.active {
		t.Error("Expected model to be unchanged")
	}
}

func TestView(t *testing.T) {
	m := New(testStyles(), []string{"Alerts", "Hosts", "Events", "Graphs"}, 0)
	m.SetWidth(80)

	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain all tab names
	if !containsString(view, "Alerts") {
		t.Error("Expected view to contain 'Alerts'")
	}
	if !containsString(view, "Hosts") {
		t.Error("Expected view to contain 'Hosts'")
	}
	if !containsString(view, "Events") {
		t.Error("Expected view to contain 'Events'")
	}
	if !containsString(view, "Graphs") {
		t.Error("Expected view to contain 'Graphs'")
	}
}

func TestViewActiveTabHighlighted(t *testing.T) {
	m := New(testStyles(), []string{"Tab1", "Tab2"}, 0)
	m.SetWidth(80)

	// Active tab should have brackets
	view := m.View()
	if !containsString(view, "[Tab1]") {
		t.Error("Expected active tab to have brackets")
	}
}

func TestActiveTabOutOfBounds(t *testing.T) {
	m := New(testStyles(), []string{"Tab1"}, 0)
	m.active = 10 // Force out of bounds

	result := m.ActiveTab()
	if result != "" {
		t.Errorf("Expected empty string for out of bounds, got %s", result)
	}
}

// Helper to check if string contains substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
