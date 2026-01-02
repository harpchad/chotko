// Package ignores manages locally ignored alert patterns.
// Ignored alerts are filtered from display but still exist in Zabbix.
package ignores

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Rule represents an ignore rule for a specific host+trigger combination.
type Rule struct {
	HostID      string    `yaml:"host_id"`
	HostName    string    `yaml:"host_name"`
	TriggerID   string    `yaml:"trigger_id"`
	TriggerName string    `yaml:"trigger_name"`
	Created     time.Time `yaml:"created"`
}

// List manages a collection of ignore rules with persistence.
type List struct {
	Ignores []Rule `yaml:"ignores"`
	path    string
	mu      sync.RWMutex
}

// Load loads the ignore list from the config directory.
// Creates an empty list if the file doesn't exist.
func Load(configDir string) (*List, error) {
	path := filepath.Join(configDir, "ignores.yaml")
	l := &List{
		Ignores: []Rule{},
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, return empty list
			return l, nil
		}
		return nil, fmt.Errorf("failed to read ignores file: %w", err)
	}

	if err := yaml.Unmarshal(data, l); err != nil {
		// Log warning but return empty list to avoid blocking startup
		return l, fmt.Errorf("failed to parse ignores file: %w", err)
	}

	return l, nil
}

// Save persists the ignore list to disk.
func (l *List) Save() error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("failed to marshal ignores: %w", err)
	}

	// Add header comment
	header := "# Chotko Ignored Alerts\n# Alerts matching these host+trigger pairs are hidden from display.\n\n"
	content := header + string(data)

	if err := os.WriteFile(l.path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write ignores file: %w", err)
	}

	return nil
}

// Add adds a new ignore rule. Returns error if the rule already exists.
func (l *List) Add(rule Rule) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check for duplicates
	for _, existing := range l.Ignores {
		if existing.HostID == rule.HostID && existing.TriggerID == rule.TriggerID {
			return fmt.Errorf("already ignored")
		}
	}

	// Set creation time if not set
	if rule.Created.IsZero() {
		rule.Created = time.Now()
	}

	l.Ignores = append(l.Ignores, rule)
	return nil
}

// Remove removes an ignore rule by index (1-based for user display).
// Returns the removed rule and true if successful, empty rule and false otherwise.
func (l *List) Remove(index int) (Rule, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Convert to 0-based index
	idx := index - 1
	if idx < 0 || idx >= len(l.Ignores) {
		return Rule{}, false
	}

	removed := l.Ignores[idx]
	l.Ignores = append(l.Ignores[:idx], l.Ignores[idx+1:]...)
	return removed, true
}

// IsIgnored returns true if the given host+trigger combination is ignored.
func (l *List) IsIgnored(hostID, triggerID string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, rule := range l.Ignores {
		if rule.HostID == hostID && rule.TriggerID == triggerID {
			return true
		}
	}
	return false
}

// Rules returns a copy of all ignore rules.
func (l *List) Rules() []Rule {
	l.mu.RLock()
	defer l.mu.RUnlock()

	rules := make([]Rule, len(l.Ignores))
	copy(rules, l.Ignores)
	return rules
}

// Len returns the number of ignore rules.
func (l *List) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.Ignores)
}

// Path returns the file path for the ignore list.
func (l *List) Path() string {
	return l.path
}
