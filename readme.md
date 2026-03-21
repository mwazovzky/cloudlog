![Tests](https://github.com/mwazovzky/cloudlog/actions/workflows/test.yml/badge.svg)

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
- **Error Handling**: Typed errors with helpful checking functions
- **Minimal Dependencies**: Only relies on the standard library and testify for tests

## Installation

```bash
go get github.com/mwazovzky/cloudlog
```

## Quick Start

### Basic Usage

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
	default:
		// Handle other errors
	}
}
```

## Logger Configuration Options

### StringFormatter Options

| Option                         | Description                                |
| ------------------------------ | ------------------------------------------ |
| `WithStringTimeFormat(format)` | Sets the time format for string formatter  |
| `WithKeyValueSeparator(sep)`   | Sets the separator between keys and values |
| `WithPairSeparator(sep)`       | Sets the separator between key-value pairs |

### LokiFormatter Options

| Option                      | Description                                          |
| --------------------------- | ---------------------------------------------------- |
| `WithLabelKeys(keys...)`    | Specifies keys to use as labels in Loki formatter    |
| `WithTimeFormat(format)`    | Sets the time format for Loki formatter              |
| `WithTimestampField(field)` | Sets the field name for timestamps in Loki formatter |
| `WithLevelField(field)`     | Sets the field name for log levels in Loki formatter |
| `WithJobField(field)`       | Sets the field name for job in Loki formatter        |

### SyncLogger Configuration

| Option                     | Description                              |
| -------------------------- | ---------------------------------------- |
| `WithJob(job)`             | Sets the default job name for the logger |
| `WithMetadata(key, value)` | Adds default metadata to all log entries |
| `WithFormatter(formatter)` | Sets a custom formatter for the logger   |

## Documentation

For complete documentation, visit [GoDoc](https://godoc.org/github.com/mwazovzky/cloudlog).

## Examples

For more examples, check the [examples directory](https://github.com/mwazovzky/cloudlog/tree/main/examples).

## License

This project is licensed under the MIT License - see the LICENSE file for details.
