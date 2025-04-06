// Package formatter provides different formatting options for log entries.
// It includes implementations for JSON and human-readable string formats,
// both with customizable options.
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

// Formatter defines the interface for formatting log entries
type Formatter interface {
	// Format converts a LogEntry into a formatted string representation
	Format(entry LogEntry) ([]byte, error)
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

// Remove the duplicate declaration of NewStringFormatter
// It's already defined in string_formatter.go

// Create LogEntryOption type for consistency
type LogEntryOption func(*LogEntry)

// Add builder-style construction
func NewLogEntryWithOptions(options ...LogEntryOption) LogEntry {
	entry := LogEntry{
		Timestamp: time.Now(),
		KeyVals:   make(map[string]interface{}),
	}

	for _, option := range options {
		option(&entry)
	}

	return entry
}

// WithJob sets the job name
func WithJob(job string) LogEntryOption {
	return func(e *LogEntry) {
		e.Job = job
	}
}

// WithLevel sets the log level
func WithLevel(level string) LogEntryOption {
	return func(e *LogEntry) {
		e.Level = level
	}
}
