// Package cloudlog provides a simplified interface for logging to cloud-based logging services.
package cloudlog

import (
	"net/http"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/mwazovzky/cloudlog/logger"
)

// Error type check functions - re-exported from errors package for convenience
var (
	IsFormatError     = errors.IsFormatError
	IsConnectionError = errors.IsConnectionError
	IsResponseError   = errors.IsResponseError
	IsInputError      = errors.IsInputError
	IsBufferFullError = errors.IsBufferFullError
	IsTimeoutError    = errors.IsTimeoutError
	IsShutdownError   = errors.IsShutdownError
)

// Interface re-exports for public API
type (
	Logger       = logger.Logger
	LogProducer  = logger.LogProducer
	LogManager   = logger.LogManager
	ContextAware = logger.ContextAware
	LogSender    = client.LogSender
)

// Option re-exports
type (
	Option      = logger.Option
	AsyncOption = logger.AsyncOption
)

// NewSync creates a new synchronous logger that implements the Logger interface
func NewSync(client client.LogSender, options ...logger.Option) logger.Logger {
	return logger.NewSync(client, options...)
}

// NewAsync creates an asynchronous logger with non-blocking behavior
func NewAsync(client client.LogSender, options ...logger.AsyncOption) logger.Logger {
	return logger.NewAsync(client, options...)
}

// WithFormatter, WithJob, WithMetadata pass through to logger package
func WithFormatter(f formatter.Formatter) Option {
	return logger.WithFormatter(f)
}

func WithJob(job string) Option {
	return logger.WithJob(job)
}

func WithMetadata(key string, value interface{}) Option {
	return logger.WithMetadata(key, value)
}

// NewClient creates a new Loki client with the given credentials
func NewClient(url, username, token string, httpClient *http.Client) LogSender {
	return client.NewLokiClient(url, username, token, httpClient)
}

// NewClientWithOptions creates a client with the provided options
func NewClientWithOptions(url, username, token string, httpClient *http.Client, opts ...client.LokiClientOption) LogSender {
	return client.NewLokiClientWithOptions(url, username, token, httpClient, opts...)
}

// NewLokiFormatter creates a formatter specifically designed for Loki output
// This formatter handles both the Loki protocol structure and log content formatting
func NewLokiFormatter(options ...formatter.LokiFormatterOption) formatter.Formatter {
	return formatter.NewLokiFormatter(options...)
}

// WithLabelKeys specifies keys from the log entry that should be added as Loki stream labels
func WithLabelKeys(keys ...string) formatter.LokiFormatterOption {
	return formatter.WithLabelKeys(keys...)
}

// WithTimeFormat sets the time format for timestamp fields in LokiFormatter
func WithTimeFormat(format string) formatter.LokiFormatterOption {
	return formatter.Loki.WithTimeFormat(format)
}

// WithTimestampField sets the field name for timestamps in LokiFormatter
func WithTimestampField(field string) formatter.LokiFormatterOption {
	return formatter.Loki.WithTimestampField(field)
}

// WithLevelField sets the field name for log level in LokiFormatter
func WithLevelField(field string) formatter.LokiFormatterOption {
	return formatter.Loki.WithLevelField(field)
}

// WithJobField sets the field name for job in LokiFormatter
func WithJobField(field string) formatter.LokiFormatterOption {
	return formatter.Loki.WithJobField(field)
}

// WithStringTimeFormat sets the time format for StringFormatter
func WithStringTimeFormat(format string) formatter.StringFormatterOption {
	return formatter.String.WithTimeFormat(format)
}

// WithKeyValueSeparator sets the separator between keys and values in StringFormatter
func WithKeyValueSeparator(separator string) formatter.StringFormatterOption {
	return formatter.WithKeyValueSeparator(separator)
}

// WithPairSeparator sets the separator between key-value pairs in StringFormatter
func WithPairSeparator(separator string) formatter.StringFormatterOption {
	return formatter.WithPairSeparator(separator)
}

// Async configuration options
func WithBufferSize(size int) AsyncOption {
	return logger.WithBufferSize(size)
}

func WithBatchSize(size int) AsyncOption {
	return logger.WithBatchSize(size)
}

func WithFlushInterval(interval time.Duration) AsyncOption {
	return logger.WithFlushInterval(interval)
}

func WithWorkers(count int) AsyncOption {
	return logger.WithWorkers(count)
}

func WithBlockOnFull(block bool) AsyncOption {
	return logger.WithBlockOnFull(block)
}

// Re-export async logger options
func WithAsyncFormatter(f formatter.Formatter) AsyncOption {
	return logger.WithAsyncFormatter(f)
}

func WithAsyncJob(job string) AsyncOption {
	return logger.WithAsyncJob(job)
}

func WithAsyncMetadata(key string, value interface{}) AsyncOption {
	return logger.WithAsyncMetadata(key, value)
}

// WithErrorHandler sets a custom handler for internal errors in AsyncLogger
func WithErrorHandler(handler errors.ErrorHandler) AsyncOption {
	return logger.WithErrorHandler(handler)
}
