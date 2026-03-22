package logger

import (
	"context"
	"fmt"

	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
)

// logger implements the Logger interface
type logger struct {
	formatter formatter.Formatter
	job       string
	metadata  map[string]interface{}
	sender    Sender
	labelKeys []string
	minLevel  int
}

// Log level constants
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelValues = map[string]int{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

// Option defines a configuration option for the logger
type Option func(*logger)

// New creates a new Logger with the given sender and options
func New(sender Sender, options ...Option) Logger {
	l := &logger{
		formatter: formatter.NewLokiFormatter(),
		job:       "application",
		metadata:  make(map[string]interface{}),
		sender:    sender,
		minLevel:  LevelDebug,
	}

	for _, option := range options {
		option(l)
	}

	return l
}

func (l *logger) Info(ctx context.Context, message string, keyvals ...interface{}) error {
	return l.log(ctx, "info", message, keyvals...)
}

func (l *logger) Error(ctx context.Context, message string, keyvals ...interface{}) error {
	return l.log(ctx, "error", message, keyvals...)
}

func (l *logger) Debug(ctx context.Context, message string, keyvals ...interface{}) error {
	return l.log(ctx, "debug", message, keyvals...)
}

func (l *logger) Warn(ctx context.Context, message string, keyvals ...interface{}) error {
	return l.log(ctx, "warn", message, keyvals...)
}

func (l *logger) With(keyvals ...interface{}) Logger {
	newLogger := &logger{
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

func (l *logger) WithJob(job string) Logger {
	return &logger{
		formatter: l.formatter,
		job:       job,
		metadata:  copyMetadata(l.metadata),
		sender:    l.sender,
		labelKeys: l.labelKeys,
		minLevel:  l.minLevel,
	}
}

// Option constructors

func WithFormatter(f formatter.Formatter) Option {
	return func(l *logger) {
		l.formatter = f
	}
}

func WithJob(job string) Option {
	return func(l *logger) {
		l.job = job
	}
}

func WithMetadata(key string, value interface{}) Option {
	return func(l *logger) {
		l.metadata[key] = value
	}
}

func WithLabelKeys(keys ...string) Option {
	return func(l *logger) {
		l.labelKeys = append([]string{}, keys...)
	}
}

func WithMinLevel(level int) Option {
	return func(l *logger) {
		l.minLevel = level
	}
}

// log is the internal logging function
func (l *logger) log(ctx context.Context, level string, message string, keyvals ...interface{}) error {
	if levelVal, ok := levelValues[level]; ok && levelVal < l.minLevel {
		return nil
	}

	// Combine message, provided key-values, and default metadata
	allKeyVals := make([]interface{}, 0, len(keyvals)+len(l.metadata)*2+2)
	allKeyVals = append(allKeyVals, "message", message)
	allKeyVals = append(allKeyVals, keyvals...)
	for k, v := range l.metadata {
		allKeyVals = append(allKeyVals, k, v)
	}

	// Create log entry
	entry := formatter.NewLogEntry(l.job, level, allKeyVals...)

	// Extract label values and remove from content
	labels := map[string]string{
		"job": l.job,
	}
	for _, key := range l.labelKeys {
		if value, exists := entry.KeyVals[key]; exists {
			labels[key] = fmt.Sprintf("%v", value)
			delete(entry.KeyVals, key)
		}
	}

	// Format content
	content, err := l.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("%w: failed to format log entry: %v", errors.ErrInvalidFormat, err)
	}

	return l.sender.Send(ctx, content, labels, entry.Timestamp)
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
