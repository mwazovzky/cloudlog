package formatter

import (
	"encoding/json"
	"time"
)

// LokiFormatterOption defines a function to configure the LokiFormatter
type LokiFormatterOption func(*LokiFormatter)

// Loki provides a namespace for Loki-specific formatter options
var Loki lokiOptions

type lokiOptions struct{}

// WithTimeFormat sets the time format for timestamp fields
func (lokiOptions) WithTimeFormat(format string) LokiFormatterOption {
	return func(f *LokiFormatter) {
		f.timeFormat = format
	}
}

// LokiFormatter formats log entries as JSON for Loki
type LokiFormatter struct {
	timeFormat string
}

// NewLokiFormatter creates a new LokiFormatter with the given options
func NewLokiFormatter(options ...LokiFormatterOption) *LokiFormatter {
	formatter := &LokiFormatter{
		timeFormat: time.RFC3339,
	}

	for _, option := range options {
		option(formatter)
	}

	return formatter
}

// Format converts a log entry to JSON bytes
func (f *LokiFormatter) Format(entry LogEntry) ([]byte, error) {
	data := make(map[string]interface{}, len(entry.KeyVals)+3)

	data["timestamp"] = entry.Timestamp.Format(f.timeFormat)
	data["job"] = entry.Job
	data["level"] = entry.Level

	for k, v := range entry.KeyVals {
		data[k] = v
	}

	return json.Marshal(data)
}
