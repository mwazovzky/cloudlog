// Package formatter provides formatting options for log entries.
// It supports LokiFormatter (JSON for Loki) and StringFormatter (human-readable console output).
package formatter

// Formatter defines the interface for formatting log entry content
type Formatter interface {
	// Format converts a log entry to formatted bytes (JSON, string, etc.)
	Format(entry LogEntry) ([]byte, error)
}
