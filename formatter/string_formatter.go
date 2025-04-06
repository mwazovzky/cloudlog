package formatter

import (
	"fmt"
	"strings"
	"time"
)

// Rename for consistency with JSON naming in same package
type StringFormatterOption func(*StringFormatter)

// WithStringTimeFormat sets the time format for the StringFormatter
func WithStringTimeFormat(format string) StringFormatterOption {
	return func(f *StringFormatter) {
		f.timeFormat = format
	}
}

// WithKeyValueSeparator sets the separator between keys and values
func WithKeyValueSeparator(separator string) StringFormatterOption {
	return func(f *StringFormatter) {
		f.kvSeparator = separator
	}
}

// WithPairSeparator sets the separator between key-value pairs
func WithPairSeparator(separator string) StringFormatterOption {
	return func(f *StringFormatter) {
		f.pairSeparator = separator
	}
}

// StringFormatter formats log entries as human-readable strings
type StringFormatter struct {
	timeFormat    string
	kvSeparator   string
	pairSeparator string
}

// NewStringFormatter creates a new StringFormatter with the given options
func NewStringFormatter(options ...StringFormatterOption) *StringFormatter {
	formatter := &StringFormatter{
		timeFormat:    time.RFC3339,
		kvSeparator:   "=",
		pairSeparator: " ",
	}

	for _, option := range options {
		option(formatter)
	}

	return formatter
}

// Format converts a log entry into a human-readable string
func (f *StringFormatter) Format(entry LogEntry) ([]byte, error) {
	var sb strings.Builder

	// Add timestamp
	sb.WriteString(entry.Timestamp.Format(f.timeFormat))
	sb.WriteString(f.pairSeparator)

	// Add job
	sb.WriteString("job")
	sb.WriteString(f.kvSeparator)
	sb.WriteString(entry.Job)
	sb.WriteString(f.pairSeparator)

	// Add level if present
	if entry.Level != "" {
		sb.WriteString("level")
		sb.WriteString(f.kvSeparator)
		sb.WriteString(entry.Level)
		sb.WriteString(f.pairSeparator)
	}

	// Add all key-value pairs
	for k, v := range entry.KeyVals {
		sb.WriteString(k)
		sb.WriteString(f.kvSeparator)
		sb.WriteString(fmt.Sprintf("%v", v))
		sb.WriteString(f.pairSeparator)
	}

	// Remove trailing separator if any
	result := sb.String()
	if len(result) > 0 && strings.HasSuffix(result, f.pairSeparator) {
		result = result[:len(result)-len(f.pairSeparator)]
	}

	return []byte(result), nil
}
