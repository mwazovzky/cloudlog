// Package cloudlog provides a structured logging system designed for integration with
// Grafana Loki or other logging backends.
//
// The library consists of several components:
//
// Core components:
//   - Logger: The main logging interface with methods like Info, Error, Debug, and Warn
//   - LogSender: Interface for sending log entries to a backend (e.g., LokiClient)
//   - Formatter: Interface for formatting log entries (JSON, string formats)
//
// Key features:
//   - Structured logging with key-value pairs
//   - Context propagation through WithContext
//   - Multiple formatters (JSON, human-readable)
//   - Custom error types and handling
//   - Comprehensive testing utilities
//   - Native Loki protocol support with proper streams formatting
//
// Basic usage:
//
//	client := cloudlog.NewClient("http://loki-url", "user", "token", &http.Client{})
//	logger := cloudlog.New(client)
//	logger.Info("Request processed", "method", "GET", "path", "/users", "time_ms", 42)
//
// For more examples, see the examples directory.
package cloudlog
