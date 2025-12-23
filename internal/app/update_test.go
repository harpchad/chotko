package app

import (
	"testing"

	"github.com/harpchad/chotko/internal/config"
	"github.com/harpchad/chotko/internal/theme"
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
			updatedModel := newModel.(Model)

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

	updatedModel := newModel.(Model)

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
