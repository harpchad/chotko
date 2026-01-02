// Package graphs provides the graphs tree component for displaying Zabbix items.
package graphs

import (
	"sort"
	"strings"

	"github.com/harpchad/chotko/internal/zabbix"
)

// NodeType represents the type of tree node.
type NodeType int

// Node type constants.
const (
	NodeTypeHost NodeType = iota
	NodeTypeCategory
	NodeTypeItem
)

// TreeNode represents a node in the tree structure.
type TreeNode struct {
	ID        string       // Unique identifier
	Name      string       // Display name
	Type      NodeType     // Type of node
	Collapsed bool         // Whether children are hidden
	Children  []*TreeNode  // Child nodes
	Depth     int          // Nesting depth
	Item      *zabbix.Item // For item nodes, the actual item
	HostID    string       // Host ID for category/item nodes
	Category  string       // Category name for item nodes
}

// Tree represents a collapsible tree of hosts, categories, and items.
type Tree struct {
	Roots       []*TreeNode          // Top-level host nodes
	FlatList    []*TreeNode          // Flattened visible nodes (for rendering)
	AllNodes    map[string]*TreeNode // All nodes by ID
	ItemsByHost map[string][]zabbix.Item
}

// NewTree creates an empty tree.
func NewTree() *Tree {
	return &Tree{
		Roots:       make([]*TreeNode, 0),
		FlatList:    make([]*TreeNode, 0),
		AllNodes:    make(map[string]*TreeNode),
		ItemsByHost: make(map[string][]zabbix.Item),
	}
}

// BuildTree constructs a tree from items grouped by host and category.
// Items are organized as: Host -> Category -> Item
func BuildTree(items []zabbix.Item, categories []string) *Tree {
	tree := NewTree()

	// Group items by host
	hostItems := make(map[string][]zabbix.Item)
	hostNames := make(map[string]string)

	for _, item := range items {
		hostID := item.GetHostID()
		hostItems[hostID] = append(hostItems[hostID], item)
		if item.HostName() != "Unknown" {
			hostNames[hostID] = item.HostName()
		}
	}

	tree.ItemsByHost = hostItems

	// Sort host IDs by name for consistent ordering
	hostIDs := make([]string, 0, len(hostItems))
	for hid := range hostItems {
		hostIDs = append(hostIDs, hid)
	}
	sort.Slice(hostIDs, func(i, j int) bool {
		return hostNames[hostIDs[i]] < hostNames[hostIDs[j]]
	})

	// Build tree for each host
	for _, hostID := range hostIDs {
		items := hostItems[hostID]
		hostName := hostNames[hostID]
		if hostName == "" {
			hostName = "Host " + hostID
		}

		hostNode := &TreeNode{
			ID:        "host:" + hostID,
			Name:      hostName,
			Type:      NodeTypeHost,
			Collapsed: true, // Start collapsed
			Depth:     0,
			HostID:    hostID,
			Children:  make([]*TreeNode, 0),
		}
		tree.AllNodes[hostNode.ID] = hostNode

		// Group items by category
		categoryItems := make(map[string][]zabbix.Item)
		for _, item := range items {
			cat := ExtractCategory(item.Key, categories)
			categoryItems[cat] = append(categoryItems[cat], item)
		}

		// Sort categories
		catNames := make([]string, 0, len(categoryItems))
		for cat := range categoryItems {
			catNames = append(catNames, cat)
		}
		sort.Strings(catNames)

		// Build category nodes
		for _, catName := range catNames {
			catItems := categoryItems[catName]

			catNode := &TreeNode{
				ID:        "cat:" + hostID + ":" + catName,
				Name:      catName,
				Type:      NodeTypeCategory,
				Collapsed: true, // Start collapsed
				Depth:     1,
				HostID:    hostID,
				Category:  catName,
				Children:  make([]*TreeNode, 0),
			}
			tree.AllNodes[catNode.ID] = catNode

			// Sort items by name
			sort.Slice(catItems, func(i, j int) bool {
				return catItems[i].Name < catItems[j].Name
			})

			// Build item nodes
			for _, item := range catItems {
				itemCopy := item // Copy to avoid pointer issues
				itemNode := &TreeNode{
					ID:       "item:" + item.ItemID,
					Name:     item.Name,
					Type:     NodeTypeItem,
					Depth:    2,
					Item:     &itemCopy,
					HostID:   hostID,
					Category: catName,
				}
				tree.AllNodes[itemNode.ID] = itemNode
				catNode.Children = append(catNode.Children, itemNode)
			}

			hostNode.Children = append(hostNode.Children, catNode)
		}

		tree.Roots = append(tree.Roots, hostNode)
	}

	tree.RebuildFlatList()
	return tree
}

