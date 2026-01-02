package alerts

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
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

// testProblems returns a slice of test problems.
func testProblems() []zabbix.Problem {
	return []zabbix.Problem{
		{
			EventID:  "1",
			Name:     "CPU usage high",
			Severity: "5",
			Hosts: []zabbix.Host{
				{Name: "server01"},
			},
		},
		{
			EventID:  "2",
			Name:     "Memory warning",
			Severity: "3",
			Hosts: []zabbix.Host{
				{Name: "server02"},
			},
		},
		{
			EventID:  "3",
			Name:     "Disk space low",
			Severity: "2",
			Hosts: []zabbix.Host{
				{Name: "server03"},
			},
		},
		{
			EventID:      "4",
			Name:         "Network latency",
			Severity:     "4",
			Acknowledged: "1",
			Hosts: []zabbix.Host{
				{Name: "server04"},
			},
		},
		{
			EventID:  "5",
			Name:     "Service unavailable",
			Severity: "5",
			Hosts: []zabbix.Host{
				{Name: "webserver01"},
			},
		},
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	styles := testStyles()
	m := New(styles)

	if m.styles != styles {
		t.Error("New() should set styles")
	}
	if m.cursor != 0 {
		t.Errorf("New().cursor = %d, want 0", m.cursor)
	}
	if m.focused {
		t.Error("New().focused should be false")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("SetSize() width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("SetSize() height = %d, want 50", m.height)
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	m.SetFocused(true)
	if !m.focused {
		t.Error("SetFocused(true) should set focused to true")
	}

	m.SetFocused(false)
	if m.focused {
		t.Error("SetFocused(false) should set focused to false")
	}
}

func TestModel_SetProblems(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	problems := testProblems()

	m.SetProblems(problems)

	total, filtered := m.Count()
	if total != 5 {
		t.Errorf("SetProblems() total = %d, want 5", total)
	}
	if filtered != 5 {
		t.Errorf("SetProblems() filtered = %d, want 5", filtered)
	}
}

func TestModel_SetMinSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		minSeverity  int
		wantFiltered int
	}{
		{"no filter", 0, 5},
		{"severity 2+", 2, 5},
		{"severity 3+", 3, 4},
		{"severity 4+", 4, 3},
		{"severity 5+", 5, 2},
		{"severity 6+ (none)", 6, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := New(testStyles())
			m.SetProblems(testProblems())
			m.SetMinSeverity(tt.minSeverity)

			_, filtered := m.Count()
			if filtered != tt.wantFiltered {
				t.Errorf("SetMinSeverity(%d) filtered = %d, want %d", tt.minSeverity, filtered, tt.wantFiltered)
			}
		})
	}
}

func TestModel_SetTextFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		filter       string
		wantFiltered int
	}{
		{"no filter", "", 5},
		{"filter by problem name", "cpu", 1},
		{"filter by host name", "server01", 2},  // matches "server01" in host and potentially problem name
		{"filter by partial name", "server", 5}, // all hosts contain "server"
		{"case insensitive", "CPU", 1},
		{"no match", "nonexistent", 0},
		{"webserver", "webserver", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := New(testStyles())
			m.SetProblems(testProblems())
			m.SetTextFilter(tt.filter)

			_, filtered := m.Count()
			if filtered != tt.wantFiltered {
				t.Errorf("SetTextFilter(%q) filtered = %d, want %d", tt.filter, filtered, tt.wantFiltered)
			}
		})
	}
}

func TestModel_Selected(t *testing.T) {
	t.Parallel()

	t.Run("no problems", func(t *testing.T) {
		t.Parallel()

		m := New(testStyles())
		if m.Selected() != nil {
			t.Error("Selected() should return nil when no problems")
		}
	})

	t.Run("with problems", func(t *testing.T) {
		t.Parallel()

		m := New(testStyles())
		m.SetProblems(testProblems())

		selected := m.Selected()
		if selected == nil {
			t.Fatal("Selected() should not return nil")
		}
		if selected.EventID != "1" {
			t.Errorf("Selected().EventID = %q, want %q", selected.EventID, "1")
		}
	})

	t.Run("after navigation", func(t *testing.T) {
		t.Parallel()

		m := New(testStyles())
		m.SetProblems(testProblems())
		m.SetFocused(true)
		m.SetSize(100, 50)

		m.MoveDown()
		selected := m.Selected()
		if selected == nil {
			t.Fatal("Selected() should not return nil")
		}
		if selected.EventID != "2" {
			t.Errorf("After MoveDown, Selected().EventID = %q, want %q", selected.EventID, "2")
		}
	})
}

