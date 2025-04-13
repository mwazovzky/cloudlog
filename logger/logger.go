package logger

// LogProducer defines methods for producing logs of different severity levels
type LogProducer interface {
	// Info logs an informational message
	Info(message string, keyvals ...interface{}) error

	// Error logs an error message
	Error(message string, keyvals ...interface{}) error

	// Debug logs a debug message
	Debug(message string, keyvals ...interface{}) error

	// Warn logs a warning message
	Warn(message string, keyvals ...interface{}) error
}

// LogManager defines methods for managing logger lifecycle
type LogManager interface {
	// Flush ensures all pending logs are sent
	Flush() error

	// Close shuts down the logger and releases resources
	Close() error
}

// ContextAware defines methods for creating loggers with context
type ContextAware interface {
	// WithContext returns a new logger with additional context key-value pairs
	WithContext(keyvals ...interface{}) Logger

	// WithJob returns a new logger with a different job name
	WithJob(job string) Logger
}

// Logger is the composite interface combining all logging capabilities
type Logger interface {
	LogProducer
	LogManager
	ContextAware
}
