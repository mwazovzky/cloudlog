package formatter

import (
	"fmt"
	"strings"
	"time"

	"github.com/mwazovzky/cloudlog/client"
)

// StringFormatterOption configures the StringFormatter
type StringFormatterOption func(*StringFormatter)

// String provides a namespace for String formatter options
var String stringOptions

type stringOptions struct{}

// WithTimeFormat sets the time format used for timestamps
func (stringOptions) WithTimeFormat(format string) StringFormatterOption {
	return func(f *StringFormatter) {
		f.SetTimeFormat(format)
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
	keyValueSep string // Field name needs to match what tests expect
	pairSep     string // Field name needs to match what tests expect
}

// SetTimeFormat sets the time format for the formatter
func (f *StringFormatter) SetTimeFormat(format string) {
	f.timeFormat = format
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

// Format converts a log entry to a formatted string
// To maintain compatibility with the Formatter interface, it returns a LokiEntry
func (f *StringFormatter) Format(entry LogEntry) (client.LokiEntry, error) {
	var builder strings.Builder

	// Add timestamp
	builder.WriteString(fmt.Sprintf("time%s%s%s", f.keyValueSep, entry.Timestamp.Format(f.timeFormat), f.pairSep))

	// Add job and level
	builder.WriteString(fmt.Sprintf("job%s%s%s", f.keyValueSep, entry.Job, f.pairSep))
	builder.WriteString(fmt.Sprintf("level%s%s%s", f.keyValueSep, entry.Level, f.pairSep))

	// Add all other key-values
	for k, v := range entry.KeyVals {
		builder.WriteString(fmt.Sprintf("%s%s%v%s", k, f.keyValueSep, v, f.pairSep))
	}

	// Create a LokiEntry from the formatted string
	content := builder.String()
	lokiEntry := client.LokiEntry{
		Streams: []client.LokiStream{
			{
				Stream: map[string]string{
					"job": entry.Job,
				},
				Values: [][]string{
					{
						fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
						content,
					},
				},
			},
		},
	}

	return lokiEntry, nil
}

// FormatBatch converts multiple log entries for the same job into a single Loki entry
func (f *StringFormatter) FormatBatch(job string, entries []LogEntry) (client.LokiEntry, error) {
	if len(entries) == 0 {
		// Return an empty entry if there are no entries
		return client.LokiEntry{
			Streams: []client.LokiStream{},
		}, nil
	}

	// Create a single stream for this job
	stream := client.LokiStream{
		Stream: map[string]string{
			"job": job,
		},
		Values: make([][]string, 0, len(entries)),
	}

	// Format each entry and add it to the stream
	for _, entry := range entries {
		var builder strings.Builder

		// Add timestamp
		builder.WriteString(fmt.Sprintf("time%s%s%s", f.keyValueSep, entry.Timestamp.Format(f.timeFormat), f.pairSep))

		// Add job and level
		builder.WriteString(fmt.Sprintf("job%s%s%s", f.keyValueSep, entry.Job, f.pairSep))
		builder.WriteString(fmt.Sprintf("level%s%s%s", f.keyValueSep, entry.Level, f.pairSep))

		// Add all other key-values
		for k, v := range entry.KeyVals {
			builder.WriteString(fmt.Sprintf("%s%s%v%s", k, f.keyValueSep, v, f.pairSep))
		}

		// Add this formatted entry to the stream's values
		stream.Values = append(stream.Values, []string{
			fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
			builder.String(),
		})
	}

	// Return the combined Loki entry
	return client.LokiEntry{
		Streams: []client.LokiStream{stream},
	}, nil
}
