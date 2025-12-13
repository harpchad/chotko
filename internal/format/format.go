// Package format provides shared formatting utilities for values and units.
package format

import (
	"fmt"
	"strings"
)

// Value formats a numeric value with appropriate units.
// Handles percentages, bytes, time units, and generic values.
func Value(value float64, units string) string {
	// Handle percentage
	if units == "%" {
		return fmt.Sprintf("%.1f%%", value)
	}

	// Handle bytes
	if units == "B" || units == "Bps" {
		return Bytes(value) + strings.TrimPrefix(units, "B")
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

// Bytes formats bytes to human-readable form using binary prefixes (1024).
func Bytes(bytes float64) string {
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

// BytesShort formats bytes to short human-readable form for axis labels.
func BytesShort(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", bytes/div, "KMGTP"[exp])
}

// YAxisValue formats a value for Y axis labels with appropriate SI suffixes.
func YAxisValue(value float64, units string) string {
	// Handle bytes specially - use binary prefixes (Ki, Mi, Gi)
	if units == "B" || units == "Bps" {
		return BytesShort(value)
	}

	// Handle percentages
	if units == "%" {
		return fmt.Sprintf("%.0f%%", value)
	}

	// For other values, use SI prefixes
	absVal := value
	if absVal < 0 {
		absVal = -absVal
	}

	switch {
	case absVal >= 1e12:
		return fmt.Sprintf("%.1fT", value/1e12)
	case absVal >= 1e9:
		return fmt.Sprintf("%.1fG", value/1e9)
	case absVal >= 1e6:
		return fmt.Sprintf("%.1fM", value/1e6)
	case absVal >= 1e3:
		return fmt.Sprintf("%.1fK", value/1e3)
	case absVal >= 1:
		return fmt.Sprintf("%.0f", value)
	case absVal >= 0.01:
		return fmt.Sprintf("%.2f", value)
	default:
		return fmt.Sprintf("%.0f", value)
	}
}