func TestModel_SelectedIndex(t *testing.T) {
	t.Parallel()

	t.Run("no selection", func(t *testing.T) {
		t.Parallel()

		m := New(testStyles())
		if m.SelectedIndex() != -1 {
			t.Error("SelectedIndex() should return -1 when no problems")
		}
	})

	t.Run("with selection", func(t *testing.T) {
		t.Parallel()

		m := New(testStyles())
		m.SetProblems(testProblems())

		if m.SelectedIndex() != 0 {
			t.Errorf("SelectedIndex() = %d, want 0", m.SelectedIndex())
		}

		m.MoveDown()
		if m.SelectedIndex() != 1 {
			t.Errorf("After MoveDown, SelectedIndex() = %d, want 1", m.SelectedIndex())
		}
	})
}

func TestModel_MoveUp(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 50)

	// Move down first, then up
	m.MoveDown()
	m.MoveDown()
	if m.cursor != 2 {
		t.Errorf("cursor = %d, want 2", m.cursor)
	}

	m.MoveUp()
	if m.cursor != 1 {
		t.Errorf("After MoveUp, cursor = %d, want 1", m.cursor)
	}

	// MoveUp at top should stay at 0
	m.MoveUp()
	m.MoveUp() // Try to go negative
	if m.cursor != 0 {
		t.Errorf("MoveUp at top, cursor = %d, want 0", m.cursor)
	}
}

func TestModel_MoveDown(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 50)

	m.MoveDown()
	if m.cursor != 1 {
		t.Errorf("MoveDown cursor = %d, want 1", m.cursor)
	}

	// Move to end
	m.MoveDown()
	m.MoveDown()
	m.MoveDown()
	if m.cursor != 4 {
		t.Errorf("cursor at end = %d, want 4", m.cursor)
	}

	// MoveDown at end should stay at 4
	m.MoveDown()
	if m.cursor != 4 {
		t.Errorf("MoveDown at end, cursor = %d, want 4", m.cursor)
	}
}

func TestModel_PageUp(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 10) // height of 10 means visibleRows = 8

	// Move to end first
	m.GoToBottom()
	if m.cursor != 4 {
		t.Errorf("cursor at bottom = %d, want 4", m.cursor)
	}

	// PageUp should move by visibleRows
	m.PageUp()
	// With 5 items and pageSize of 8, PageUp from 4 should go to 0
	if m.cursor < 0 {
		t.Errorf("PageUp cursor = %d, should not be negative", m.cursor)
	}
}

func TestModel_PageDown(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 10)

	// PageDown from start
	m.PageDown()
	// With 5 items and pageSize of 8, PageDown from 0 would try to go to 8
	// but should be clamped to 4 (last item)
	if m.cursor != 4 {
		t.Errorf("PageDown cursor = %d, want 4", m.cursor)
	}
}

func TestModel_GoToTop(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 50)

	// Move down first
	m.MoveDown()
	m.MoveDown()

	m.GoToTop()
	if m.cursor != 0 {
		t.Errorf("GoToTop cursor = %d, want 0", m.cursor)
	}
	if m.offset != 0 {
		t.Errorf("GoToTop offset = %d, want 0", m.offset)
	}
}

func TestModel_GoToBottom(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 50)

	m.GoToBottom()
	if m.cursor != 4 {
		t.Errorf("GoToBottom cursor = %d, want 4", m.cursor)
	}
}

