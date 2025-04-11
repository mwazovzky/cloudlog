package formatter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
)

// LokiFormatterOption defines a function to configure the LokiFormatter
type LokiFormatterOption func(*LokiFormatter)

// Loki provides a namespace for Loki-specific formatter options to avoid naming conflicts
// while still allowing intuitive naming when used with dot notation: Loki.WithTimeFormat(...)
var Loki lokiOptions

type lokiOptions struct{}

// WithLabelKeys specifies keys from the log entry that should be added as Loki stream labels
func WithLabelKeys(keys ...string) LokiFormatterOption {
	return func(lf *LokiFormatter) {
		lf.labelKeys = keys
	}
}

// WithTimeFormat sets the time format for timestamp fields
func (lokiOptions) WithTimeFormat(format string) LokiFormatterOption {
	return func(f *LokiFormatter) {
		f.SetTimeFormat(format)
	}
}

// WithTimestampField sets the field name for timestamp
func (lokiOptions) WithTimestampField(field string) LokiFormatterOption {
	return func(f *LokiFormatter) {
		f.SetFieldName("timestamp", field)
	}
}

// WithLevelField sets the field name for log level
func (lokiOptions) WithLevelField(field string) LokiFormatterOption {
	return func(f *LokiFormatter) {
		f.SetFieldName("level", field)
	}
}

// WithJobField sets the field name for job
func (lokiOptions) WithJobField(field string) LokiFormatterOption {
	return func(f *LokiFormatter) {
		f.SetFieldName("job", field)
	}
}

// LokiFormatter formats log entries directly in Loki protocol format
// It handles both the Loki protocol structure and internal log content formatting
type LokiFormatter struct {
	labelKeys      []string // Keys to extract from log entry and use as labels
	timeFormat     string   // Format for timestamp values
	timestampField string   // Field name for timestamp in JSON output
	levelField     string   // Field name for log level in JSON output
	jobField       string   // Field name for job in JSON output
}

// Implement FormatterConfig interface

// SetTimeFormat sets the time format used for timestamp fields
func (f *LokiFormatter) SetTimeFormat(format string) {
	f.timeFormat = format
}

// SetFieldName sets a custom field name for a standard field
func (f *LokiFormatter) SetFieldName(field, name string) {
	switch field {
	case "timestamp":
		f.timestampField = name
	case "level":
		f.levelField = name
	case "job":
		f.jobField = name
	}
}

// NewLokiFormatter creates a new LokiFormatter with the given options
func NewLokiFormatter(options ...LokiFormatterOption) *LokiFormatter {
	formatter := &LokiFormatter{
		labelKeys:      []string{},   // No extra labels by default
		timeFormat:     time.RFC3339, // Standard time format
		timestampField: "timestamp",  // Default field names follow standard conventions
		levelField:     "level",
		jobField:       "job",
	}

	for _, option := range options {
		option(formatter)
	}

	return formatter
}

// Format converts a log entry directly into a client.LokiEntry
func (f *LokiFormatter) Format(entry LogEntry) (client.LokiEntry, error) {
	// Format the log content using internal JSON formatting
	content, err := f.formatContent(entry)
	if err != nil {
		return client.LokiEntry{}, fmt.Errorf("%w: failed to format log content: %v", errors.ErrInvalidFormat, err)
	}

	// Extract labels from the entry's KeyVals if they match the labelKeys
	labels := map[string]string{
		"job": entry.Job,
	}

	for _, key := range f.labelKeys {
		if value, exists := entry.KeyVals[key]; exists {
			// Convert the value to string for Loki labels
			labels[key] = fmt.Sprintf("%v", value)
		}
	}

	// Create and return a Loki entry
	return client.LokiEntry{
		Streams: []client.LokiStream{
			{
				Stream: labels,
				Values: [][]string{
					{
						fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
						string(content),
					},
				},
			},
		},
	}, nil
}

// FormatBatch converts multiple log entries for the same job into a single Loki entry
func (f *LokiFormatter) FormatBatch(job string, entries []LogEntry) (client.LokiEntry, error) {
	if len(entries) == 0 {
		// Return an empty entry if there are no entries
		return client.LokiEntry{
			Streams: []client.LokiStream{},
		}, nil
	}

	// Extract labels that should be common for all entries
	// We'll use the job and any label keys found in the first entry
	labels := map[string]string{
		"job": job,
	}

	// If entries exist, extract label keys from the first entry
	if len(entries) > 0 {
		for _, key := range f.labelKeys {
			if value, exists := entries[0].KeyVals[key]; exists {
				// Convert the value to string for Loki labels
				labels[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	// Create a single stream for all entries
	stream := client.LokiStream{
		Stream: labels,
		Values: make([][]string, 0, len(entries)),
	}

	// Add each entry to the stream values
	for _, entry := range entries {
		content, err := f.formatContent(entry)
		if err != nil {
			// Skip entries that fail to format
			continue
		}

		stream.Values = append(stream.Values, []string{
			fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
			string(content),
		})
	}

	// Create and return the Loki entry with all entries in one stream
	return client.LokiEntry{
		Streams: []client.LokiStream{stream},
	}, nil
}

// formatContent provides JSON formatting for the log content
func (f *LokiFormatter) formatContent(entry LogEntry) ([]byte, error) {
	data := make(map[string]interface{})

	// Add timestamp
	data[f.timestampField] = entry.Timestamp.Format(f.timeFormat)
	// Add job and level
	data[f.jobField] = entry.Job
	data[f.levelField] = entry.Level

	// Add all key-value pairs
	for k, v := range entry.KeyVals {
		data[k] = v
	}

	// Marshal to JSON
	return json.Marshal(data)
}
