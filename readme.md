# CloudLog - Structured Logging for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/mwazovzky/cloudlog)](https://goreportcard.com/report/github.com/mwazovzky/cloudlog)
[![GoDoc](https://godoc.org/github.com/mwazovzky/cloudlog?status.svg)](https://godoc.org/github.com/mwazovzky/cloudlog)

CloudLog is a structured logging library designed for sending logs to Grafana Loki and other logging backends. It provides a clean, simple interface for logging structured data with key-value pairs and customizable formatting.

## Features

- **Structured Logging**: Key-value pair logging with support for all data types
- **Multiple Log Levels**: Info, Error, Debug, and Warn levels
- **Context Propagation**: Add context to loggers that gets included in all messages
- **Custom Formatters**: JSON and human-readable string formats built-in
- **Loki Integration**: Native support for Grafana Loki with proper streams protocol
- **Error Handling**: Type-based error handling with useful categorization
- **Testing Utilities**: Comprehensive tools for testing logging behavior

## Installation

```bash
go get github.com/mwazovzky/cloudlog
```

## Quick Start

```go
package main

import (
	"github.com/mwazovzky/cloudlog"
	"net/http"
	"time"
)

func main() {
	// Create Loki client
	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := cloudlog.NewClient(
		"http://loki-instance/api/v1/push",
		"username",
		"token",
		httpClient,
	)

	// Create logger
	logger := cloudlog.New(client)

	// Basic logging
	logger.Info("Application started", "version", "1.0.0")

	// With context
	reqLogger := logger.WithContext("request_id", "abc-123")
	reqLogger.Info("Request received", "method", "GET", "path", "/users")

	// Different log levels
	reqLogger.Debug("Processing request", "params", "limit=100")
	reqLogger.Warn("High latency detected", "latency_ms", 250)
	reqLogger.Error("Request failed", "status", 500, "error", "database connection lost")
}
```

## Package Structure

- **cloudlog** (root): Convenience functions and type aliases
- **client**: Log-sending implementations (LokiClient)
- **logger**: Core logging functionality
- **formatter**: Log formatting options (JSON, string)
- **errors**: Error types and handling utilities
- **testing**: Testing utilities

## Advanced Usage

### Custom Formatters

```go
// JSON formatter with custom options
jsonFormatter := formatter.NewJSONFormatter(
    formatter.WithTimeFormat(time.RFC1123),
    formatter.WithTimestampField("@timestamp"),
)

// Human-readable formatter
stringFormatter := formatter.NewStringFormatter(
    formatter.WithKeyValueSeparator(": "),
    formatter.WithPairSeparator(" | "),
)

logger := cloudlog.New(client, cloudlog.WithFormatter(stringFormatter))
```

### Context and Job Names

```go
// Add context to logs
userLogger := logger.WithContext(
    "user_id", "user-123",
    "session_id", "sess-456",
)

// Change job/source name
apiLogger := logger.WithJob("api-service")
```

### Error Handling

```go
err := logger.Info("Test message")
if err != nil {
    if cloudlog.IsConnectionError(err) {
        // Handle connection error
    } else if cloudlog.IsFormatError(err) {
        // Handle formatting error
    }
}
```

### Testing

```go
func TestFunction(t *testing.T) {
    // Create test logger
    testLogger := testing.NewTestLogger()
    logger := cloudlog.New(testLogger)

    // Run function that logs
    functionUnderTest(logger)

    // Verify logs
    if !testLogger.ContainsMessage("Operation completed") {
        t.Error("Expected log message not found")
    }

    errorLogs := testLogger.LogsOfLevel("error")
    if len(errorLogs) > 0 {
        t.Error("Function logged unexpected errors")
    }
}
```

### Asynchronous and Batch Logging

```go
// Asynchronous logging
asyncLogger := cloudlog.NewAsync(client, cloudlog.WithJob("async-logger"))
asyncLogger.Info("Async log message")

// Batch logging
batchConfig := cloudlog.DefaultDeliveryConfig()
batchConfig.BatchSize = 3
batchLogger := cloudlog.NewBatchLoggerWithConfig(client, batchConfig)
batchLogger.Info("Batch log message")

// Ensure logs are flushed before shutdown
asyncLogger.Flush()
batchLogger.Flush()
```

## Examples

For complete examples, see the [examples directory](./examples).

## Configuration

CloudLog can be configured through environment variables or code:

```go
// With options
logger := cloudlog.New(client,
    cloudlog.WithJob("my-service"),
    cloudlog.WithMetadata("environment", "production"),
)
```

## Loki Protocol Support

CloudLog's Loki client implements the Loki push API protocol properly:

```go
// The client automatically formats logs into the required Loki format:
{
  "streams": [
    {
      "stream": {
        "job": "your-service-name"
      },
      "values": [
        ["1626882892000000000", "{\"level\":\"info\",\"message\":\"your log message\",\"key\":\"value\"}"]
      ]
    }
  ]
}
```

This ensures logs are correctly ingested by Grafana Loki.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
