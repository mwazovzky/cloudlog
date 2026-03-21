package logger

import "context"

// Logger defines the interface for structured logging
type Logger interface {
	// Info logs an informational message
	Info(ctx context.Context, message string, keyvals ...interface{}) error
	// Error logs an error message
	Error(ctx context.Context, message string, keyvals ...interface{}) error
	// Debug logs a debug message
	Debug(ctx context.Context, message string, keyvals ...interface{}) error
	// Warn logs a warning message
	Warn(ctx context.Context, message string, keyvals ...interface{}) error
	// With returns a new logger with additional metadata key-value pairs
	With(keyvals ...interface{}) Logger
	// WithJob returns a new logger with a different job name
	WithJob(job string) Logger
}
