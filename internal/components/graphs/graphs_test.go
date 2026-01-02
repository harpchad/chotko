package graphs

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

func TestBuildTree(t *testing.T) {
	items := []zabbix.Item{
		{
			ItemID:    "1",
			HostID:    "100",
			Name:      "CPU utilization",
			Key:       "system.cpu.util",
			ValueType: "0",
			Units:     "%",
			LastValue: "25.5",
			Hosts: []zabbix.Host{
				{HostID: "100", Name: "web-server-01"},
			},
		},
		{
			ItemID:    "2",
			HostID:    "100",
			Name:      "Memory available",
			Key:       "vm.memory.available",
			ValueType: "3",
			Units:     "B",
			LastValue: "1073741824",
			Hosts: []zabbix.Host{
				{HostID: "100", Name: "web-server-01"},
			},
		},
		{
			ItemID:    "3",
			HostID:    "200",
			Name:      "Load average (1m)",
			Key:       "system.load[avg1]",
			ValueType: "0",
			Units:     "",
			LastValue: "0.5",
			Hosts: []zabbix.Host{
				{HostID: "200", Name: "db-server-01"},
			},
		},
	}

	categories := []string{"system.cpu", "system.load", "vm.memory"}

	tree := BuildTree(items, categories)

	// Should have 2 hosts
	if len(tree.Roots) != 2 {
		t.Errorf("Expected 2 host roots, got %d", len(tree.Roots))
	}

	// Total items should be 3
	if tree.ItemCount() != 3 {
		t.Errorf("Expected 3 items, got %d", tree.ItemCount())
	}

	// Initially only host nodes should be visible (collapsed)
	if tree.VisibleCount() != 2 {
		t.Errorf("Expected 2 visible nodes (hosts), got %d", tree.VisibleCount())
	}
}

func TestExpandCollapse(t *testing.T) {
	items := []zabbix.Item{
		{
			ItemID:    "1",
			HostID:    "100",
			Name:      "CPU utilization",
			Key:       "system.cpu.util",
			ValueType: "0",
			Hosts: []zabbix.Host{
				{HostID: "100", Name: "test-host"},
			},
		},
	}

	tree := BuildTree(items, []string{"system.cpu"})

	// Initially collapsed - only host visible
	if tree.VisibleCount() != 1 {
		t.Errorf("Expected 1 visible (collapsed), got %d", tree.VisibleCount())
	}

	// Expand all
	tree.ExpandAll()
	// Should see: host -> category -> item = 3 nodes
	if tree.VisibleCount() != 3 {
		t.Errorf("Expected 3 visible (expanded), got %d", tree.VisibleCount())
	}

	// Collapse all
	tree.CollapseAll()
	if tree.VisibleCount() != 1 {
		t.Errorf("Expected 1 visible (collapsed), got %d", tree.VisibleCount())
	}
}

func TestExtractCategory(t *testing.T) {
	tests := []struct {
		key        string
		categories []string
		expected   string
	}{
		{"system.cpu.util", []string{"system.cpu", "vm.memory"}, "CPU"},
		{"vm.memory.available", []string{"system.cpu", "vm.memory"}, "Memory"},
		{"net.if.in[eth0]", []string{"system.cpu", "net.if"}, "Network"},
		{"some.unknown.key", []string{"system.cpu"}, "Other"},
		{"system.load[avg1]", []string{"system.load"}, "Load"},
	}

	for _, tt := range tests {
		result := ExtractCategory(tt.key, tt.categories)
		if result != tt.expected {
			t.Errorf("ExtractCategory(%q, %v) = %q, want %q",
				tt.key, tt.categories, result, tt.expected)
		}
	}
}

func TestFormatCategoryName(t *testing.T) {
	tests := []struct {
		prefix   string
		expected string
	}{
		{"system.cpu", "CPU"},
		{"system.load", "Load"},
		{"vm.memory", "Memory"},
		{"vfs.fs", "Filesystem"},
		{"net.if", "Network"},
		{"proc", "Processes"},
		{"custom.metric", "Metric"},
	}

	for _, tt := range tests {
		result := FormatCategoryName(tt.prefix)
		if result != tt.expected {
			t.Errorf("FormatCategoryName(%q) = %q, want %q",
				tt.prefix, result, tt.expected)
		}
	}
}