func TestModel_GoToBottom_Empty(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(100, 50)

	m.GoToBottom()
	if m.cursor != 0 {
		t.Errorf("GoToBottom on empty list cursor = %d, want 0", m.cursor)
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	cmd := m.Init()

	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_Update_NotFocused(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetFocused(false)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, cmd := m.Update(msg)

	if newModel.cursor != 0 {
		t.Error("Update should not change cursor when not focused")
	}
	if cmd != nil {
		t.Error("Update should return nil cmd when not focused")
	}
}

func TestModel_Update_KeyBindings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		keys       []tea.KeyMsg
		wantCursor int
	}{
		{
			name:       "down arrow",
			keys:       []tea.KeyMsg{{Type: tea.KeyDown}},
			wantCursor: 1,
		},
		{
			name:       "j key",
			keys:       []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune{'j'}}},
			wantCursor: 1,
		},
		{
			name:       "up arrow after down",
			keys:       []tea.KeyMsg{{Type: tea.KeyDown}, {Type: tea.KeyUp}},
			wantCursor: 0,
		},
		{
			name:       "k key after down",
			keys:       []tea.KeyMsg{{Type: tea.KeyDown}, {Type: tea.KeyRunes, Runes: []rune{'k'}}},
			wantCursor: 0,
		},
		{
			name:       "home key",
			keys:       []tea.KeyMsg{{Type: tea.KeyDown}, {Type: tea.KeyDown}, {Type: tea.KeyHome}},
			wantCursor: 0,
		},
		{
			name:       "g key (home)",
			keys:       []tea.KeyMsg{{Type: tea.KeyDown}, {Type: tea.KeyDown}, {Type: tea.KeyRunes, Runes: []rune{'g'}}},
			wantCursor: 0,
		},
		{
			name:       "end key",
			keys:       []tea.KeyMsg{{Type: tea.KeyEnd}},
			wantCursor: 4,
		},
		{
			name:       "G key (end)",
			keys:       []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune{'G'}}},
			wantCursor: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := New(testStyles())
			m.SetProblems(testProblems())
			m.SetFocused(true)
			m.SetSize(100, 50)

			var newModel Model
			for _, key := range tt.keys {
				newModel, _ = m.Update(key)
				m = newModel
			}

			if newModel.cursor != tt.wantCursor {
				t.Errorf("After %s, cursor = %d, want %d", tt.name, newModel.cursor, tt.wantCursor)
			}
		})
	}
}

func TestModel_View_ZeroSize(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(5, 3) // Too small

	view := m.View()
	if view != "" {
		t.Errorf("View() with zero size should return empty string, got %q", view)
	}
}

func TestModel_View_Basic(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(80, 20)
	m.SetFocused(true)

	view := m.View()

	// Should contain header with count
	if !strings.Contains(view, "ALERTS") {
		t.Error("View should contain ALERTS header")
	}
	if !strings.Contains(view, "(5)") {
		t.Error("View should contain problem count")
	}
}

func TestModel_View_FilteredCount(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetMinSeverity(5)
	m.SetSize(80, 20)
	m.SetFocused(true)

	view := m.View()

	// Should show filtered/total format
	if !strings.Contains(view, "(2/5)") {
		t.Error("View should show filtered/total when filters active")
	}
}

func TestModel_View_FocusedVsBlurred(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(80, 20)

	// Get focused view
	m.SetFocused(true)
	focusedView := m.View()

	// Get blurred view
	m.SetFocused(false)
	blurredView := m.View()

	// Views should both render (we can't easily compare lipgloss styles)
	// Just verify both render without error
	if focusedView == "" {
		t.Error("Focused view should not be empty")
	}
	if blurredView == "" {
		t.Error("Blurred view should not be empty")
	}
}

func TestModel_CombinedFilters(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())

	// Apply both severity and text filter
	m.SetMinSeverity(4)
	m.SetTextFilter("server")

	_, filtered := m.Count()
	// Severity 4+ AND contains "server" in host or problem name:
	// - server01 (sev 5) - host contains "server"
	// - server04 (sev 4) - host contains "server"
	// - webserver01 (sev 5) - host contains "server"
	// = 3 items
	if filtered != 3 {
		t.Errorf("Combined filters filtered = %d, want 3", filtered)
	}
}

func TestModel_FilterResetsCursor(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 50)

	// Move cursor down
	m.MoveDown()
	m.MoveDown()
	if m.cursor != 2 {
		t.Errorf("cursor = %d, want 2", m.cursor)
	}

	// Apply filter that reduces list
	m.SetMinSeverity(5) // Only 2 items (index 0, 1)

	// Cursor should be adjusted to be within bounds
	if m.cursor >= 2 {
		t.Errorf("After filter, cursor = %d, should be < 2", m.cursor)
	}
}

