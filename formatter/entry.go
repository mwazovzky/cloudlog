package formatter

import (
	"time"
)

// LogEntry represents a single log entry with all its metadata
type LogEntry struct {
	Timestamp time.Time
	Job       string
	Level     string
	KeyVals   map[string]interface{}
}

// NewLogEntry creates a new LogEntry with the current timestamp and parsed key-value pairs
func NewLogEntry(job string, level string, keyvals ...interface{}) LogEntry {
	entry := LogEntry{
		Timestamp: time.Now(),
		Job:       job,
		Level:     level,
		KeyVals:   make(map[string]interface{}),
	}

	// Parse key-value pairs
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok || i+1 >= len(keyvals) {
			continue // Skip invalid pairs
		}
		entry.KeyVals[key] = keyvals[i+1]
	}

	return entry
}
