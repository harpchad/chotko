package format

import (
	"testing"
)

func TestValue(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		units string
		want  string
	}{
		{"percentage", 75.5, "%", "75.5%"},
		{"bytes small", 512, "B", "512"},
		{"bytes KB", 1536, "B", "1.5K"},
		{"bytes MB", 1572864, "B", "1.5M"},
		{"bytes per sec", 1536, "Bps", "1.5Kps"},
		{"seconds small", 0.5, "s", "500ms"},
		{"seconds large", 2.5, "s", "2.50s"},
		{"large number", 1500000, "", "1.5M"},
		{"medium number", 1500, "", "1.5K"},
		{"integer", 42, "", "42"},
		{"decimal", 3.14159, "", "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Value(tt.value, tt.units)
			if got != tt.want {
				t.Errorf("Value(%v, %q) = %q, want %q", tt.value, tt.units, got, tt.want)
			}
		})
	}
}

func TestBytes(t *testing.T) {
	tests := []struct {
		bytes float64
		want  string
	}{
		{0, "0"},
		{512, "512"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1073741824, "1.0G"},
		{1099511627776, "1.0T"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Bytes(tt.bytes)
			if got != tt.want {
				t.Errorf("Bytes(%v) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestYAxisValue(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		units string
		want  string
	}{
		{"percentage", 75, "%", "75%"},
		{"bytes", 1073741824, "B", "1.0G"},
		{"tera", 1e12, "", "1.0T"},
		{"giga", 1e9, "", "1.0G"},
		{"mega", 1e6, "", "1.0M"},
		{"kilo", 1e3, "", "1.0K"},
		{"normal", 50, "", "50"},
		{"decimal", 0.05, "", "0.05"},
		{"tiny", 0.001, "", "0"},
		{"negative giga", -1e9, "", "-1.0G"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := YAxisValue(tt.value, tt.units)
			if got != tt.want {
				t.Errorf("YAxisValue(%v, %q) = %q, want %q", tt.value, tt.units, got, tt.want)
			}
		})
	}
}
