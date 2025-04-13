// Package formatter provides different formatting options for log entries.
// It supports two main formats:
//
// 1. LokiFormatter - For Grafana Loki protocol output, optimized for cloud-based logging
// 2. StringFormatter - For human-readable console output
//
// Each formatter supports customization via options for field names, time formats,
// and other formatting preferences.
package formatter

import "github.com/mwazovzky/cloudlog/client"

// Formatter defines the interface for formatting log entries
type Formatter interface {
	// Format converts a log entry to a Loki entry
	Format(entry LogEntry) (client.LokiEntry, error)

	// FormatBatch converts multiple log entries for the same job into a single Loki entry
	// All entries are assumed to be for the same job
	FormatBatch(job string, entries []LogEntry) (client.LokiEntry, error)
}

// Additional interfaces and types may be defined below
