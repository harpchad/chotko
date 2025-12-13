package graphs

import (
	"testing"

	"github.com/harpchad/chotko/internal/zabbix"
)

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