// ExtractCategory determines the category for an item based on its key.
// Returns the matching category prefix or "Other" if no match.
func ExtractCategory(key string, categories []string) string {
	for _, cat := range categories {
		if strings.HasPrefix(key, cat) {
			return FormatCategoryName(cat)
		}
	}
	return "Other"
}

// FormatCategoryName formats a category key prefix into a display name.
func FormatCategoryName(prefix string) string {
	// Map common prefixes to friendly names
	nameMap := map[string]string{
		"system.cpu":  "CPU",
		"system.load": "Load",
		"vm.memory":   "Memory",
		"vfs.fs":      "Filesystem",
		"net.if":      "Network",
		"proc":        "Processes",
	}

	if name, ok := nameMap[prefix]; ok {
		return name
	}

	// Fallback: capitalize first letter of last component
	parts := strings.Split(prefix, ".")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if last != "" {
			return strings.ToUpper(last[:1]) + last[1:]
		}
	}
	return prefix
}

// RebuildFlatList regenerates the flattened list of visible nodes.
func (t *Tree) RebuildFlatList() {
	t.FlatList = make([]*TreeNode, 0)
	for _, root := range t.Roots {
		t.flattenNode(root)
	}
}

// flattenNode recursively adds visible nodes to the flat list.
func (t *Tree) flattenNode(node *TreeNode) {
	t.FlatList = append(t.FlatList, node)
	if !node.Collapsed {
		for _, child := range node.Children {
			t.flattenNode(child)
		}
	}
}

// ToggleNode toggles the collapsed state of a node.
func (t *Tree) ToggleNode(id string) {
	if node, ok := t.AllNodes[id]; ok {
		if len(node.Children) > 0 {
			node.Collapsed = !node.Collapsed
			t.RebuildFlatList()
		}
	}
}

// ExpandAll expands all nodes.
func (t *Tree) ExpandAll() {
	for _, node := range t.AllNodes {
		if len(node.Children) > 0 {
			node.Collapsed = false
		}
	}
	t.RebuildFlatList()
}

// CollapseAll collapses all nodes.
func (t *Tree) CollapseAll() {
	for _, node := range t.AllNodes {
		if len(node.Children) > 0 {
			node.Collapsed = true
		}
	}
	t.RebuildFlatList()
}

// GetNode returns a node by ID.
func (t *Tree) GetNode(id string) *TreeNode {
	return t.AllNodes[id]
}

// ItemCount returns the total number of items in the tree.
func (t *Tree) ItemCount() int {
	count := 0
	for _, node := range t.AllNodes {
		if node.Type == NodeTypeItem {
			count++
		}
	}
	return count
}

// VisibleCount returns the number of visible nodes.
func (t *Tree) VisibleCount() int {
	return len(t.FlatList)
}

// GetVisibleNode returns the node at the given visible index.
func (t *Tree) GetVisibleNode(index int) *TreeNode {
	if index >= 0 && index < len(t.FlatList) {
		return t.FlatList[index]
	}
	return nil
}

// FindNodeIndex returns the visible index of a node by ID.
func (t *Tree) FindNodeIndex(id string) int {
	for i, node := range t.FlatList {
		if node.ID == id {
			return i
		}
	}
	return -1
}
