package app

import (
	"testing"

	"github.com/harpchad/chotko/internal/config"
	"github.com/harpchad/chotko/internal/ignores"
	"github.com/harpchad/chotko/internal/theme"
	"github.com/harpchad/chotko/internal/zabbix"
)

// testConfig returns a minimal config for testing.
func testConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			URL: "http://localhost/api_jsonrpc.php",
		},
		Auth: config.AuthConfig{
			Token: "test-token",
		},
		Display: config.DisplayConfig{
			RefreshInterval: 30,
			MinSeverity:     0,
		},
	}
}

// TestRefreshTickMsg_AlwaysHandled verifies that RefreshTickMsg is handled
// regardless of modal state, preventing the auto-refresh from stopping.
func TestRefreshTickMsg_AlwaysHandled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupModel func(*Model)
	}{
		{
			name:       "normal state",
			setupModel: func(_ *Model) {},
		},
		{
			name: "editor visible",
			setupModel: func(m *Model) {
				m.showEditor = true
			},
		},
		{
			name: "help modal visible",
			setupModel: func(m *Model) {
				m.showHelp = true
			},
		},
		{
			name: "error modal visible",
			setupModel: func(m *Model) {
				m.showError = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create fresh model for each test
			cfg := testConfig()
			thm := theme.DefaultTheme()
			m := New(cfg, thm)
			m.connected = true

			// Apply test-specific setup
			tt.setupModel(m)

			// Send RefreshTickMsg
			msg := RefreshTickMsg{}
			newModel, cmd := m.Update(msg)

			// Verify the model was updated (loading state should be set)
			updatedModel, ok := newModel.(Model)
			if !ok {
				t.Fatalf("expected Model type, got %T", newModel)
			}

			// The command should not be nil - it should contain the next tick
			if cmd == nil {
				t.Errorf("%s: RefreshTickMsg should return a command to schedule next tick", tt.name)
			}

			// Verify loading state was set (when connected)
			if !updatedModel.loading {
				t.Errorf("%s: loading should be true after RefreshTickMsg", tt.name)
			}
		})
	}
}

// TestRefreshTickMsg_NotConnected verifies refresh behavior when not connected.
func TestRefreshTickMsg_NotConnected(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	thm := theme.DefaultTheme()
	m := New(cfg, thm)
	m.connected = false // Not connected

	msg := RefreshTickMsg{}
	newModel, cmd := m.Update(msg)

	updatedModel, ok := newModel.(Model)
	if !ok {
		t.Fatalf("expected Model type, got %T", newModel)
	}

	// Command should still be returned (to keep the timer running)
	if cmd == nil {
		t.Error("RefreshTickMsg should always return a command for the next tick")
	}

	// Loading should not be set when not connected
	if updatedModel.loading {
		t.Error("loading should be false when not connected")
	}
}

// TestRefreshTickMsg_AlreadyLoading verifies refresh doesn't double-trigger.
func TestRefreshTickMsg_AlreadyLoading(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	thm := theme.DefaultTheme()
	m := New(cfg, thm)
	m.connected = true
	m.loading = true // Already loading

	msg := RefreshTickMsg{}
	_, cmd := m.Update(msg)

	// Command should still be returned (to keep the timer running)
	if cmd == nil {
		t.Error("RefreshTickMsg should always return a command for the next tick")
	}
}

