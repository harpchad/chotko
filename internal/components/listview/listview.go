// Package listview provides shared list navigation and scrolling logic
// for list-based UI components.
package listview

// State holds the navigation state for a list view.
type State struct {
	cursor  int  // Current cursor position in filtered list
	offset  int  // Scroll offset for viewport
	width   int  // Available width
	height  int  // Available height (excluding borders)
	focused bool // Whether this list has focus
	count   int  // Total number of items (after filtering)
}

// New creates a new list view state.
func New() *State {
	return &State{}
}

// SetSize updates the dimensions.
func (s *State) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.ensureVisible()
}

// SetCount updates the item count and adjusts cursor if needed.
func (s *State) SetCount(count int) {
	s.count = count
	if s.cursor >= count && count > 0 {
		s.cursor = count - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	s.ensureVisible()
}

// SetFocused sets the focus state.
func (s *State) SetFocused(focused bool) {
	s.focused = focused
}

// Focused returns whether the list is focused.
func (s *State) Focused() bool {
	return s.focused
}

// Cursor returns the current cursor position.
func (s *State) Cursor() int {
	return s.cursor
}

// SetCursor sets the cursor position.
func (s *State) SetCursor(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos >= s.count && s.count > 0 {
		pos = s.count - 1
	}
	s.cursor = pos
	s.ensureVisible()
}

// Offset returns the current scroll offset.
func (s *State) Offset() int {
	return s.offset
}

// VisibleRows returns the number of visible rows.
func (s *State) VisibleRows() int {
	return s.height - 2 // Account for borders
}

// MoveUp moves the cursor up by one.
func (s *State) MoveUp() {
	if s.cursor > 0 {
		s.cursor--
		s.ensureVisible()
	}
}

// MoveDown moves the cursor down by one.
func (s *State) MoveDown() {
	if s.cursor < s.count-1 {
		s.cursor++
		s.ensureVisible()
	}
}

// PageUp moves the cursor up by one page.
func (s *State) PageUp() {
	s.cursor -= s.VisibleRows()
	if s.cursor < 0 {
		s.cursor = 0
	}
	s.ensureVisible()
}

// PageDown moves the cursor down by one page.
func (s *State) PageDown() {
	s.cursor += s.VisibleRows()
	if s.cursor >= s.count {
		s.cursor = s.count - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	s.ensureVisible()
}

// GoToTop moves the cursor to the first item.
func (s *State) GoToTop() {
	s.cursor = 0
	s.offset = 0
}

// GoToBottom moves the cursor to the last item.
func (s *State) GoToBottom() {
	if s.count > 0 {
		s.cursor = s.count - 1
		s.ensureVisible()
	}
}

// Scroll scrolls the view by delta lines.
func (s *State) Scroll(delta int) {
	s.offset += delta
	if s.offset < 0 {
		s.offset = 0
	}
	maxOffset := s.count - s.VisibleRows()
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.offset > maxOffset {
		s.offset = maxOffset
	}
}

// ensureVisible ensures the cursor is visible in the viewport.
func (s *State) ensureVisible() {
	visible := s.VisibleRows()
	if visible <= 0 {
		return
	}

	// Scroll up if cursor is above viewport
	if s.cursor < s.offset {
		s.offset = s.cursor
	}

	// Scroll down if cursor is below viewport
	if s.cursor >= s.offset+visible {
		s.offset = s.cursor - visible + 1
	}

	// Clamp offset
	if s.offset < 0 {
		s.offset = 0
	}
}

// Width returns the current width.
func (s *State) Width() int {
	return s.width
}

// Height returns the current height.
func (s *State) Height() int {
	return s.height
}

// Count returns the item count.
func (s *State) Count() int {
	return s.count
}
