// Package logger provides the core logging functionality.
package logger

import (
	"fmt"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
)

// Logger is the main logger implementation
type Logger struct {
	client    client.LogSender
	formatter formatter.Formatter
	job       string
	metadata  map[string]interface{}
}

// Option defines a function to configure Logger
type Option func(*Logger)

// WithFormatter sets the formatter for the logger
func WithFormatter(f formatter.Formatter) Option {
	return func(l *Logger) {
		l.formatter = f
	}
}

// WithJob sets the default job name
func WithJob(job string) Option {
	return func(l *Logger) {
		l.job = job
	}
}

// WithMetadata adds default metadata to all log entries
func WithMetadata(key string, value interface{}) Option {
	return func(l *Logger) {
		l.metadata[key] = value
	}
}

// New creates a new Logger instance
func New(client client.LogSender, options ...Option) *Logger {
	logger := &Logger{
		client:    client,
		formatter: formatter.NewJSONFormatter(),
		job:       "application",
		metadata:  make(map[string]interface{}),
	}

	for _, option := range options {
		option(logger)
	}

	return logger
}

// Info logs an info level message
func (l *Logger) Info(message string, keyvals ...interface{}) error {
	return l.log("info", message, keyvals...)
}

// Error logs an error level message
func (l *Logger) Error(message string, keyvals ...interface{}) error {
	return l.log("error", message, keyvals...)
}

// Debug logs a debug level message
func (l *Logger) Debug(message string, keyvals ...interface{}) error {
	return l.log("debug", message, keyvals...)
}

// Warn logs a warning level message
func (l *Logger) Warn(message string, keyvals ...interface{}) error {
	return l.log("warn", message, keyvals...)
}

// WithContext returns a new logger with additional context
func (l *Logger) WithContext(keyvals ...interface{}) *Logger {
	newLogger := &Logger{
		client:    l.client,
		formatter: l.formatter,
		job:       l.job,
		metadata:  make(map[string]interface{}),
	}

	// Copy existing metadata
	for k, v := range l.metadata {
		newLogger.metadata[k] = v
	}

	// Add new context as metadata
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			if key, ok := keyvals[i].(string); ok {
				newLogger.metadata[key] = keyvals[i+1]
			}
		}
	}

	return newLogger
}

// WithJob returns a new logger with a different job name
func (l *Logger) WithJob(job string) *Logger {
	newLogger := &Logger{
		client:    l.client,
		formatter: l.formatter,
		job:       job,
		metadata:  make(map[string]interface{}),
	}

	// Copy existing metadata
	for k, v := range l.metadata {
		newLogger.metadata[k] = v
	}

	return newLogger
}

// log is the internal logging function
func (l *Logger) log(level string, message string, keyvals ...interface{}) error {
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

	// Format and send the log entry
	formatted, err := l.formatter.Format(entry)
	if err != nil {
		return errors.FormatError(err, "failed to format log entry")
	}

	if err := l.client.Send(l.job, formatted); err != nil {
		return fmt.Errorf("failed to send log entry: %w", err)
	}

	return nil
}
