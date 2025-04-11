package logger

import (
	"fmt"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
)

// SyncLogger implements the Logger interface for synchronous logging
type SyncLogger struct {
	formatter formatter.Formatter
	job       string
	metadata  map[string]interface{}
	sender    client.LogSender
}

// Option defines a configuration option for SyncLogger
type Option func(*SyncLogger)

// NewSync creates a new Logger instance with synchronous delivery
func NewSync(client client.LogSender, options ...Option) Logger {
	logger := &SyncLogger{
		formatter: formatter.NewLokiFormatter(), // Changed from JSONFormatter to LokiFormatter
		job:       "application",
		metadata:  make(map[string]interface{}),
		sender:    client,
	}

	for _, option := range options {
		option(logger) // Apply options directly to SyncLogger
	}

	return logger
}

// Info logs an info level message
func (l *SyncLogger) Info(message string, keyvals ...interface{}) error {
	return l.log("info", message, keyvals...)
}

// Error logs an error level message
func (l *SyncLogger) Error(message string, keyvals ...interface{}) error {
	return l.log("error", message, keyvals...)
}

// Debug logs a debug level message
func (l *SyncLogger) Debug(message string, keyvals ...interface{}) error {
	return l.log("debug", message, keyvals...)
}

// Warn logs a warning level message
func (l *SyncLogger) Warn(message string, keyvals ...interface{}) error {
	return l.log("warn", message, keyvals...)
}

// Close gracefully shuts down the logger
func (l *SyncLogger) Close() error {
	// No special handling needed for sync logger
	return nil
}

// Flush forces delivery of any buffered messages
func (l *SyncLogger) Flush() error {
	// No buffering for sync logger
	return nil
}

// WithContext returns a new logger with additional context
func (l *SyncLogger) WithContext(keyvals ...interface{}) Logger {
	newLogger := &SyncLogger{
		formatter: l.formatter,
		job:       l.job,
		metadata:  copyMetadata(l.metadata),
		sender:    l.sender,
	}

	// Add new context as metadata
	processKeyvals(newLogger.metadata, keyvals...)

	return newLogger
}

// WithJob returns a new logger with a different job name
func (l *SyncLogger) WithJob(job string) Logger {
	newLogger := &SyncLogger{
		formatter: l.formatter,
		job:       job,
		metadata:  copyMetadata(l.metadata),
		sender:    l.sender,
	}

	return newLogger
}

// WithFormatter sets a custom formatter for the logger
func WithFormatter(f formatter.Formatter) Option {
	return func(l *SyncLogger) {
		l.formatter = f
	}
}

// WithJob sets the default job name for the logger
func WithJob(job string) Option {
	return func(l *SyncLogger) {
		l.job = job
	}
}

// WithMetadata adds default metadata to all log entries
func WithMetadata(key string, value interface{}) Option {
	return func(l *SyncLogger) {
		l.metadata[key] = value
	}
}

// log is the internal logging function
func (l *SyncLogger) log(level string, message string, keyvals ...interface{}) error {
	// Combine passed key-values with default metadata
	allKeyVals := make([]interface{}, 0, len(keyvals)+len(l.metadata)*2+2)

	// Add message first
	allKeyVals = append(allKeyVals, "message", message)

	// Add provided key-values
	allKeyVals = append(allKeyVals, keyvals...)

	// Add default metadata
	for k, v := range l.metadata {
		allKeyVals = append(allKeyVals, k, v)
	}

	// Create log entry
	entry := formatter.NewLogEntry(l.job, level, allKeyVals...)

	// Format the log entry
	formatted, err := l.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("%w: failed to format log entry: %v", errors.ErrInvalidFormat, err)
	}

	// Direct synchronous sending
	return l.sender.Send(formatted)
}

// Helper function to copy metadata
func copyMetadata(metadata map[string]interface{}) map[string]interface{} {
	newMetadata := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		newMetadata[k] = v
	}
	return newMetadata
}

// Helper function to process key-value pairs into metadata
func processKeyvals(metadata map[string]interface{}, keyvals ...interface{}) {
	for i := 0; i < len(keyvals)-1; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		metadata[key] = keyvals[i+1]
	}
}

// WithMetadataMap adds multiple metadata key-value pairs to the logger
func WithMetadataMap(md map[string]interface{}) Option {
	return func(l *SyncLogger) {
		for k, v := range md {
			l.metadata[k] = v
		}
	}
}
