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

// Loki provides a namespace for Loki-specific formatter options
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
		f.timeFormat = format
	}
}

// LokiFormatter formats log entries directly in Loki protocol format
type LokiFormatter struct {
	labelKeys  []string // Keys to extract from log entry and use as labels
	timeFormat string   // Format for timestamp values
}

// NewLokiFormatter creates a new LokiFormatter with the given options
func NewLokiFormatter(options ...LokiFormatterOption) *LokiFormatter {
	formatter := &LokiFormatter{
		labelKeys:  []string{},
		timeFormat: time.RFC3339,
	}

	for _, option := range options {
		option(formatter)
	}

	return formatter
}

// Format converts a log entry directly into a client.LokiEntry
func (f *LokiFormatter) Format(entry LogEntry) (client.LokiEntry, error) {
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
			labels[key] = fmt.Sprintf("%v", value)
		}
	}

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

// formatContent provides JSON formatting for the log content
func (f *LokiFormatter) formatContent(entry LogEntry) ([]byte, error) {
	data := make(map[string]interface{})

	data["timestamp"] = entry.Timestamp.Format(f.timeFormat)
	data["job"] = entry.Job
	data["level"] = entry.Level

	for k, v := range entry.KeyVals {
		data[k] = v
	}

	return json.Marshal(data)
}
