package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Check default values
	if cfg.Server.URL != "" {
		t.Errorf("expected empty server URL, got %q", cfg.Server.URL)
	}
	if cfg.Display.RefreshInterval != 30 {
		t.Errorf("expected refresh interval 30, got %d", cfg.Display.RefreshInterval)
	}
	if cfg.Display.MinSeverity != 0 {
		t.Errorf("expected min severity 0, got %d", cfg.Display.MinSeverity)
	}
	if cfg.Display.Theme != "nord" {
		t.Errorf("expected theme 'nord', got %q", cfg.Display.Theme)
	}
}

func TestDir(t *testing.T) {
	tests := []struct {
		name       string
		xdgConfig  string
		wantSuffix string
		setEnv     bool
	}{
		{
			name:       "with XDG_CONFIG_HOME",
			xdgConfig:  "/custom/config",
			wantSuffix: "/custom/config/chotko",
			setEnv:     true,
		},
		{
			name:       "without XDG_CONFIG_HOME",
			xdgConfig:  "",
			wantSuffix: ".config/chotko",
			setEnv:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore XDG_CONFIG_HOME
			oldXDG := os.Getenv("XDG_CONFIG_HOME")
			defer func() {
				if oldXDG != "" {
					_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
				} else {
					_ = os.Unsetenv("XDG_CONFIG_HOME")
				}
			}()

			if tt.setEnv {
				if tt.xdgConfig != "" {
					_ = os.Setenv("XDG_CONFIG_HOME", tt.xdgConfig)
				} else {
					_ = os.Unsetenv("XDG_CONFIG_HOME")
				}
			}

			got := Dir()

			if tt.xdgConfig != "" {
				if got != tt.wantSuffix {
					t.Errorf("Dir() = %q, want %q", got, tt.wantSuffix)
				}
			} else {
				// Just check it ends with the expected suffix
				if !contains(got, tt.wantSuffix) {
					t.Errorf("Dir() = %q, want suffix %q", got, tt.wantSuffix)
				}
			}
		})
	}
}

func TestPath(t *testing.T) {
	path := Path()

	if !contains(path, "chotko") {
		t.Errorf("Path() = %q, expected to contain 'chotko'", path)
	}
	if !contains(path, "config.yaml") {
		t.Errorf("Path() = %q, expected to contain 'config.yaml'", path)
	}
}

func TestLoadFromFile_Success(t *testing.T) {
	// Create temp config file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := `
server:
  url: "https://zabbix.example.com"
auth:
  token: "test-token"
display:
  refresh_interval: 60
  min_severity: 3
  theme: "dracula"
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if cfg.Server.URL != "https://zabbix.example.com" {
		t.Errorf("Server.URL = %q, want %q", cfg.Server.URL, "https://zabbix.example.com")
	}
	if cfg.Auth.Token != "test-token" {
		t.Errorf("Auth.Token = %q, want %q", cfg.Auth.Token, "test-token")
	}
	if cfg.Display.RefreshInterval != 60 {
		t.Errorf("Display.RefreshInterval = %d, want 60", cfg.Display.RefreshInterval)
	}
	if cfg.Display.MinSeverity != 3 {
		t.Errorf("Display.MinSeverity = %d, want 3", cfg.Display.MinSeverity)
	}
	if cfg.Display.Theme != "dracula" {
		t.Errorf("Display.Theme = %q, want %q", cfg.Display.Theme, "dracula")
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadFromFile() expected error for nonexistent file")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error = %q, expected to contain 'not found'", err.Error())
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := `
server:
  url: [invalid yaml
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadFromFile(configPath)
	if err == nil {
		t.Error("LoadFromFile() expected error for invalid YAML")
	}
	if !contains(err.Error(), "parse") {
		t.Errorf("error = %q, expected to contain 'parse'", err.Error())
	}
}

func TestLoadFromFile_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Only specify some values, others should use defaults
	content := `
server:
  url: "https://zabbix.example.com"
auth:
  token: "test-token"
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Specified values
	if cfg.Server.URL != "https://zabbix.example.com" {
		t.Errorf("Server.URL = %q, want %q", cfg.Server.URL, "https://zabbix.example.com")
	}

	// Default values should be applied
	if cfg.Display.RefreshInterval != 30 {
		t.Errorf("Display.RefreshInterval = %d, want default 30", cfg.Display.RefreshInterval)
	}
	if cfg.Display.Theme != "nord" {
		t.Errorf("Display.Theme = %q, want default 'nord'", cfg.Display.Theme)
	}
}

