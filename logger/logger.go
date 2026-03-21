package logger

// Logger defines the interface for structured logging
type Logger interface {
	// Info logs an informational message
	Info(message string, keyvals ...interface{}) error
	// Error logs an error message
	Error(message string, keyvals ...interface{}) error
	// Debug logs a debug message
	Debug(message string, keyvals ...interface{}) error
	// Warn logs a warning message
	Warn(message string, keyvals ...interface{}) error
	// WithContext returns a new logger with additional context key-value pairs
	WithContext(keyvals ...interface{}) Logger
	// WithJob returns a new logger with a different job name
	WithJob(job string) Logger
	// Flush ensures all pending logs are sent
	Flush() error
	// Close shuts down the logger and releases resources
	Close() error
}
