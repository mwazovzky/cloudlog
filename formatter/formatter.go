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
}

