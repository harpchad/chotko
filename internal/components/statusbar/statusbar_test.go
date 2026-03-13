package statusbar

import (
	"strings"
	"testing"

	"github.com/harpchad/chotko/internal/theme"
)

func TestModel_HideAcknowledgedFilterIsActive(t *testing.T) {
	t.Parallel()

	m := New(theme.NewStyles(theme.DefaultTheme()))
	m.SetWidth(120)
	m.SetFilter(0, "", true)

	if !m.HasActiveFilter() {
		t.Fatal("HasActiveFilter() = false, want true")
	}

	view := m.View()
	if !strings.Contains(view, "unacked only") {
		t.Fatalf("View() = %q, want filter indicator", view)
	}
}