func TestModel_Count(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	// Empty
	total, filtered := m.Count()
	if total != 0 || filtered != 0 {
		t.Errorf("Empty Count() = (%d, %d), want (0, 0)", total, filtered)
	}

	// With problems
	m.SetProblems(testProblems())
	total, filtered = m.Count()
	if total != 5 || filtered != 5 {
		t.Errorf("Count() = (%d, %d), want (5, 5)", total, filtered)
	}

	// With filter
	m.SetMinSeverity(5)
	total, filtered = m.Count()
	if total != 5 || filtered != 2 {
		t.Errorf("Filtered Count() = (%d, %d), want (5, 2)", total, filtered)
	}
}

func TestModel_visibleRows(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(100, 20)

	// visibleRows = height - 2 (header and border)
	visible := m.visibleRows()
	if visible != 18 {
		t.Errorf("visibleRows() = %d, want 18", visible)
	}
}

func TestModel_ensureVisible(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 5) // Small height for testing scrolling

	// visibleRows = 5 - 2 = 3

	// Move to bottom
	m.GoToBottom() // cursor = 4

	// offset should be adjusted so cursor is visible
	// With 3 visible rows and cursor at 4, offset should be 4 - 3 + 1 = 2
	if m.offset < 2 {
		t.Errorf("After GoToBottom, offset = %d, should be >= 2", m.offset)
	}
}

func TestModel_AcknowledgedDisplay(t *testing.T) {
	t.Parallel()

	// Problem at index 3 is acknowledged
	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(80, 20)

	// Navigate to acknowledged problem
	m.MoveDown()
	m.MoveDown()
	m.MoveDown() // Now at index 3

	selected := m.Selected()
	if selected == nil {
		t.Fatal("Selected should not be nil")
	}
	if !selected.IsAcknowledged() {
		t.Error("Problem at index 3 should be acknowledged")
	}
}

func TestModel_EmptyProblems(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems([]zabbix.Problem{})
	m.SetSize(80, 20)

	// Should handle empty list gracefully
	total, filtered := m.Count()
	if total != 0 || filtered != 0 {
		t.Errorf("Empty problems Count() = (%d, %d), want (0, 0)", total, filtered)
	}

	if m.Selected() != nil {
		t.Error("Selected() on empty list should return nil")
	}

	// Navigation should not panic
	m.MoveUp()
	m.MoveDown()
	m.PageUp()
	m.PageDown()
	m.GoToTop()
	m.GoToBottom()

	// View should not panic
	_ = m.View()
}

func TestModel_PageUpPageDown_EdgeCases(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetProblems(testProblems())
	m.SetSize(100, 100) // Large size, all items visible

	// PageDown when all items visible should go to last
	m.PageDown()
	if m.cursor != 4 {
		t.Errorf("PageDown cursor = %d, want 4", m.cursor)
	}

	// PageUp from last should go to first when all visible
	m.PageUp()
	if m.cursor < 0 {
		t.Errorf("PageUp cursor = %d, should not be negative", m.cursor)
	}
}

func TestModel_LongHostName(t *testing.T) {
	t.Parallel()

	problems := []zabbix.Problem{
		{
			EventID:  "1",
			Name:     "Test problem",
			Severity: "5",
			Hosts: []zabbix.Host{
				{Name: "very-long-hostname-that-exceeds-limit"},
			},
		},
	}

	m := New(testStyles())
	m.SetProblems(problems)
	m.SetSize(80, 20)

	// Should not panic and should truncate
	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestModel_LongProblemName(t *testing.T) {
	t.Parallel()

	problems := []zabbix.Problem{
		{
			EventID:  "1",
			Name:     "This is a very long problem name that will definitely need to be truncated to fit in the display area",
			Severity: "5",
			Hosts: []zabbix.Host{
				{Name: "server01"},
			},
		},
	}

	m := New(testStyles())
	m.SetProblems(problems)
	m.SetSize(60, 20) // Narrow width

	// Should not panic and should truncate
	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}
