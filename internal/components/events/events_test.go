package events

import (
	"testing"

	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// testStyles returns a theme.Styles instance for testing.
func testStyles() *theme.Styles {
	return theme.NewStyles(theme.DefaultTheme())
}

func TestNew(t *testing.T) {
	t.Parallel()

	styles := testStyles()
	m := New(styles)

	if m.styles != styles {
		t.Error("Expected styles to be set")
	}
	if m.cursor != 0 {
		t.Error("Expected cursor to start at 0")
	}
	if m.focused {
		t.Error("Expected focused to be false initially")
	}
}

func TestSetEvents(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	events := []zabbix.Event{
		{EventID: "1", Name: "Disk space low", Severity: "3"},
		{EventID: "2", Name: "CPU high", Severity: "4"},
		{EventID: "3", Name: "Memory warning", Severity: "2"},
	}

	m.SetEvents(events)

	total, filtered := m.Count()
	if total != 3 {
		t.Errorf("Expected total count 3, got %d", total)
	}
	if filtered != 3 {
		t.Errorf("Expected filtered count 3, got %d", filtered)
	}
}

func TestSetTextFilter(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	events := []zabbix.Event{
		{
			EventID:  "1",
			Name:     "Disk space low",
			Severity: "3",
			Hosts:    []zabbix.Host{{Host: "web-server", Name: "Web Server"}},
		},
		{
			EventID:  "2",
			Name:     "CPU high",
			Severity: "4",
			Hosts:    []zabbix.Host{{Host: "db-server", Name: "Database Server"}},
		},
		{
			EventID:  "3",
			Name:     "Memory warning",
			Severity: "2",
			Hosts:    []zabbix.Host{{Host: "cache-server", Name: "Cache Server"}},
		},
	}

	m.SetEvents(events)

	// Filter by event name
	m.SetTextFilter("disk")

	total, filtered := m.Count()
	if total != 3 {
		t.Errorf("Expected total count 3, got %d", total)
	}
	if filtered != 1 {
		t.Errorf("Expected filtered count 1, got %d", filtered)
	}

	selected := m.Selected()
	if selected == nil || selected.EventID != "1" {
		t.Error("Expected Disk space event to be selected")
	}

	// Filter by host
	m.SetTextFilter("database")

	_, filtered = m.Count()
	if filtered != 1 {
		t.Errorf("Expected filtered count 1 for database filter, got %d", filtered)
	}
}

func TestNavigation(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 20)

	events := []zabbix.Event{
		{EventID: "1", Name: "Event 1"},
		{EventID: "2", Name: "Event 2"},
		{EventID: "3", Name: "Event 3"},
	}

	m.SetEvents(events)

	// Initial selection
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	// Move down
	m.MoveDown()
	if m.cursor != 1 {
		t.Errorf("Expected cursor at 1 after MoveDown, got %d", m.cursor)
	}

	// Move down again
	m.MoveDown()
	if m.cursor != 2 {
		t.Errorf("Expected cursor at 2 after second MoveDown, got %d", m.cursor)
	}

	// Move down at end (should stay)
	m.MoveDown()
	if m.cursor != 2 {
		t.Errorf("Expected cursor to stay at 2, got %d", m.cursor)
	}

	// Move up
	m.MoveUp()
	if m.cursor != 1 {
		t.Errorf("Expected cursor at 1 after MoveUp, got %d", m.cursor)
	}

	// Go to top
	m.GoToTop()
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0 after GoToTop, got %d", m.cursor)
	}

	// Go to bottom
	m.GoToBottom()
	if m.cursor != 2 {
		t.Errorf("Expected cursor at 2 after GoToBottom, got %d", m.cursor)
	}
}