func TestGetAlertCountsBySeverity_HideAcknowledgedAndIgnored(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.Display.HideAcknowledged = true
	thm := theme.DefaultTheme()
	m := New(cfg, thm)
	ignoreList, err := ignores.Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	m.ignoreList = ignoreList

	if err := m.ignoreList.Add(ignores.Rule{HostID: "host-2", TriggerID: "trigger-2"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	m.problems = []zabbix.Problem{
		{EventID: "1", Severity: "4", Acknowledged: "0", Object: ObjectTypeTrigger, ObjectID: "trigger-1", Hosts: []zabbix.Host{{HostID: "host-1"}}},
		{EventID: "2", Severity: "4", Acknowledged: "1", Object: ObjectTypeTrigger, ObjectID: "trigger-ack", Hosts: []zabbix.Host{{HostID: "host-ack"}}},
		{EventID: "3", Severity: "3", Acknowledged: "0", Object: ObjectTypeTrigger, ObjectID: "trigger-2", Hosts: []zabbix.Host{{HostID: "host-2"}}},
		{EventID: "4", Severity: "2", Acknowledged: "0", Object: ObjectTypeTrigger, ObjectID: "trigger-3", Hosts: []zabbix.Host{{HostID: "host-3"}}},
	}

	counts := m.getAlertCountsBySeverity()

	if counts[4] != 1 {
		t.Fatalf("counts[4] = %d, want 1", counts[4])
	}
	if counts[3] != 0 {
		t.Fatalf("counts[3] = %d, want 0", counts[3])
	}
	if counts[2] != 1 {
		t.Fatalf("counts[2] = %d, want 1", counts[2])
	}
}

func TestWindowTitle_ExcludesHiddenAcknowledgedAlerts(t *testing.T) {
	t.Parallel()

	emojiTitle := false
	cfg := testConfig()
	cfg.Display.HideAcknowledged = true
	cfg.Display.EmojiTitle = &emojiTitle
	thm := theme.DefaultTheme()
	m := New(cfg, thm)
	m.connected = true
	m.problems = []zabbix.Problem{
		{EventID: "1", Severity: "4", Acknowledged: "0"},
		{EventID: "2", Severity: "4", Acknowledged: "1"},
		{EventID: "3", Severity: "2", Acknowledged: "0"},
	}

	got := m.windowTitle()
	want := "chotko: [H:1 W:1]"
	if got != want {
		t.Fatalf("windowTitle() = %q, want %q", got, want)
	}
}

func TestHandleAcknowledgeResultMsg_UpdatesLocalStateBeforeReload(t *testing.T) {
	t.Parallel()

	emojiTitle := false
	cfg := testConfig()
	cfg.Display.HideAcknowledged = true
	cfg.Display.EmojiTitle = &emojiTitle
	thm := theme.DefaultTheme()
	m := New(cfg, thm)
	m.connected = true
	m.problems = []zabbix.Problem{{EventID: "1", Severity: "4", Acknowledged: "0", Name: "CPU high"}}
	m.alertList.SetProblems(m.problems)

	newModel, cmd := m.handleAcknowledgeResultMsg(AcknowledgeResultMsg{EventID: "1", Success: true})
	updated, ok := newModel.(Model)
	if !ok {
		t.Fatalf("expected Model type, got %T", newModel)
	}

	if cmd == nil {
		t.Fatal("handleAcknowledgeResultMsg() returned nil command")
	}
	if !updated.problems[0].IsAcknowledged() {
		t.Fatal("problem should be acknowledged locally before reload")
	}
	if selected := updated.alertList.Selected(); selected != nil {
		t.Fatalf("Selected() = %+v, want nil after hiding acknowledged alert", selected)
	}
	if got, want := updated.windowTitle(), "chotko: OK"; got != want {
		t.Fatalf("windowTitle() = %q, want %q", got, want)
	}
}

func TestHandleCloseResultMsg_UpdatesLocalStateBeforeReload(t *testing.T) {
	t.Parallel()

	emojiTitle := false
	cfg := testConfig()
	cfg.Display.EmojiTitle = &emojiTitle
	thm := theme.DefaultTheme()
	m := New(cfg, thm)
	m.connected = true
	m.problems = []zabbix.Problem{{EventID: "1", Severity: "4", Name: "CPU high"}}
	m.alertList.SetProblems(m.problems)

	newModel, cmd := m.handleCloseResultMsg(CloseResultMsg{EventID: "1", Success: true})
	updated, ok := newModel.(Model)
	if !ok {
		t.Fatalf("expected Model type, got %T", newModel)
	}

	if cmd == nil {
		t.Fatal("handleCloseResultMsg() returned nil command")
	}
	if len(updated.problems) != 0 {
		t.Fatalf("len(problems) = %d, want 0", len(updated.problems))
	}
	if selected := updated.alertList.Selected(); selected != nil {
		t.Fatalf("Selected() = %+v, want nil after close", selected)
	}
	if got, want := updated.windowTitle(), "chotko: OK"; got != want {
		t.Fatalf("windowTitle() = %q, want %q", got, want)
	}
}
