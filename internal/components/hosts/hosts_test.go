package hosts

import (
	"os"
	"testing"

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

func TestNew(t *testing.T) {
	t.Parallel()

	styles := testStyles()
	m := New(styles)

	if m.styles != styles {
		t.Error("Expected styles to be set")
	}
	if m.Cursor() != 0 {
		t.Error("Expected cursor to start at 0")
	}
	if m.Focused() {
		t.Error("Expected focused to be false initially")
	}
}

func TestSetHosts(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	hosts := []zabbix.Host{
		{HostID: "1", Host: "server1", Name: "Server 1"},
		{HostID: "2", Host: "server2", Name: "Server 2"},
		{HostID: "3", Host: "server3", Name: "Server 3"},
	}

	m.SetHosts(hosts)

	total, filtered := m.ItemCount()
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

	hosts := []zabbix.Host{
		{HostID: "1", Host: "web-server", Name: "Web Server"},
		{HostID: "2", Host: "db-server", Name: "Database Server"},
		{HostID: "3", Host: "cache-server", Name: "Cache Server"},
	}

	m.SetHosts(hosts)

	// Filter by "web"
	m.SetTextFilter("web")

	total, filtered := m.ItemCount()
	if total != 3 {
		t.Errorf("Expected total count 3, got %d", total)
	}
	if filtered != 1 {
		t.Errorf("Expected filtered count 1, got %d", filtered)
	}

	selected := m.Selected()
	if selected == nil || selected.HostID != "1" {
		t.Error("Expected Web Server to be selected")
	}
}

func TestNavigation(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 20)

	hosts := []zabbix.Host{
		{HostID: "1", Host: "server1", Name: "Server 1"},
		{HostID: "2", Host: "server2", Name: "Server 2"},
		{HostID: "3", Host: "server3", Name: "Server 3"},
	}

	m.SetHosts(hosts)

	// Initial selection
	if m.Cursor() != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.Cursor())
	}

	// Move down
	m.MoveDown()
	if m.Cursor() != 1 {
		t.Errorf("Expected cursor at 1 after MoveDown, got %d", m.Cursor())
	}

	// Move down again
	m.MoveDown()
	if m.Cursor() != 2 {
		t.Errorf("Expected cursor at 2 after second MoveDown, got %d", m.Cursor())
	}

	// Move down at end (should stay)
	m.MoveDown()
	if m.Cursor() != 2 {
		t.Errorf("Expected cursor to stay at 2, got %d", m.Cursor())
	}

	// Move up
	m.MoveUp()
	if m.Cursor() != 1 {
		t.Errorf("Expected cursor at 1 after MoveUp, got %d", m.Cursor())
	}

	// Go to top
	m.GoToTop()
	if m.Cursor() != 0 {
		t.Errorf("Expected cursor at 0 after GoToTop, got %d", m.Cursor())
	}

	// Go to bottom
	m.GoToBottom()
	if m.Cursor() != 2 {
		t.Errorf("Expected cursor at 2 after GoToBottom, got %d", m.Cursor())
	}
}

func TestSelected(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	// No hosts - should return nil
	if m.Selected() != nil {
		t.Error("Expected nil when no hosts")
	}

	hosts := []zabbix.Host{
		{HostID: "1", Host: "server1", Name: "Server 1"},
		{HostID: "2", Host: "server2", Name: "Server 2"},
	}

	m.SetHosts(hosts)

	selected := m.Selected()
	if selected == nil {
		t.Fatal("Expected non-nil selected host")
	}
	if selected.HostID != "1" {
		t.Errorf("Expected HostID '1', got '%s'", selected.HostID)
	}

	m.MoveDown()
	selected = m.Selected()
	if selected == nil {
		t.Fatal("Expected non-nil selected host after MoveDown")
	}
	if selected.HostID != "2" {
		t.Errorf("Expected HostID '2', got '%s'", selected.HostID)
	}
}

func TestSelectedIndex(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	// No hosts - should return -1
	if m.SelectedIndex() != -1 {
		t.Error("Expected -1 when no hosts")
	}

	hosts := []zabbix.Host{
		{HostID: "1", Host: "server1", Name: "Server 1"},
		{HostID: "2", Host: "server2", Name: "Server 2"},
	}

	m.SetHosts(hosts)

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

	if m.Focused() {
		t.Error("Expected not focused initially")
	}

	m.SetFocused(true)
	if !m.Focused() {
		t.Error("Expected focused after SetFocused(true)")
	}

	m.SetFocused(false)
	if m.Focused() {
		t.Error("Expected not focused after SetFocused(false)")
	}
}

func TestSetSize(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	m.SetSize(100, 50)

	if m.Width() != 100 {
		t.Errorf("Expected width 100, got %d", m.Width())
	}
	if m.Height() != 50 {
		t.Errorf("Expected height 50, got %d", m.Height())
	}
}

func TestView(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 20)

	hosts := []zabbix.Host{
		{
			HostID:          "1",
			Host:            "server1",
			Name:            "Server 1",
			ActiveAvailable: "1",
			Interfaces: []zabbix.Interface{
				{IP: "192.168.1.1", Main: "1"},
			},
		},
	}

	m.SetHosts(hosts)

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
		t.Error("Expected non-empty view even with no hosts")
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

func TestFilterByIP(t *testing.T) {
	t.Parallel()

	m := New(testStyles())

	hosts := []zabbix.Host{
		{
			HostID: "1",
			Host:   "server1",
			Name:   "Server 1",
			Interfaces: []zabbix.Interface{
				{IP: "192.168.1.10", Main: "1"},
			},
		},
		{
			HostID: "2",
			Host:   "server2",
			Name:   "Server 2",
			Interfaces: []zabbix.Interface{
				{IP: "10.0.0.5", Main: "1"},
			},
		},
	}

	m.SetHosts(hosts)

	// Filter by IP prefix
	m.SetTextFilter("192.168")

	total, filtered := m.ItemCount()
	if total != 2 {
		t.Errorf("Expected total count 2, got %d", total)
	}
	if filtered != 1 {
		t.Errorf("Expected filtered count 1, got %d", filtered)
	}
}

func TestPageNavigation(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 10) // Small height for paging

	// Create many hosts
	hosts := make([]zabbix.Host, 20)
	for i := range hosts {
		hosts[i] = zabbix.Host{
			HostID: string(rune('0' + i)),
			Host:   "server",
			Name:   "Server",
		}
	}

	m.SetHosts(hosts)

	// Page down
	m.PageDown()
	if m.Cursor() == 0 {
		t.Error("PageDown should move cursor")
	}

	// Page up
	m.PageUp()
	// Should move back toward top
	if m.Cursor() >= 10 {
		t.Errorf("PageUp should move cursor back, got %d", m.Cursor())
	}
}

func TestGoToTopBottom(t *testing.T) {
	t.Parallel()

	m := New(testStyles())
	m.SetSize(80, 20)

	hosts := make([]zabbix.Host, 10)
	for i := range hosts {
		hosts[i] = zabbix.Host{
			HostID: string(rune('0' + i)),
			Host:   "server",
			Name:   "Server",
		}
	}

	m.SetHosts(hosts)

	m.GoToBottom()
	if m.Cursor() != 9 {
		t.Errorf("GoToBottom should move to last item (9), got %d", m.Cursor())
	}

	m.GoToTop()
	if m.Cursor() != 0 {
		t.Errorf("GoToTop should move to first item (0), got %d", m.Cursor())
	}
}