func TestSelected(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	// No events - should return nil
	if m.Selected() != nil {
		t.Error("Expected nil when no events")
	}

	events := []zabbix.Event{
		{EventID: "1", Name: "Event 1"},
		{EventID: "2", Name: "Event 2"},
	}

	m.SetEvents(events)

	selected := m.Selected()
	if selected == nil {
		t.Fatal("Expected non-nil selected event")
	}
	if selected.EventID != "1" {
		t.Errorf("Expected EventID '1', got '%s'", selected.EventID)
	}

	m.MoveDown()
	selected = m.Selected()
	if selected == nil {
		t.Fatal("Expected non-nil selected event after MoveDown")
	}
	if selected.EventID != "2" {
		t.Errorf("Expected EventID '2', got '%s'", selected.EventID)
	}
}

func TestSelectedIndex(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	// No events - should return -1
	if m.SelectedIndex() != -1 {
		t.Error("Expected -1 when no events")
	}

	events := []zabbix.Event{
		{EventID: "1", Name: "Event 1"},
		{EventID: "2", Name: "Event 2"},
	}

	m.SetEvents(events)

	if m.SelectedIndex() != 0 {
		t.Errorf("Expected index 0, got %d", m.SelectedIndex())
	}

	m.MoveDown()
	if m.SelectedIndex() != 1 {
		t.Errorf("Expected index 1, got %d", m.SelectedIndex())
	}
}

func TestSetFocused(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	if m.focused {
		t.Error("Expected not focused initially")
	}

	m.SetFocused(true)
	if !m.focused {
		t.Error("Expected focused after SetFocused(true)")
	}

	m.SetFocused(false)
	if m.focused {
		t.Error("Expected not focused after SetFocused(false)")
	}
}

func TestSetSize(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("Expected height 50, got %d", m.height)
	}
}

func TestView(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 20)

	events := []zabbix.Event{
		{
			EventID:  "1",
			Name:     "Disk space low",
			Severity: "3",
			Clock:    "1700000000",
			Hosts:    []zabbix.Host{{Host: "server1", Name: "Server 1"}},
		},
	}

	m.SetEvents(events)

	view := m.View()

	// Should contain header
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestViewEmpty(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 20)

	view := m.View()

	// Should still render (empty list)
	if view == "" {
		t.Error("Expected non-empty view even with no events")
	}
}

func TestViewZeroSize(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(5, 3) // Too small

	view := m.View()
	if view != "" {
		t.Errorf("View with zero size should return empty string, got %q", view)
	}
}

func TestPageNavigation(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 10) // Small height for paging

	// Create many events
	events := make([]zabbix.Event, 20)
	for i := range events {
		events[i] = zabbix.Event{
			EventID: string(rune('0' + i)),
			Name:    "Event",
		}
	}

	m.SetEvents(events)

	// Page down
	m.PageDown()
	if m.cursor == 0 {
		t.Error("PageDown should move cursor")
	}

	// Page up
	m.PageUp()
	// Should move back toward top
	if m.cursor >= 10 {
		t.Errorf("PageUp should move cursor back, got %d", m.cursor)
	}
}

func TestGoToTopBottom(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 20)

	events := make([]zabbix.Event, 10)
	for i := range events {
		events[i] = zabbix.Event{
			EventID: string(rune('0' + i)),
			Name:    "Event",
		}
	}

	m.SetEvents(events)

	m.GoToBottom()
	if m.cursor != 9 {
		t.Errorf("GoToBottom should move to last item (9), got %d", m.cursor)
	}

	m.GoToTop()
	if m.cursor != 0 {
		t.Errorf("GoToTop should move to first item (0), got %d", m.cursor)
	}
}

func TestRecoveryEvent(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 20)

	// Create a recovery event (has REventID set)
	events := []zabbix.Event{
		{
			EventID:  "1",
			Name:     "Disk space recovered",
			Severity: "3",
			Clock:    "1700000000",
			RClock:   "1700003600", // 1 hour later
			REventID: "100",        // Has recovery event ID = this is a resolved problem
			Hosts:    []zabbix.Host{{Host: "server1", Name: "Server 1"}},
		},
	}

	m.SetEvents(events)

	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view for recovery event")
	}

	selected := m.Selected()
	if selected == nil {
		t.Fatal("Expected non-nil selected event")
	}
	if !selected.IsRecovery() {
		t.Error("Expected event to be identified as recovery")
	}
}