func TestToggleNode(t *testing.T) {
	items := []zabbix.Item{
		{
			ItemID:    "1",
			HostID:    "100",
			Name:      "CPU utilization",
			Key:       "system.cpu.util",
			ValueType: "0",
			Hosts: []zabbix.Host{
				{HostID: "100", Name: "test-host"},
			},
		},
	}

	tree := BuildTree(items, []string{"system.cpu"})

	hostNode := tree.Roots[0]
	if !hostNode.Collapsed {
		t.Error("Host node should start collapsed")
	}

	// Toggle to expand
	tree.ToggleNode(hostNode.ID)
	if hostNode.Collapsed {
		t.Error("Host node should be expanded after toggle")
	}
	// Now visible: host + category = 2
	if tree.VisibleCount() != 2 {
		t.Errorf("Expected 2 visible after expanding host, got %d", tree.VisibleCount())
	}

	// Toggle again to collapse
	tree.ToggleNode(hostNode.ID)
	if !hostNode.Collapsed {
		t.Error("Host node should be collapsed after second toggle")
	}
	if tree.VisibleCount() != 1 {
		t.Errorf("Expected 1 visible after collapsing host, got %d", tree.VisibleCount())
	}
}

func TestNew(t *testing.T) {
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
	if m.tree == nil {
		t.Error("Expected tree to be initialized")
	}
}

func TestSetSize(t *testing.T) {
	m := New(testStyles())
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("Expected height 50, got %d", m.height)
	}
}

func TestSetFocused(t *testing.T) {
	m := New(testStyles())

	m.SetFocused(true)
	if !m.focused {
		t.Error("Expected focused to be true")
	}

	m.SetFocused(false)
	if m.focused {
		t.Error("Expected focused to be false")
	}
}

func TestScroll(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})
	m.SetSize(80, 5) // Small height to force scrolling

	// Expand to have scrollable content
	m.ExpandAll()

	// Make sure we have enough content to scroll
	visibleCount := m.VisibleNodeCount()
	visibleRows := m.visibleRows()

	if visibleCount <= visibleRows {
		t.Skip("Not enough content to test scrolling")
	}

	// Scroll down
	m.Scroll(1)
	if m.offset < 1 {
		t.Errorf("Expected offset >= 1 after scrolling down, got %d", m.offset)
	}

	// Scroll up past beginning
	m.Scroll(-100)
	if m.offset != 0 {
		t.Errorf("Expected offset 0 after scrolling past beginning, got %d", m.offset)
	}
}

func TestVisibleNodeCount(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})

	// Initially collapsed - only hosts visible
	count := m.VisibleNodeCount()
	if count != 2 {
		t.Errorf("Expected 2 visible nodes (collapsed), got %d", count)
	}

	// Expand all
	m.ExpandAll()
	count = m.VisibleNodeCount()
	if count < 2 {
		t.Errorf("Expected more than 2 visible nodes (expanded), got %d", count)
	}
}

func TestClickNode(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})
	m.SetSize(80, 20)
	m.SetFocused(true)

	// Click first node (should select and expand host)
	_ = m.ClickNode(0)
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0 after click, got %d", m.cursor)
	}

	// The host should now be expanded (more nodes visible)
	if m.VisibleNodeCount() <= 1 {
		t.Error("Expected host to expand after click")
	}

	// Click on invalid index should return nil
	cmd := m.ClickNode(-1)
	if cmd != nil {
		t.Error("Expected nil command for invalid index")
	}

	cmd = m.ClickNode(1000)
	if cmd != nil {
		t.Error("Expected nil command for out-of-bounds index")
	}
}

func TestView(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})
	m.SetSize(80, 20)

	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain header
	if !containsString(view, "GRAPHS") {
		t.Error("Expected view to contain GRAPHS header")
	}
}

func TestViewZeroSize(t *testing.T) {
	m := New(testStyles())
	m.SetSize(5, 3) // Too small

	view := m.View()
	if view != "" {
		t.Error("Expected empty view for zero-size")
	}
}

func TestMoveUpDown(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})
	m.SetSize(80, 20)
	m.SetFocused(true)

	// Initially at 0
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	// Move down
	m.MoveDown()
	if m.cursor != 1 {
		t.Errorf("Expected cursor at 1 after MoveDown, got %d", m.cursor)
	}

	// Move up
	m.MoveUp()
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0 after MoveUp, got %d", m.cursor)
	}

	// Move up at top should stay at 0
	m.MoveUp()
	if m.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", m.cursor)
	}
}

