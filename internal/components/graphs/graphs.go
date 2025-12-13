package graphs

import (
	"fmt"
	"strings"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
}

// New creates a new graphs tree model.
func New(styles *theme.Styles) Model {
	return Model{
		styles:     styles,
		tree:       NewTree(),
		sparklines: make(map[string]string),
		history:    make(map[string][]zabbix.History),
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

// SetItems updates the items and rebuilds the tree.
func (m *Model) SetItems(items []zabbix.Item, categories []string) {
	m.tree = BuildTree(items, categories)
	m.cursor = 0
	m.offset = 0
	m.sparklines = make(map[string]string)
}

// SetHistory updates history data for items and regenerates sparklines.
func (m *Model) SetHistory(history map[string][]zabbix.History) {
	m.history = history
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
func (m *Model) Toggle() {
	if node := m.Selected(); node != nil {
		m.tree.ToggleNode(node.ID)
	}
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
			m.Toggle()
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
		b.WriteString(row)
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
	var rowContent string
	switch node.Type {
	case NodeTypeHost:
		rowContent = m.renderHostNode(node)
	case NodeTypeCategory:
		rowContent = m.renderCategoryNode(node)
	case NodeTypeItem:
		rowContent = m.renderItemNode(node)
	}

	// Calculate available width for content
	prefixWidth := len(indent) + len(indicator)
	contentWidth := m.width - prefixWidth - 4 // margin

	if len(rowContent) > contentWidth {
		rowContent = rowContent[:contentWidth-3] + "..."
	}

	// Combine parts
	row := indent + indicator + rowContent

	if selected {
		return m.styles.AlertSelected.Width(m.width - 2).Render(row)
	}

	return m.styles.AlertNormal.Width(m.width - 2).Render(row)
}

// renderHostNode renders a host node.
func (m Model) renderHostNode(node *TreeNode) string {
	// Count items under this host
	itemCount := 0
	for _, cat := range node.Children {
		itemCount += len(cat.Children)
	}
	return fmt.Sprintf("%s (%d)", node.Name, itemCount)
}

// renderCategoryNode renders a category node.
func (m Model) renderCategoryNode(node *TreeNode) string {
	return fmt.Sprintf("%s (%d)", node.Name, len(node.Children))
}

// renderItemNode renders an item node with value and sparkline.
func (m Model) renderItemNode(node *TreeNode) string {
	if node.Item == nil {
		return node.Name
	}

	item := node.Item

	// Format value with units
	value := formatValue(item.LastValueFloat(), item.Units)

	// Get sparkline if available
	spark := ""
	if sl, ok := m.sparklines[item.ItemID]; ok {
		spark = "  " + sl
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

	// Build the row
	nameStyle := m.styles.AlertHost.Width(nameWidth)
	valueStyle := m.styles.AlertDuration.Width(valueWidth).Align(lipgloss.Right)
	sparkStyle := m.styles.Subtle

	return nameStyle.Render(name) + valueStyle.Render(value) + sparkStyle.Render(spark)
}

// formatValue formats a numeric value with appropriate units.
func formatValue(value float64, units string) string {
	// Handle percentage
	if units == "%" {
		return fmt.Sprintf("%.1f%%", value)
	}

	// Handle bytes
	if units == "B" || units == "Bps" {
		return formatBytes(value) + strings.TrimPrefix(units, "B")
	}

	// Handle time units
	if units == "s" {
		if value < 1 {
			return fmt.Sprintf("%.0fms", value*1000)
		}
		return fmt.Sprintf("%.2fs", value)
	}

	// Default formatting
	if value >= 1000000 {
		return fmt.Sprintf("%.1fM", value/1000000)
	}
	if value >= 1000 {
		return fmt.Sprintf("%.1fK", value/1000)
	}
	if value == float64(int64(value)) {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

// formatBytes formats bytes to human-readable form.
func formatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", bytes/div, "KMGTPE"[exp])
}

// GetHistory returns the history data for an item.
func (m Model) GetHistory(itemID string) []zabbix.History {
	return m.history[itemID]
}
