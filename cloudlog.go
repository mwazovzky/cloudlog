// Package cloudlog provides a structured logging system designed for
// integration with Grafana Loki. It offers a simple façade over the
// functionality in subpackages, making it easy to get started while
// still allowing advanced usage through direct subpackage imports.

package cloudlog

import (
	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/delivery"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/mwazovzky/cloudlog/logger"
)

// Export error types for convenience
var (
	ErrInvalidFormat    = errors.ErrInvalidFormat
	ErrConnectionFailed = errors.ErrConnectionFailed
	ErrResponseError    = errors.ErrResponseError
	ErrInvalidInput     = errors.ErrInvalidInput
	ErrBufferFull       = errors.ErrBufferFull
	ErrTimeout          = errors.ErrTimeout
	ErrShutdown         = errors.ErrShutdown
)

// IsConnectionError returns true if the error is related to connection failures
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}

// IsResponseError returns true if the error is related to an error response
func IsResponseError(err error) bool {
	return errors.Is(err, ErrResponseError)
}

// IsFormatError returns true if the error is related to formatting issues
func IsFormatError(err error) bool {
	return errors.Is(err, ErrInvalidFormat)
}

// IsInputError returns true if the error is related to invalid inputs
func IsInputError(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsBufferFullError returns true if error is related to buffer full condition
func IsBufferFullError(err error) bool {
	return errors.Is(err, ErrBufferFull)
}

// IsTimeoutError returns true if error is related to timeout
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsShutdownError returns true if error is related to shutdown problems
func IsShutdownError(err error) bool {
	return errors.Is(err, ErrShutdown)
}

// Logger is an alias for logger.Logger to maintain backwards compatibility
type Logger = logger.Logger

// DeliveryConfig is an alias for delivery.Config to make it available from the root package
type DeliveryConfig = delivery.Config

// DeliveryStatus is an alias for delivery.DeliveryStatus
type DeliveryStatus = delivery.DeliveryStatus

// LogDeliverer is an alias for delivery.LogDeliverer
type LogDeliverer = delivery.LogDeliverer

// NewClient creates a new Loki client
func NewClient(url, username, token string, httpClient client.Doer) client.LogSender {
	return client.NewLokiClient(url, username, token, httpClient)
}

// NewClientWithOptions creates a client with additional options
func NewClientWithOptions(url, username, token string, httpClient client.Doer,
	options ...client.LokiClientOption) client.LogSender {
	return client.NewLokiClientWithOptions(url, username, token, httpClient, options...)
}

// DefaultDeliveryConfig returns the default delivery configuration
func DefaultDeliveryConfig() DeliveryConfig {
	return delivery.DefaultConfig()
}

// New creates a new logger with the given client and synchronous delivery
func New(c client.LogSender, options ...logger.Option) *logger.Logger {
	return logger.New(c, options...)
}

// NewWithDeliverer creates a logger with a custom deliverer
func NewWithDeliverer(deliverer delivery.LogDeliverer, options ...logger.Option) *logger.Logger {
	return logger.NewWithDeliverer(deliverer, options...)
}

// NewAsync creates a new logger with asynchronous delivery using default config
func NewAsync(c client.LogSender, options ...logger.Option) *logger.Logger {
	config := delivery.DefaultConfig()
	config.Async = true
	return logger.NewAsync(c, config, options...)
}

// NewAsyncWithConfig creates a new logger with custom async configuration
func NewAsyncWithConfig(c client.LogSender, config delivery.Config, options ...logger.Option) *logger.Logger {
	return logger.NewAsync(c, config, options...)
}

// NewBatchLogger creates a new logger with batch processing
func NewBatchLogger(c client.LogSender, options ...logger.Option) *logger.Logger {
	config := delivery.DefaultConfig()
	config.Async = true

	batchDeliverer := delivery.NewBatchDeliverer(c, config)
	return logger.NewWithDeliverer(batchDeliverer, options...)
}

// NewBatchLoggerWithConfig creates a new logger with custom batch configuration
func NewBatchLoggerWithConfig(c client.LogSender, config delivery.Config, options ...logger.Option) *logger.Logger {
	batchDeliverer := delivery.NewBatchDeliverer(c, config)
	return logger.NewWithDeliverer(batchDeliverer, options...)
}

// NewSyncDeliverer creates a synchronous deliverer
func NewSyncDeliverer(sender client.LogSender) delivery.LogDeliverer {
	return delivery.NewSyncDeliverer(sender)
}

// NewAsyncDeliverer creates an asynchronous deliverer
func NewAsyncDeliverer(sender client.LogSender, config delivery.Config) delivery.LogDeliverer {
	return delivery.NewAsyncDeliverer(sender, config)
}

// NewBatchDeliverer creates a batch deliverer
func NewBatchDeliverer(sender client.LogSender, config delivery.Config) delivery.LogDeliverer {
	return delivery.NewBatchDeliverer(sender, config)
}

// WithFormatter sets the formatter for the logger
func WithFormatter(f formatter.Formatter) logger.Option {
	return logger.WithFormatter(f)
}

// WithJob sets the job name for the logger
func WithJob(job string) logger.Option {
	return logger.WithJob(job)
}

// WithMetadata adds default metadata to all log entries
func WithMetadata(key string, value interface{}) logger.Option {
	return logger.WithMetadata(key, value)
}

// IsError checks if an error is present
func IsError(err error) bool {
	return err != nil
}
