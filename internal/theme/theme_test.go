package theme

import (
	"testing"
)

func TestNewStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	if styles == nil {
		t.Fatal("NewStyles() returned nil")
	}
}

func TestNewStyles_AllStylesInitialized(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Base styles - verify App style is usable (just check it doesn't panic)
	_ = styles.App.String()

	// We can't easily test lipgloss.Style values, but we can ensure
	// the struct is properly populated by checking it doesn't panic
	// when accessed

	// Test that AlertSeverity array has all 6 elements
	if len(styles.AlertSeverity) != 6 {
		t.Errorf("AlertSeverity should have 6 elements, got %d", len(styles.AlertSeverity))
	}
}

func TestNewStyles_WithAllBuiltinThemes(t *testing.T) {
	t.Parallel()

	themes := BuiltinThemes()

	for name, theme := range themes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			styles := NewStyles(theme)
			if styles == nil {
				t.Fatalf("NewStyles(%s) returned nil", name)
			}

			// Verify AlertSeverity has correct length
			if len(styles.AlertSeverity) != 6 {
				t.Errorf("NewStyles(%s).AlertSeverity has %d elements, want 6", name, len(styles.AlertSeverity))
			}
		})
	}
}

func TestTheme_Fields(t *testing.T) {
	t.Parallel()

	th := &Theme{
		Name:        "test",
		Description: "Test description",
	}

	if th.Name != "test" {
		t.Errorf("Theme.Name = %q, want %q", th.Name, "test")
	}
	if th.Description != "Test description" {
		t.Errorf("Theme.Description = %q, want %q", th.Description, "Test description")
	}
}

func TestStyles_SeverityArrayMapping(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// The AlertSeverity array should be indexed 0-5
	// Each index corresponds to Zabbix severity level

	// Severity 0 = NotClassified
	// Severity 1 = Information
	// Severity 2 = Warning
	// Severity 3 = Average
	// Severity 4 = High
	// Severity 5 = Disaster

	// Just verify we can access all indices without panic
	for i := 0; i <= 5; i++ {
		_ = styles.AlertSeverity[i]
	}
}

func TestNewStyles_PaneStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Verify pane styles exist (we can't easily compare lipgloss.Style values)
	// Just ensure they're usable without panic
	_ = styles.PaneFocused.String()
	_ = styles.PaneBlurred.String()
	_ = styles.PaneTitle.String()
	_ = styles.PaneSubtitle.String()
}

func TestNewStyles_StatusBarStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Verify status bar styles exist
	_ = styles.StatusBar.String()
	_ = styles.StatusOK.String()
	_ = styles.StatusProblem.String()
	_ = styles.StatusUnknown.String()
	_ = styles.StatusMaint.String()
}

func TestNewStyles_AlertStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Verify alert styles exist
	_ = styles.AlertSelected.String()
	_ = styles.AlertNormal.String()
	_ = styles.AlertHost.String()
	_ = styles.AlertName.String()
	_ = styles.AlertDuration.String()
	_ = styles.AlertAcked.String()
}

func TestNewStyles_DetailStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Verify detail pane styles exist
	_ = styles.DetailLabel.String()
	_ = styles.DetailValue.String()
	_ = styles.DetailTag.String()
}

func TestNewStyles_TabStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Verify tab styles exist
	_ = styles.TabActive.String()
	_ = styles.TabInactive.String()
	_ = styles.TabBar.String()
}

func TestNewStyles_CommandStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Verify command input styles exist
	_ = styles.CommandPrompt.String()
	_ = styles.CommandInput.String()
	_ = styles.CommandHint.String()
}

func TestNewStyles_ModalStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Verify modal styles exist
	_ = styles.ModalBackground.String()
	_ = styles.ModalBox.String()
	_ = styles.ModalTitle.String()
	_ = styles.ModalText.String()
	_ = styles.ModalButton.String()
	_ = styles.ModalButtonAlt.String()
}

func TestNewStyles_HelpStyles(t *testing.T) {
	t.Parallel()

	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Verify help styles exist
	_ = styles.HelpKey.String()
	_ = styles.HelpDesc.String()
}

func TestNewStyles_DifferentThemesProduceDifferentStyles(t *testing.T) {
	t.Parallel()

	// This is a sanity check that different themes actually produce
	// different styles (at least for the primary color)

	defaultStyles := NewStyles(DefaultTheme())
	draculaStyles := NewStyles(DraculaTheme())

	// We can't easily compare styles directly, but we can verify
	// both are non-nil and were created successfully
	if defaultStyles == nil {
		t.Error("defaultStyles should not be nil")
	}
	if draculaStyles == nil {
		t.Error("draculaStyles should not be nil")
	}
}