func TestPageUpDown(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})
	m.SetSize(80, 20)
	m.SetFocused(true)
	m.ExpandAll()

	// Page down
	m.PageDown()
	if m.cursor == 0 {
		t.Error("Expected cursor to move after PageDown")
	}

	// Page up
	savedCursor := m.cursor
	m.PageUp()
	if m.cursor >= savedCursor {
		t.Error("Expected cursor to decrease after PageUp")
	}
}

func TestGoToTopBottom(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})
	m.SetSize(80, 20)
	m.SetFocused(true)

	// Go to bottom
	m.GoToBottom()
	expectedBottom := m.VisibleNodeCount() - 1
	if m.cursor != expectedBottom {
		t.Errorf("Expected cursor at %d after GoToBottom, got %d", expectedBottom, m.cursor)
	}

	// Go to top
	m.GoToTop()
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0 after GoToTop, got %d", m.cursor)
	}
	if m.offset != 0 {
		t.Errorf("Expected offset at 0 after GoToTop, got %d", m.offset)
	}
}

func TestSelected(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})
	m.SetSize(80, 20)

	node := m.Selected()
	if node == nil {
		t.Fatal("Expected selected node")
	}
	if node.Type != NodeTypeHost {
		t.Errorf("Expected host node, got %v", node.Type)
	}
}

func TestSelectedItem(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})
	m.SetSize(80, 20)

	// Initially selected is host, not item
	item := m.SelectedItem()
	if item != nil {
		t.Error("Expected nil item when host is selected")
	}

	// Expand and navigate to item
	m.ExpandAll()
	// Move to an item node (skip host and category nodes)
	for i := 0; i < m.VisibleNodeCount(); i++ {
		m.cursor = i
		if node := m.Selected(); node != nil && node.Type == NodeTypeItem {
			item = m.SelectedItem()
			if item == nil {
				t.Error("Expected item when item node is selected")
			}
			break
		}
	}
}

func TestSetHostLoading(t *testing.T) {
	m := New(testStyles())

	m.SetHostLoading("host1", true)
	if !m.IsHostLoading("host1") {
		t.Error("Expected host1 to be loading")
	}

	m.SetHostLoading("host1", false)
	if m.IsHostLoading("host1") {
		t.Error("Expected host1 to not be loading")
	}
}

func TestMergeHistory(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})

	history := map[string][]zabbix.History{
		"1": {
			{ItemID: "1", Value: "50.5", Clock: "1700000000"},
			{ItemID: "1", Value: "51.0", Clock: "1700000060"},
		},
	}

	m.MergeHistory(history)

	retrieved := m.GetHistory("1")
	if len(retrieved) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(retrieved))
	}
}

func TestGetHostItems(t *testing.T) {
	m := New(testStyles())
	items := createTestItems()
	m.SetItems(items, []string{"system.cpu", "vm.memory"})

	hostItems := m.GetHostItems("100")
	if len(hostItems) == 0 {
		t.Error("Expected items for host 100")
	}

	// Non-existent host
	hostItems = m.GetHostItems("999")
	if len(hostItems) != 0 {
		t.Error("Expected no items for non-existent host")
	}
}

// Helper function to create test items
func createTestItems() []zabbix.Item {
	return []zabbix.Item{
		{
			ItemID:    "1",
			HostID:    "100",
			Name:      "CPU utilization",
			Key:       "system.cpu.util",
			ValueType: "0",
			Units:     "%",
			LastValue: "25.5",
			Hosts: []zabbix.Host{
				{HostID: "100", Name: "web-server-01"},
			},
		},
		{
			ItemID:    "2",
			HostID:    "100",
			Name:      "Memory available",
			Key:       "vm.memory.available",
			ValueType: "3",
			Units:     "B",
			LastValue: "1073741824",
			Hosts: []zabbix.Host{
				{HostID: "100", Name: "web-server-01"},
			},
		},
		{
			ItemID:    "3",
			HostID:    "200",
			Name:      "Load average (1m)",
			Key:       "system.cpu.load",
			ValueType: "0",
			Units:     "",
			LastValue: "0.5",
			Hosts: []zabbix.Host{
				{HostID: "200", Name: "db-server-01"},
			},
		},
	}
}

// Helper to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
