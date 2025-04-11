# CloudLog - Structured Logging for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/mwazovzky/cloudlog)](https://goreportcard.com/report/github.com/mwazovzky/cloudlog)
[![GoDoc](https://godoc.org/github.com/mwazovzky/cloudlog?status.svg)](https://godoc.org/github.com/mwazovzky/cloudlog)

CloudLog is a structured logging library for Go applications that provides seamless integration with Grafana Loki and other logging backends.

## Features

- **Structured Logging**: Type-safe key-value pair logging
- **Multiple Log Levels**: Info, Error, Debug, and Warn
- **Context Propagation**: Attach persistent metadata to loggers
- **Flexible Formatting**: JSON, human-readable string, and Loki protocol
- **Grafana Loki Integration**: Native protocol support with proper labeling
- **Synchronous & Asynchronous Modes**: Block or non-block as needed
- **Batched Logging**: High-throughput with configurable batching
- **Error Handling**: Typed errors with helpful checking functions
- **Minimal Dependencies**: Only relies on the standard library and testify for tests

## Installation

```bash
go get github.com/mwazovzky/cloudlog
```

## Quick Start

### Synchronous Logging (Simple Use Cases)

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mwazovzky/cloudlog"
)

func main() {
	// Create HTTP client with timeout
	httpClient := &http.Client{Timeout: 5 * time.Second}

	// Create Loki client
	client := cloudlog.NewClient(
		"http://loki:3100/loki/api/v1/push",
		"username",
		"token",
		httpClient,
	)

	// Create synchronous logger
	logger := cloudlog.NewSync(
		client,
		cloudlog.WithJob("user-service"),
	)

	// Log structured data - blocks until sent
	if err := logger.Info("Service started",
		"version", "1.2.0",
		"env", "production",
	); err != nil {
		fmt.Println("Failed to log:", err)
	}
}
```

### Asynchronous Logging (High Volume)

```go
// Create asynchronous logger for high-throughput scenarios
asyncLogger := cloudlog.NewAsync(
	client,
	cloudlog.WithJob("api-service"),
	cloudlog.WithBufferSize(10000),        // Buffer up to 10,000 log entries
	cloudlog.WithBatchSize(100),           // Send in batches of 100
	cloudlog.WithFlushInterval(1*time.Second), // Flush at least every second
	cloudlog.WithWorkers(4),               // Use 4 worker goroutines
)

// Non-blocking log calls
asyncLogger.Info("Request processed",
	"path", "/api/users",
	"method", "GET",
	"status", 200,
	"duration_ms", 45,
)

// Before application exit, ensure logs are sent
asyncLogger.Flush()  // Wait for all buffered logs to be sent
asyncLogger.Close()  // Release resources
```

## Context Propagation

```go
// Create context-specific logger
userLogger := logger.WithContext(
	"user_id", "user-123",
	"session_id", "abc-xyz",
)

// Each log includes the context automatically
userLogger.Info("User authenticated", "method", "oauth")
userLogger.Warn("Password expiring", "days_left", 5)
```

## Custom Formatting

```go
// String formatter for console output
stringFormatter := cloudlog.NewStringFormatter(
	cloudlog.WithStringTimeFormat("2006-01-02 15:04:05"),
	cloudlog.WithKeyValueSeparator(": "),
	cloudlog.WithPairSeparator(" | "),
)

// Configure logger with formatter
logger := cloudlog.NewSync(
	consoleClient,
	cloudlog.WithJob("api"),
	cloudlog.WithFormatter(stringFormatter),
)
```

## Error Handling

All logging methods return errors that can be checked with helper functions:

```go
err := logger.Info("Operation complete")
if err != nil {
	switch {
	case cloudlog.IsConnectionError(err):
		// Handle connection failure (retry, fallback, etc.)
	case cloudlog.IsFormatError(err):
		// Handle formatting error
	case cloudlog.IsBufferFullError(err):
		// Handle buffer full situation (async logger)
	default:
		// Handle other errors
	}
}
```

## AsyncLogger Configuration

| Option                        | Description                             | Default |
| ----------------------------- | --------------------------------------- | ------- |
| `WithBufferSize(size)`        | Maximum number of log entries in buffer | 1000    |
| `WithBatchSize(size)`         | Number of logs to send in each batch    | 100     |
| `WithFlushInterval(duration)` | Maximum time between flushes            | 5s      |
| `WithWorkers(count)`          | Number of worker goroutines             | 2       |
| `WithBlockOnFull(bool)`       | Whether to block when buffer is full    | false   |

## Documentation

For complete documentation, visit [GoDoc](https://godoc.org/github.com/mwazovzky/cloudlog).

## Examples

For more examples, check the [examples directory](https://github.com/mwazovzky/cloudlog/tree/main/examples).

## Contributing

Contributions are welcome! Please ensure tests pass with `go test ./...` before submitting a pull request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
