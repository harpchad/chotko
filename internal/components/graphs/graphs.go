package graphs

import (
	"fmt"
	"strings"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/harpchad/chotko/internal/format"
	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// Model represents the graphs tree component.
type Model struct {
	styles     *theme.Styles
	tree       *Tree
	cursor     int
	offset     int
	width      int
	height     int
	focused    bool
	textFilter string

	// Sparkline cache: itemID -> rendered sparkline string
	sparklines map[string]string
	// History data: itemID -> []History
	history map[string][]zabbix.History
	// Loading state: hostID -> is loading
	loadingHosts map[string]bool
}

// New creates a new graphs tree model.
func New(styles *theme.Styles) Model {
	return Model{
		styles:       styles,
		tree:         NewTree(),
		sparklines:   make(map[string]string),
		history:      make(map[string][]zabbix.History),
		loadingHosts: make(map[string]bool),
	}
}

// SetSize sets the component dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetFocused sets the focus state.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// SetItems updates the items and rebuilds the tree, preserving expanded state.
func (m *Model) SetItems(items []zabbix.Item, categories []string) {
	// Save current expanded state before rebuilding
	expandedNodes := make(map[string]bool)
	if m.tree != nil {
		for id, node := range m.tree.AllNodes {
			if !node.Collapsed {
				expandedNodes[id] = true
			}
		}
	}

	// Save current selection
	var selectedID string
	if selected := m.Selected(); selected != nil {
		selectedID = selected.ID
	}

	// Rebuild tree
	m.tree = BuildTree(items, categories)

	// Restore expanded state
	for id := range expandedNodes {
		if node, ok := m.tree.AllNodes[id]; ok {
			node.Collapsed = false
		}
	}
	m.tree.RebuildFlatList()

	// Restore selection if possible
	if selectedID != "" {
		if idx := m.tree.FindNodeIndex(selectedID); idx >= 0 {
			m.cursor = idx
			m.ensureVisible()
		}
	}

	m.sparklines = make(map[string]string)
}

// SetHistory updates history data for items and regenerates sparklines.
func (m *Model) SetHistory(history map[string][]zabbix.History) {
	m.history = history
	m.regenerateSparklines()
}

// MergeHistory adds history data for items without clearing existing data.
func (m *Model) MergeHistory(history map[string][]zabbix.History) {
	if m.history == nil {
		m.history = make(map[string][]zabbix.History)
	}
	for itemID, hist := range history {
		m.history[itemID] = hist
	}
	m.regenerateSparklines()
}

// regenerateSparklines creates sparkline strings for all items.
func (m *Model) regenerateSparklines() {
	m.sparklines = make(map[string]string)

	for itemID, hist := range m.history {
		if len(hist) == 0 {
			continue
		}

		// Extract values for sparkline
		values := make([]float64, len(hist))
		for i, h := range hist {
			values[i] = h.ValueFloat()
		}

		// Create sparkline (width 8, height 1)
		sl := sparkline.New(8, 1)
		sl.PushAll(values)
		sl.Draw()
		m.sparklines[itemID] = sl.View()
	}
}

// SetTextFilter sets the text filter (not implemented for tree yet).
func (m *Model) SetTextFilter(filter string) {
	m.textFilter = strings.ToLower(filter)
	// TODO: Implement filtering for tree view
}

// Selected returns the currently selected node.
func (m Model) Selected() *TreeNode {
	return m.tree.GetVisibleNode(m.cursor)
}

// SelectedItem returns the currently selected item (if the selection is an item node).
func (m Model) SelectedItem() *zabbix.Item {
	if node := m.Selected(); node != nil && node.Type == NodeTypeItem {
		return node.Item
	}
	return nil
}

// Count returns the total item count and visible node count.
func (m Model) Count() (total, visible int) {
	return m.tree.ItemCount(), m.tree.VisibleCount()
}

// MoveUp moves the cursor up.
func (m *Model) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
		m.ensureVisible()
	}
}

// MoveDown moves the cursor down.
func (m *Model) MoveDown() {
	if m.cursor < m.tree.VisibleCount()-1 {
		m.cursor++
		m.ensureVisible()
	}
}

