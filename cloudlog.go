// Package cloudlog provides a simplified interface for logging to cloud-based logging services.
package cloudlog

import (
	"net/http"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/mwazovzky/cloudlog/logger"
)

// Error type check functions
var (
	IsFormatError     = errors.IsFormatError
	IsConnectionError = errors.IsConnectionError
	IsResponseError   = errors.IsResponseError
)

// Type re-exports
type (
	Logger    = logger.Logger
	LogSender = client.LogSender
	Option    = logger.Option
)

// NewSync creates a new synchronous logger
func NewSync(client client.LogSender, options ...logger.Option) logger.Logger {
	return logger.NewSync(client, options...)
}

// NewClient creates a new Loki client with the given credentials
func NewClient(url, username, token string, httpClient *http.Client) LogSender {
	return client.NewLokiClient(url, username, token, httpClient)
}

// Logger options
func WithFormatter(f formatter.Formatter) Option {
	return logger.WithFormatter(f)
}

func WithJob(job string) Option {
	return logger.WithJob(job)
}

func WithMetadata(key string, value interface{}) Option {
	return logger.WithMetadata(key, value)
}

// Formatter constructors and options
func NewLokiFormatter(options ...formatter.LokiFormatterOption) formatter.Formatter {
	return formatter.NewLokiFormatter(options...)
}

func WithLabelKeys(keys ...string) formatter.LokiFormatterOption {
	return formatter.WithLabelKeys(keys...)
}

func WithTimeFormat(format string) formatter.LokiFormatterOption {
	return formatter.Loki.WithTimeFormat(format)
}
