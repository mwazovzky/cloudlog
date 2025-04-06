package formatter

import (
	"encoding/json"
	"time"

	"github.com/mwazovzky/cloudlog/errors"
)

// JSONFormatterOption defines a function to configure the JSONFormatter
type JSONFormatterOption func(*JSONFormatter)

// WithTimeFormat sets the time format for the JSONFormatter
func WithTimeFormat(format string) JSONFormatterOption {
	return func(f *JSONFormatter) {
		f.timeFormat = format
	}
}

// WithTimestampField sets the name of the timestamp field
func WithTimestampField(field string) JSONFormatterOption {
	return func(f *JSONFormatter) {
		f.timestampField = field
	}
}

// WithLevelField sets the name of the level field
func WithLevelField(field string) JSONFormatterOption {
	return func(f *JSONFormatter) {
		f.levelField = field
	}
}

// WithJobField sets the name of the job field
func WithJobField(field string) JSONFormatterOption {
	return func(f *JSONFormatter) {
		f.jobField = field
	}
}

// JSONFormatter formats log entries as JSON
type JSONFormatter struct {
	timeFormat     string
	timestampField string
	levelField     string
	jobField       string
}

// NewJSONFormatter creates a new JSONFormatter with the given options
func NewJSONFormatter(options ...JSONFormatterOption) *JSONFormatter {
	formatter := &JSONFormatter{
		timeFormat:     time.RFC3339,
		timestampField: "timestamp",
		levelField:     "level",
		jobField:       "job",
	}

	for _, option := range options {
		option(formatter)
	}

	return formatter
}

// Format converts a log entry into a JSON string
func (f *JSONFormatter) Format(entry LogEntry) ([]byte, error) {
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
	formatted, err := json.Marshal(data)
	if err != nil {
		return nil, errors.FormatError(err, "failed to marshal JSON")
	}

	return formatted, nil
}
