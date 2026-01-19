package zabbix

import "testing"

func TestSeverity_Name(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"disaster", SeverityDisaster, "Disaster"},
		{"high", SeverityHigh, "High"},
		{"average", SeverityAverage, "Average"},
		{"warning", SeverityWarning, "Warning"},
		{"information", SeverityInformation, "Information"},
		{"not classified", SeverityNotClassified, "Not classified"},
		{"negative", Severity(-1), "Not classified"},
		{"too high", Severity(6), "Not classified"},
		{"way too high", Severity(100), "Not classified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.severity.Name(); got != tt.want {
				t.Errorf("Severity(%d).Name() = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

func TestSeverity_ShortCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"disaster", SeverityDisaster, "D"},
		{"high", SeverityHigh, "H"},
		{"average", SeverityAverage, "A"},
		{"warning", SeverityWarning, "W"},
		{"information", SeverityInformation, "I"},
		{"not classified", SeverityNotClassified, "N"},
		{"negative", Severity(-1), "?"},
		{"too high", Severity(6), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.severity.ShortCode(); got != tt.want {
				t.Errorf("Severity(%d).ShortCode() = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

func TestSeverity_Emoji(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"disaster", SeverityDisaster, "💥"},
		{"high", SeverityHigh, "🔥"},
		{"average", SeverityAverage, "🚨"},
		{"warning", SeverityWarning, "⚠️"},
		{"information", SeverityInformation, "ⓘ"},
		{"not classified", SeverityNotClassified, "○"},
		{"negative", Severity(-1), "○"},
		{"too high", Severity(6), "○"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.severity.Emoji(); got != tt.want {
				t.Errorf("Severity(%d).Emoji() = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

func TestSeverity_String(t *testing.T) {
	t.Parallel()

	// String() should return the same as Name()
	for sev := Severity(0); sev <= 5; sev++ {
		if got, want := sev.String(), sev.Name(); got != want {
			t.Errorf("Severity(%d).String() = %q, want %q", sev, got, want)
		}
	}
}

func TestSeverityConstants(t *testing.T) {
	t.Parallel()

	// Verify constants match Zabbix API values
	tests := []struct {
		name string
		sev  Severity
		want int
	}{
		{"not classified", SeverityNotClassified, 0},
		{"information", SeverityInformation, 1},
		{"warning", SeverityWarning, 2},
		{"average", SeverityAverage, 3},
		{"high", SeverityHigh, 4},
		{"disaster", SeverityDisaster, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := int(tt.sev); got != tt.want {
				t.Errorf("Severity constant %s = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}
