package logger

import (
	"context"
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
	labelKeys []string // keys to promote to Loki stream labels
	minLevel  int      // minimum log level to send
}

// Log level constants
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

// levelValues maps level strings to their numeric values
var levelValues = map[string]int{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

// Option defines a configuration option for SyncLogger
type Option func(*SyncLogger)

// NewSync creates a new Logger instance with synchronous delivery
func NewSync(client client.LogSender, options ...Option) Logger {
	logger := &SyncLogger{
		formatter: formatter.NewLokiFormatter(),
		job:       "application",
		metadata:  make(map[string]interface{}),
		sender:    client,
		minLevel:  LevelDebug,
	}

	for _, option := range options {
		option(logger)
	}

	return logger
}

// Info logs an info level message
func (l *SyncLogger) Info(ctx context.Context, message string, keyvals ...interface{}) error {
	return l.log(ctx, "info", message, keyvals...)
}

// Error logs an error level message
func (l *SyncLogger) Error(ctx context.Context, message string, keyvals ...interface{}) error {
	return l.log(ctx, "error", message, keyvals...)
}

// Debug logs a debug level message
func (l *SyncLogger) Debug(ctx context.Context, message string, keyvals ...interface{}) error {
	return l.log(ctx, "debug", message, keyvals...)
}

// Warn logs a warning level message
func (l *SyncLogger) Warn(ctx context.Context, message string, keyvals ...interface{}) error {
	return l.log(ctx, "warn", message, keyvals...)
}

// With returns a new logger with additional metadata
func (l *SyncLogger) With(keyvals ...interface{}) Logger {
	newLogger := &SyncLogger{
		formatter: l.formatter,
		job:       l.job,
		metadata:  copyMetadata(l.metadata),
		sender:    l.sender,
		labelKeys: l.labelKeys,
		minLevel:  l.minLevel,
	}

	processKeyvals(newLogger.metadata, keyvals...)
	return newLogger
}

// WithJob returns a new logger with a different job name
func (l *SyncLogger) WithJob(job string) Logger {
	return &SyncLogger{
		formatter: l.formatter,
		job:       job,
		metadata:  copyMetadata(l.metadata),
		sender:    l.sender,
		labelKeys: l.labelKeys,
		minLevel:  l.minLevel,
	}
}

// Option constructors

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

// WithLabelKeys specifies keys to promote to Loki stream labels
func WithLabelKeys(keys ...string) Option {
	return func(l *SyncLogger) {
		l.labelKeys = keys
	}
}

// WithMinLevel sets the minimum log level (LevelDebug, LevelInfo, LevelWarn, LevelError)
func WithMinLevel(level int) Option {
	return func(l *SyncLogger) {
		l.minLevel = level
	}
}

// log is the internal logging function
func (l *SyncLogger) log(ctx context.Context, level string, message string, keyvals ...interface{}) error {
	// Check minimum level
	if levelVal, ok := levelValues[level]; ok && levelVal < l.minLevel {
		return nil
	}

	// Combine passed key-values with default metadata
	allKeyVals := make([]interface{}, 0, len(keyvals)+len(l.metadata)*2+2)
	allKeyVals = append(allKeyVals, "message", message)
	allKeyVals = append(allKeyVals, keyvals...)
	for k, v := range l.metadata {
		allKeyVals = append(allKeyVals, k, v)
	}

	// Create log entry
	entry := formatter.NewLogEntry(l.job, level, allKeyVals...)

	// Extract label values and remove them from content
	labels := map[string]string{
		"job": l.job,
	}
	for _, key := range l.labelKeys {
		if value, exists := entry.KeyVals[key]; exists {
			labels[key] = fmt.Sprintf("%v", value)
			delete(entry.KeyVals, key)
		}
	}

	// Format the log content
	content, err := l.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("%w: failed to format log entry: %v", errors.ErrInvalidFormat, err)
	}

	// Build LokiEntry and send
	lokiEntry := client.LokiEntry{
		Streams: []client.LokiStream{
			{
				Stream: labels,
				Values: [][]string{
					{
						fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
						string(content),
					},
				},
			},
		},
	}

	return l.sender.Send(ctx, lokiEntry)
}

func copyMetadata(metadata map[string]interface{}) map[string]interface{} {
	newMetadata := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		newMetadata[k] = v
	}
	return newMetadata
}

func processKeyvals(metadata map[string]interface{}, keyvals ...interface{}) {
	for i := 0; i < len(keyvals)-1; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		metadata[key] = keyvals[i+1]
	}
}
