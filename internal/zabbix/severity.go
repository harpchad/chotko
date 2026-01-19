// Package zabbix provides a client for the Zabbix JSON-RPC API.
package zabbix

// Severity represents a Zabbix severity level.
type Severity int

// Severity level constants matching Zabbix API values.
const (
	SeverityNotClassified Severity = 0
	SeverityInformation   Severity = 1
	SeverityWarning       Severity = 2
	SeverityAverage       Severity = 3
	SeverityHigh          Severity = 4
	SeverityDisaster      Severity = 5
)

// severityNames maps severity levels to human-readable names.
var severityNames = map[Severity]string{
	SeverityNotClassified: "Not classified",
	SeverityInformation:   "Information",
	SeverityWarning:       "Warning",
	SeverityAverage:       "Average",
	SeverityHigh:          "High",
	SeverityDisaster:      "Disaster",
}

// severityShortCodes maps severity levels to single-letter codes.
var severityShortCodes = map[Severity]string{
	SeverityNotClassified: "N",
	SeverityInformation:   "I",
	SeverityWarning:       "W",
	SeverityAverage:       "A",
	SeverityHigh:          "H",
	SeverityDisaster:      "D",
}

// severityEmojis maps severity levels to emoji representations.
var severityEmojis = map[Severity]string{
	SeverityDisaster:      "💥",
	SeverityHigh:          "🔥",
	SeverityAverage:       "🚨",
	SeverityWarning:       "⚠️",
	SeverityInformation:   "ⓘ",
	SeverityNotClassified: "○",
}

// Name returns the human-readable name for the severity level.
// For unknown/invalid severity values, returns "Not classified".
func (s Severity) Name() string {
	if name, ok := severityNames[s]; ok {
		return name
	}
	return "Not classified"
}

// ShortCode returns the single-letter code for the severity level.
func (s Severity) ShortCode() string {
	if code, ok := severityShortCodes[s]; ok {
		return code
	}
	return "?"
}

// Emoji returns the emoji representation for the severity level.
func (s Severity) Emoji() string {
	if emoji, ok := severityEmojis[s]; ok {
		return emoji
	}
	return "○"
}

// String implements fmt.Stringer.
func (s Severity) String() string {
	return s.Name()
}
