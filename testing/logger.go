// Package testing provides utilities for testing code that uses the cloudlog library.
// The TestLogger captures logs for inspection in test assertions.
package testing

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
)

// TestLogger is designed for use in test environments
type TestLogger struct {
	mu        sync.Mutex
	logs      []TestLogEntry
	formatter formatter.Formatter
}

// TestLogEntry represents a captured log entry for testing
type TestLogEntry struct {
	Timestamp time.Time
	Level     string
	Job       string
	Message   string
	Data      map[string]interface{}
	Formatted []byte // The formatted log entry
}

// NewTestLogger creates a logger that captures logs for testing
func NewTestLogger() *TestLogger {
	return &TestLogger{
		logs:      make([]TestLogEntry, 0),
		formatter: formatter.NewJSONFormatter(), // Updated from NewJsonFormatter to NewJSONFormatter
	}
}

// Send captures log entries for testing
func (t *TestLogger) Send(job string, formatted []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(formatted, &data); err != nil {
		return errors.FormatError(err, "failed to unmarshal log entry for testing")
	}

	entry := TestLogEntry{
		Timestamp: time.Now(),
		Job:       job,
		Formatted: formatted,
		Data:      make(map[string]interface{}),
	}

	// Extract common fields
	if level, ok := data["level"].(string); ok {
		entry.Level = level
		delete(data, "level")
	}

	if msg, ok := data["message"].(string); ok {
		entry.Message = msg
		delete(data, "message")
	}

	// Copy remaining fields
	for k, v := range data {
		entry.Data[k] = v
	}

	// Store the log entry
	t.mu.Lock()
	t.logs = append(t.logs, entry)
	t.mu.Unlock()

	return nil
}

// Logs returns all captured log entries
func (t *TestLogger) Logs() []TestLogEntry {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Return a copy to avoid race conditions
	result := make([]TestLogEntry, len(t.logs))
	copy(result, t.logs)
	return result
}

// LogsOfLevel returns all logs of the specified level
func (t *TestLogger) LogsOfLevel(level string) []TestLogEntry {
	t.mu.Lock()
	defer t.mu.Unlock()

	result := make([]TestLogEntry, 0)
	for _, log := range t.logs {
		if log.Level == level {
			result = append(result, log)
		}
	}
	return result
}

// Clear removes all captured logs
func (t *TestLogger) Clear() {
	t.mu.Lock()
	t.logs = make([]TestLogEntry, 0)
	t.mu.Unlock()
}

// ContainsMessage checks if a message containing the given text was logged
func (t *TestLogger) ContainsMessage(text string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, log := range t.logs {
		if log.Message == text {
			return true
		}
	}
	return false
}

// ContainsEntry checks if a log matching the criteria exists
func (t *TestLogger) ContainsEntry(level, text string, keyvals ...interface{}) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Build criteria map
	criteria := make(map[string]interface{})
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			if key, ok := keyvals[i].(string); ok {
				criteria[key] = keyvals[i+1]
			}
		}
	}

	for _, log := range t.logs {
		if level != "" && log.Level != level {
			continue
		}

		if text != "" && log.Message != text {
			continue
		}

		// Check key-values
		match := true
		for k, v := range criteria {
			if log.Data[k] != v {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}
