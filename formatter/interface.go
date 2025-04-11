// Package formatter provides different formatting options for log entries.
// It supports both Loki protocol for cloud logging and human-readable string
// formats for console output, with customizable options.
package formatter

// This file has been merged with formatter.go
// Formatter interface is now defined in formatter.go

// FormatterConfig defines common configuration capabilities for formatters
type FormatterConfig interface {
	// SetTimeFormat sets the time format to be used when formatting timestamps
	SetTimeFormat(format string)
}

// Each formatter type defines its own option type for type safety:
// - LokiFormatterOption func(*LokiFormatter)
// - StringFormatterOption func(*StringFormatter)

// Common option patterns are implemented specifically for each formatter:
// WithTimeFormat - Sets time format (available for all formatters)
// WithTimestampField, WithLevelField, WithJobField - Field name options for LokiFormatter
// WithKeyValueSeparator, WithPairSeparator - Format options for StringFormatter