func TestExists(t *testing.T) {
	// Save and restore XDG_CONFIG_HOME to control path
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if oldXDG != "" {
			_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	t.Run("file exists", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.Setenv("XDG_CONFIG_HOME", dir)

		// Create the config file
		configDir := filepath.Join(dir, "chotko")
		if err := os.MkdirAll(configDir, 0o750); err != nil {
			t.Fatalf("failed to create config dir: %v", err)
		}
		configPath := filepath.Join(configDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte("server:\n  url: test"), 0o600); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		if !Exists() {
			t.Error("Exists() = false, want true")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.Setenv("XDG_CONFIG_HOME", dir)

		if Exists() {
			t.Error("Exists() = true, want false")
		}
	})
}

func TestSaveToFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "subdir", "config.yaml")

	cfg := &Config{
		Server: ServerConfig{URL: "https://zabbix.example.com"},
		Auth:   AuthConfig{Token: "test-token"},
		Display: DisplayConfig{
			RefreshInterval: 45,
			MinSeverity:     2,
			Theme:           "gruvbox",
		},
	}

	if err := SaveToFile(cfg, configPath); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("SaveToFile() did not create file")
	}

	// Load it back and verify
	loaded, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if loaded.Server.URL != cfg.Server.URL {
		t.Errorf("Server.URL = %q, want %q", loaded.Server.URL, cfg.Server.URL)
	}
	if loaded.Auth.Token != cfg.Auth.Token {
		t.Errorf("Auth.Token = %q, want %q", loaded.Auth.Token, cfg.Auth.Token)
	}
	if loaded.Display.RefreshInterval != cfg.Display.RefreshInterval {
		t.Errorf("Display.RefreshInterval = %d, want %d", loaded.Display.RefreshInterval, cfg.Display.RefreshInterval)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with token",
			config: &Config{
				Server:  ServerConfig{URL: "https://zabbix.example.com"},
				Auth:    AuthConfig{Token: "test-token"},
				Display: DisplayConfig{RefreshInterval: 30},
			},
			wantErr: false,
		},
		{
			name: "valid with username/password",
			config: &Config{
				Server:  ServerConfig{URL: "https://zabbix.example.com"},
				Auth:    AuthConfig{Username: "admin", Password: "password"},
				Display: DisplayConfig{RefreshInterval: 30},
			},
			wantErr: false,
		},
		{
			name: "missing server URL",
			config: &Config{
				Auth:    AuthConfig{Token: "test-token"},
				Display: DisplayConfig{RefreshInterval: 30},
			},
			wantErr: true,
			errMsg:  "server URL",
		},
		{
			name: "missing auth",
			config: &Config{
				Server:  ServerConfig{URL: "https://zabbix.example.com"},
				Display: DisplayConfig{RefreshInterval: 30},
			},
			wantErr: true,
			errMsg:  "authentication",
		},
		{
			name: "missing password with username",
			config: &Config{
				Server:  ServerConfig{URL: "https://zabbix.example.com"},
				Auth:    AuthConfig{Username: "admin"},
				Display: DisplayConfig{RefreshInterval: 30},
			},
			wantErr: true,
			errMsg:  "authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %q, want to contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestConfig_Validate_RejectsLowRefreshInterval(t *testing.T) {
	cfg := &Config{
		Server:  ServerConfig{URL: "https://zabbix.example.com"},
		Auth:    AuthConfig{Token: "test-token"},
		Display: DisplayConfig{RefreshInterval: 2}, // Less than minimum
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should return error for refresh interval < 5")
	}

	if !contains(err.Error(), "refresh interval") {
		t.Errorf("Validate() error = %q, want to contain 'refresh interval'", err.Error())
	}
}

func TestConfig_Validate_Severity(t *testing.T) {
	tests := []struct {
		name     string
		severity int
		wantErr  bool
	}{
		{"negative", -1, true},
		{"too high", 10, true},
		{"valid low", 0, false},
		{"valid high", 5, false},
		{"valid mid", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server:  ServerConfig{URL: "https://zabbix.example.com"},
				Auth:    AuthConfig{Token: "test-token"},
				Display: DisplayConfig{RefreshInterval: 30, MinSeverity: tt.severity},
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !contains(err.Error(), "severity") {
				t.Errorf("Validate() error = %q, want to contain 'severity'", err.Error())
			}
		})
	}
}

func TestConfig_UseToken(t *testing.T) {
	tests := []struct {
		name string
		auth AuthConfig
		want bool
	}{
		{
			name: "with token",
			auth: AuthConfig{Token: "test-token"},
			want: true,
		},
		{
			name: "without token",
			auth: AuthConfig{Username: "admin", Password: "password"},
			want: false,
		},
		{
			name: "empty auth",
			auth: AuthConfig{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Auth: tt.auth}
			if got := cfg.UseToken(); got != tt.want {
				t.Errorf("UseToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(s != "" && substr != "" && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