// PageUp moves the cursor up by one page.
func (m *Model) PageUp() {
	m.cursor -= m.visibleRows()
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

// PageDown moves the cursor down by one page.
func (m *Model) PageDown() {
	m.cursor += m.visibleRows()
	if m.cursor >= m.tree.VisibleCount() {
		m.cursor = m.tree.VisibleCount() - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

// GoToTop moves the cursor to the first item.
func (m *Model) GoToTop() {
	m.cursor = 0
	m.offset = 0
}

// GoToBottom moves the cursor to the last item.
func (m *Model) GoToBottom() {
	m.cursor = max(0, m.tree.VisibleCount()-1)
	m.ensureVisible()
}

// Toggle toggles expand/collapse of the current node.
// Returns the host ID if a host was expanded and needs history loading, empty string otherwise.
func (m *Model) Toggle() string {
	node := m.Selected()
	if node == nil {
		return ""
	}

	wasCollapsed := node.Collapsed
	m.tree.ToggleNode(node.ID)

	// If a host node was expanded and we don't have history for it, return host ID
	if node.Type == NodeTypeHost && wasCollapsed && !m.HasHostHistory(node.HostID) {
		m.SetHostLoading(node.HostID, true)
		return node.HostID
	}

	return ""
}

// ExpandAll expands all nodes.
func (m *Model) ExpandAll() {
	m.tree.ExpandAll()
}

// CollapseAll collapses all nodes.
func (m *Model) CollapseAll() {
	m.tree.CollapseAll()
	m.cursor = 0
	m.offset = 0
}

// visibleRows returns the number of visible rows.
func (m Model) visibleRows() int {
	return m.height - 2 // Account for header and border
}

// ensureVisible ensures the cursor is visible in the viewport.
func (m *Model) ensureVisible() {
	visible := m.visibleRows()
	if visible <= 0 {
		return
	}

	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// HostExpandedMsg is sent when a host node is expanded and needs history loading.
type HostExpandedMsg struct {
	HostID string
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.MoveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.MoveDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup", "ctrl+u"))):
			m.PageUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown", "ctrl+d"))):
			m.PageDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys("home", "g"))):
			m.GoToTop()
		case key.Matches(msg, key.NewBinding(key.WithKeys("end", "G"))):
			m.GoToBottom()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
			if hostID := m.Toggle(); hostID != "" {
				// Host was expanded and needs history, send message
				return m, func() tea.Msg {
					return HostExpandedMsg{HostID: hostID}
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("E"))):
			m.ExpandAll()
		case key.Matches(msg, key.NewBinding(key.WithKeys("C"))):
			m.CollapseAll()
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	// Handle zero-size case
	if m.width < 10 || m.height < 5 {
		return ""
	}

	var b strings.Builder

	// Header
	total, visible := m.Count()
	header := fmt.Sprintf("GRAPHS (%d items", total)
	if visible != total {
		header += fmt.Sprintf(", %d visible", visible)
	}
	header += ")"
	b.WriteString(m.styles.PaneTitle.Render(header))
	b.WriteString("\n")

	// Calculate visible range
	visibleRows := m.visibleRows()
	if visibleRows < 1 {
		visibleRows = 1
	}

	endIdx := min(m.offset+visibleRows, m.tree.VisibleCount())

	// Render rows
	for i := m.offset; i < endIdx; i++ {
		node := m.tree.GetVisibleNode(i)
		if node == nil {
			continue
		}
		row := m.renderNode(node, i == m.cursor)
		// Mark row with zone for mouse click detection
		nodeID := fmt.Sprintf("graph_node_%d", i)
		b.WriteString(zone.Mark(nodeID, row))
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Pad remaining space
	rendered := endIdx - m.offset
	for i := rendered; i < visibleRows; i++ {
		b.WriteString("\n")
	}

	// Apply pane style
	content := b.String()
	if m.focused {
		return m.styles.PaneFocused.Width(m.width).Height(m.height).Render(content)
	}
	return m.styles.PaneBlurred.Width(m.width).Height(m.height).Render(content)
}

// renderNode renders a single tree node.
func (m Model) renderNode(node *TreeNode, selected bool) string {
	// Indentation
	indent := strings.Repeat("  ", node.Depth)

	// Collapse indicator
	var indicator string
	switch node.Type {
	case NodeTypeHost, NodeTypeCategory:
		if node.Collapsed {
			indicator = "▸ "
		} else {
			indicator = "▾ "
		}
	case NodeTypeItem:
		indicator = "  " // No indicator for leaf nodes
	}

	// Build row content based on node type
	// When selected, we render plain text to allow the selection background to show
	var rowContent string
	switch node.Type {
	case NodeTypeHost:
		rowContent = m.renderHostNode(node, selected)
	case NodeTypeCategory:
		rowContent = m.renderCategoryNode(node)
	case NodeTypeItem:
		rowContent = m.renderItemNode(node, selected)
	}

	// Combine parts
	row := indent + indicator + rowContent

	// Apply selection or normal style
	if selected {
		// Pad to full width for consistent highlight
		if len(row) < m.width-2 {
			row += strings.Repeat(" ", m.width-2-len(row))
		}
		return m.styles.AlertSelected.Render(row)
	}

	return m.styles.AlertNormal.Width(m.width - 2).Render(row)
}

// renderHostNode renders a host node.
func (m Model) renderHostNode(node *TreeNode, _ bool) string {
	// Count items under this host
	itemCount := 0
	for _, cat := range node.Children {
		itemCount += len(cat.Children)
	}

	name := fmt.Sprintf("%s (%d)", node.Name, itemCount)

	// Show loading indicator if history is being loaded
	if m.loadingHosts[node.HostID] {
		name += " ⟳"
	}

	return name
}

// renderCategoryNode renders a category node.
func (m Model) renderCategoryNode(node *TreeNode) string {
	return fmt.Sprintf("%s (%d)", node.Name, len(node.Children))
}

// renderItemNode renders an item node with value and sparkline.
func (m Model) renderItemNode(node *TreeNode, selected bool) string {
	if node.Item == nil {
		return node.Name
	}

	item := node.Item

	// Format value with units
	value := format.Value(item.LastValueFloat(), item.Units)

	// Get sparkline if available
	spark := ""
	if sl, ok := m.sparklines[item.ItemID]; ok {
		spark = sl
	}

	// Calculate widths
	valueWidth := 12
	sparkWidth := 10
	nameWidth := m.width - valueWidth - sparkWidth - 8 - node.Depth*2

	if nameWidth < 10 {
		nameWidth = 10
	}

	name := item.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	// When selected, render plain text to allow background highlighting
	if selected {
		// Pad value and spark to fixed widths for alignment
		valuePadded := fmt.Sprintf("%*s", valueWidth, value)
		sparkPadded := fmt.Sprintf("  %-*s", sparkWidth-2, spark)
		return fmt.Sprintf("%-*s%s%s", nameWidth, name, valuePadded, sparkPadded)
	}

	// Build the row with styles for non-selected items
	nameStyle := m.styles.AlertHost.Width(nameWidth)
	valueStyle := m.styles.AlertDuration.Width(valueWidth).Align(lipgloss.Right)
	sparkStyle := m.styles.Subtle.Width(sparkWidth)

	return nameStyle.Render(name) + valueStyle.Render(value) + "  " + sparkStyle.Render(spark)
}

// GetHistory returns the history data for an item.
func (m Model) GetHistory(itemID string) []zabbix.History {
	return m.history[itemID]
}

// GetHostItems returns all items belonging to a specific host.
func (m Model) GetHostItems(hostID string) []zabbix.Item {
	if m.tree == nil {
		return nil
	}
	return m.tree.ItemsByHost[hostID]
}

// HasHostHistory returns true if history has been loaded for any item of the host.
func (m Model) HasHostHistory(hostID string) bool {
	items := m.GetHostItems(hostID)
	for _, item := range items {
		if _, ok := m.history[item.ItemID]; ok {
			return true
		}
	}
	return false
}

// SetHostLoading sets the loading state for a host.
func (m *Model) SetHostLoading(hostID string, loading bool) {
	if m.loadingHosts == nil {
		m.loadingHosts = make(map[string]bool)
	}
	if loading {
		m.loadingHosts[hostID] = true
	} else {
		delete(m.loadingHosts, hostID)
	}
}

// IsHostLoading returns true if the host is currently loading history.
func (m Model) IsHostLoading(hostID string) bool {
	return m.loadingHosts[hostID]
}

// Scroll scrolls the list by delta lines (positive = down, negative = up).
func (m *Model) Scroll(delta int) {
	m.offset += delta
	if m.offset < 0 {
		m.offset = 0
	}
	maxOffset := m.tree.VisibleCount() - m.visibleRows()
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

// VisibleNodeCount returns the number of currently visible nodes.
func (m Model) VisibleNodeCount() int {
	return m.tree.VisibleCount()
}

// ClickNode handles a click on a tree node at the given visible index.
// It selects the node and toggles expansion if it's a parent node.
// Returns a command if history needs to be loaded.
func (m *Model) ClickNode(visibleIndex int) tea.Cmd {
	if visibleIndex < 0 || visibleIndex >= m.tree.VisibleCount() {
		return nil
	}

	m.cursor = visibleIndex
	m.ensureVisible()

	node := m.Selected()
	if node == nil {
		return nil
	}

	// Toggle expand/collapse for parent nodes (host or category)
	if node.Type == NodeTypeHost || node.Type == NodeTypeCategory {
		if hostID := m.Toggle(); hostID != "" {
			// Host was expanded and needs history
			return func() tea.Msg {
				return HostExpandedMsg{HostID: hostID}
			}
		}
	}

	return nil
}
