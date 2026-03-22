![Tests](https://github.com/mwazovzky/cloudlog/actions/workflows/test.yml/badge.svg)

# CloudLog - Structured Logging for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/mwazovzky/cloudlog)](https://goreportcard.com/report/github.com/mwazovzky/cloudlog)
[![GoDoc](https://godoc.org/github.com/mwazovzky/cloudlog?status.svg)](https://godoc.org/github.com/mwazovzky/cloudlog)

CloudLog is a structured logging library for Go applications with Grafana Loki integration.

## Features

- **Structured Logging**: Key-value pair logging with `context.Context` support
- **Multiple Log Levels**: Info, Error, Debug, Warn with level filtering
- **Metadata Propagation**: Attach persistent metadata to loggers
- **Grafana Loki Integration**: Native protocol support with label promotion
- **Error Handling**: Typed sentinel errors with helper functions
- **Minimal Dependencies**: Standard library + testify for tests

## Installation

```bash
go get github.com/mwazovzky/cloudlog
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mwazovzky/cloudlog"
)

func main() {
	ctx := context.Background()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := cloudlog.NewClient(
		"http://loki:3100/loki/api/v1/push",
		"username", "token", httpClient,
	)

	sender := cloudlog.NewSyncSender(client)
	logger := cloudlog.New(sender, cloudlog.WithJob("user-service"))

	if err := logger.Info(ctx, "Service started",
		"version", "1.2.0",
		"env", "production",
	); err != nil {
		fmt.Println("Failed to log:", err)
	}
}
```

## Metadata

```go
// Add persistent metadata
userLogger := logger.With("user_id", "user-123", "session_id", "abc-xyz")

userLogger.Info(ctx, "User authenticated", "method", "oauth")
userLogger.Warn(ctx, "Password expiring", "days_left", 5)
```

## Loki Labels

Promote keys to Loki stream labels (removes them from log content):

```go
logger := cloudlog.New(sender,
	cloudlog.WithJob("api-service"),
	cloudlog.WithLabelKeys("request_id", "user_id"),
)
```

## Level Filtering

```go
logger := cloudlog.New(sender,
	cloudlog.WithMinLevel(cloudlog.LevelWarn), // only Warn and Error
)
```

## Error Handling

```go
err := logger.Info(ctx, "Operation complete")
if err != nil {
	switch {
	case cloudlog.IsConnectionError(err):
		// Handle connection failure
	case cloudlog.IsFormatError(err):
		// Handle formatting error
	}
}
```

## Configuration Options

| Option                     | Description                                    |
| -------------------------- | ---------------------------------------------- |
| `WithJob(job)`             | Sets the job name (Loki stream label)          |
| `WithMetadata(key, value)` | Adds default metadata to all log entries       |
| `WithFormatter(formatter)` | Sets a custom formatter                        |
| `WithLabelKeys(keys...)`   | Promotes keys to Loki stream labels            |
| `WithMinLevel(level)`      | Sets minimum log level                         |
| `WithTimeFormat(format)`   | Sets timestamp format (LokiFormatter option)   |

## Documentation

For complete documentation, visit [GoDoc](https://godoc.org/github.com/mwazovzky/cloudlog).

## Examples

For more examples, check the [examples directory](https://github.com/mwazovzky/cloudlog/tree/main/examples).

## License

This project is licensed under the MIT License - see the LICENSE file for details.
