package zabbix

import (
	"testing"
	"time"
)

func TestProblem_SeverityInt(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		want     int
	}{
		{"severity 0", "0", 0},
		{"severity 1", "1", 1},
		{"severity 2", "2", 2},
		{"severity 3", "3", 3},
		{"severity 4", "4", 4},
		{"severity 5", "5", 5},
		{"empty string", "", 0},
		{"invalid string", "invalid", 0},
		{"negative", "-1", -1}, // strconv.Atoi parses negative numbers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Problem{Severity: tt.severity}
			if got := p.SeverityInt(); got != tt.want {
				t.Errorf("SeverityInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestProblem_IsAcknowledged(t *testing.T) {
	tests := []struct {
		name         string
		acknowledged string
		want         bool
	}{
		{"acknowledged", "1", true},
		{"not acknowledged", "0", false},
		{"empty", "", false},
		{"other value", "2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Problem{Acknowledged: tt.acknowledged}
			if got := p.IsAcknowledged(); got != tt.want {
				t.Errorf("IsAcknowledged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProblem_IsSuppressed(t *testing.T) {
	tests := []struct {
		name       string
		suppressed string
		want       bool
	}{
		{"suppressed", "1", true},
		{"not suppressed", "0", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Problem{Suppressed: tt.suppressed}
			if got := p.IsSuppressed(); got != tt.want {
				t.Errorf("IsSuppressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProblem_StartTime(t *testing.T) {
	tests := []struct {
		name  string
		clock string
		want  time.Time
	}{
		{
			name:  "valid timestamp",
			clock: "1700000000",
			want:  time.Unix(1700000000, 0),
		},
		{
			name:  "zero timestamp",
			clock: "0",
			want:  time.Time{}, // zero time for invalid clock
		},
		{
			name:  "empty string",
			clock: "",
			want:  time.Time{}, // zero time for invalid clock
		},
		{
			name:  "invalid string",
			clock: "invalid",
			want:  time.Time{}, // zero time for invalid clock
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Problem{Clock: tt.clock}
			got := p.StartTime()
			if !got.Equal(tt.want) {
				t.Errorf("StartTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProblem_Duration(t *testing.T) {
	// Use a fixed "now" time for testing
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour).Unix()

	p := Problem{Clock: itoa(oneHourAgo)}
	got := p.Duration()

	// Duration should be approximately 1 hour (allow some slack for test execution)
	if got < 59*time.Minute || got > 61*time.Minute {
		t.Errorf("Duration() = %v, want ~1 hour", got)
	}
}

func TestProblem_DurationString(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		clock    int64
		contains string
	}{
		{
			name:     "seconds",
			clock:    now.Add(-30 * time.Second).Unix(),
			contains: "s",
		},
		{
			name:     "minutes",
			clock:    now.Add(-5 * time.Minute).Unix(),
			contains: "m",
		},
		{
			name:     "hours",
			clock:    now.Add(-2 * time.Hour).Unix(),
			contains: "h",
		},
		{
			name:     "days",
			clock:    now.Add(-48 * time.Hour).Unix(),
			contains: "d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Problem{Clock: itoa(tt.clock)}
			got := p.DurationString()
			if !containsStr(got, tt.contains) {
				t.Errorf("DurationString() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestProblem_HostName(t *testing.T) {
	tests := []struct {
		name  string
		hosts []Host
		want  string
	}{
		{
			name: "with Name field",
			hosts: []Host{
				{Name: "Display Name", Host: "hostname"},
			},
			want: "Display Name",
		},
		{
			name: "fallback to Host field",
			hosts: []Host{
				{Name: "", Host: "hostname"},
			},
			want: "hostname",
		},
		{
			name:  "empty hosts",
			hosts: []Host{},
			want:  "Unknown", // Returns "Unknown" when no hosts
		},
		{
			name:  "nil hosts",
			hosts: nil,
			want:  "Unknown", // Returns "Unknown" when no hosts
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Problem{Hosts: tt.hosts}
			if got := p.HostName(); got != tt.want {
				t.Errorf("HostName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProblem_HostIP(t *testing.T) {
	tests := []struct {
		name  string
		hosts []Host
		want  string
	}{
		{
			name: "with interface",
			hosts: []Host{
				{Interfaces: []Interface{{IP: "192.168.1.1"}}},
			},
			want: "192.168.1.1",
		},
		{
			name: "empty interfaces",
			hosts: []Host{
				{Interfaces: []Interface{}},
			},
			want: "",
		},
		{
			name:  "empty hosts",
			hosts: []Host{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Problem{Hosts: tt.hosts}
			if got := p.HostIP(); got != tt.want {
				t.Errorf("HostIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHost_IsMonitored(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"monitored", "0", true},
		{"not monitored", "1", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Host{Status: tt.status}
			if got := h.IsMonitored(); got != tt.want {
				t.Errorf("IsMonitored() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHost_InMaintenance(t *testing.T) {
	tests := []struct {
		name              string
		maintenanceStatus string
		want              bool
	}{
		{"in maintenance", "1", true},
		{"not in maintenance", "0", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Host{MaintenanceStatus: tt.maintenanceStatus}
			if got := h.InMaintenance(); got != tt.want {
				t.Errorf("InMaintenance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHost_IsAvailable(t *testing.T) {
	tests := []struct {
		name            string
		activeAvailable string
		want            int
	}{
		{"unknown (0)", "0", 0},
		{"available (1)", "1", 1},
		{"unavailable (2)", "2", 2},
		{"empty", "", 0},
		{"invalid", "invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Host{ActiveAvailable: tt.activeAvailable}
			if got := h.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHost_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		host     Host
		wantName string
	}{
		{
			name:     "with Name",
			host:     Host{Name: "Display Name", Host: "hostname"},
			wantName: "Display Name",
		},
		{
			name:     "fallback to Host",
			host:     Host{Name: "", Host: "hostname"},
			wantName: "hostname",
		},
		{
			name:     "both empty",
			host:     Host{Name: "", Host: ""},
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.host.DisplayName(); got != tt.wantName {
				t.Errorf("DisplayName() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

// Helper functions

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var s string
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
