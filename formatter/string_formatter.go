package formatter

import (
	"fmt"
	"strings"
	"time"
)

// StringFormatterOption configures the StringFormatter
type StringFormatterOption func(*StringFormatter)

// String provides a namespace for String formatter options
var String stringOptions

type stringOptions struct{}

// WithTimeFormat sets the time format used for timestamps
func (stringOptions) WithTimeFormat(format string) StringFormatterOption {
	return func(f *StringFormatter) {
		f.timeFormat = format
	}
}

// WithKeyValueSeparator sets the separator between keys and values
func WithKeyValueSeparator(separator string) StringFormatterOption {
	return func(f *StringFormatter) {
		f.keyValueSep = separator
	}
}

// WithPairSeparator sets the separator between key-value pairs
func WithPairSeparator(separator string) StringFormatterOption {
	return func(f *StringFormatter) {
		f.pairSep = separator
	}
}

// StringFormatter formats log entries as readable strings
type StringFormatter struct {
	timeFormat  string
	keyValueSep string
	pairSep     string
}

// NewStringFormatter creates a new StringFormatter
func NewStringFormatter(options ...StringFormatterOption) *StringFormatter {
	formatter := &StringFormatter{
		timeFormat:  time.RFC3339,
		keyValueSep: "=",
		pairSep:     " ",
	}

	for _, option := range options {
		option(formatter)
	}

	return formatter
}

// Format converts a log entry to a human-readable string
func (f *StringFormatter) Format(entry LogEntry) ([]byte, error) {
	var builder strings.Builder

	fmt.Fprintf(&builder, "time%s%s%s", f.keyValueSep, entry.Timestamp.Format(f.timeFormat), f.pairSep)
	fmt.Fprintf(&builder, "job%s%s%s", f.keyValueSep, entry.Job, f.pairSep)
	fmt.Fprintf(&builder, "level%s%s%s", f.keyValueSep, entry.Level, f.pairSep)

	for k, v := range entry.KeyVals {
		fmt.Fprintf(&builder, "%s%s%v%s", k, f.keyValueSep, v, f.pairSep)
	}

	return []byte(builder.String()), nil
}
