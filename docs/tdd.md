# Technical Design Document — CloudLog

## Overview

CloudLog is a structured logging library for Go with Grafana Loki integration. It provides key-value pair logging, context propagation, level filtering, and pluggable formatting.

## Architecture

### Layers

```
cloudlog (facade) → logger → formatter + sender → client → HTTP
```

Each layer has a single responsibility:

| Layer       | Responsibility                          | Key Interface |
| ----------- | --------------------------------------- | ------------- |
| `cloudlog`  | Public API facade, re-exports           | —             |
| `logger`    | Formatting, metadata, level filtering   | `Logger`      |
| `sender`    | Delivery strategy, protocol translation | `Sender`      |
| `formatter` | Content serialization (JSON, string)    | `Formatter`   |
| `client`    | HTTP transport to Loki                  | `LogSender`   |
| `errors`    | Sentinel errors, classification         | —             |

### Interfaces

```go
// Logger — structured logging with metadata propagation
type Logger interface {
    Info(ctx context.Context, message string, keyvals ...interface{}) error
    Error(ctx context.Context, message string, keyvals ...interface{}) error
    Debug(ctx context.Context, message string, keyvals ...interface{}) error
    Warn(ctx context.Context, message string, keyvals ...interface{}) error
    With(keyvals ...interface{}) Logger
    WithJob(job string) Logger
}

// Sender — delivers formatted log entries
type Sender interface {
    Send(ctx context.Context, content []byte, labels map[string]string, timestamp time.Time) error
}

// Formatter — serializes log entries to bytes
type Formatter interface {
    Format(entry LogEntry) ([]byte, error)
}

// LogSender — low-level transport (Loki HTTP push)
type LogSender interface {
    Send(ctx context.Context, entry LokiEntry) error
}
```

### Data Flow

```
logger.Info(ctx, "user login", "user_id", "123")
  │
  ├── 1. Level check (skip if below minLevel)
  ├── 2. Merge message + keyvals + default metadata
  ├── 3. Create LogEntry (timestamp = time.Now())
  ├── 4. Extract labelKeys from LogEntry → labels map (remove from content)
  ├── 5. Formatter.Format(entry) → []byte (JSON or string)
  │
  └── Sender.Send(ctx, content, labels, timestamp)
        │
        └── SyncSender: build LokiEntry → LogSender.Send(ctx, entry)
              │
              └── LokiClient: JSON marshal → HTTP POST with basic auth
```

## Design Decisions

### Logger takes Sender, not LogSender

The logger delegates delivery to a `Sender` interface, not directly to HTTP transport. This separation allows different delivery strategies (sync, async) without changing the logger.

```go
// Sync: blocks on HTTP per log call
sender := cloudlog.NewSyncSender(lokiClient)

// Async (future): buffers and batches in background
sender := cloudlog.NewAsyncSender(lokiClient, ...)

// Same logger either way
logger := cloudlog.New(sender, cloudlog.WithJob("my-service"))
```

### Formatter returns []byte, not LokiEntry

Formatters produce content bytes only. The Loki protocol envelope (`LokiEntry` with streams, labels, timestamps) is built by the sender. This keeps formatters transport-agnostic.

### context.Context on all log methods

Follows Go standard library conventions (slog, database/sql). Context flows through to `http.NewRequestWithContext`, enabling cancellation and tracing.

### With() returns new logger (immutable)

`With()` and `WithJob()` create new logger instances with copied metadata. No shared mutable state between derived loggers. Safe for concurrent use.

### Label keys extracted before formatting

When `WithLabelKeys("user_id")` is set, the logger removes promoted keys from `LogEntry.KeyVals` before formatting. This avoids duplication between Loki stream labels and log content.

### Level filtering skips silently

`WithMinLevel(LevelWarn)` causes `Debug` and `Info` calls to return `nil` immediately without formatting or sending. No error, no allocation.

## Error Handling

Sentinel errors with `fmt.Errorf("%w: ...")` wrapping:

| Error              | When                          | Returned by      |
| ------------------ | ----------------------------- | ---------------- |
| `ErrInvalidFormat` | JSON marshaling fails         | Formatter, Client|
| `ErrConnectionFailed` | HTTP request fails         | Client           |
| `ErrResponseError` | Loki returns HTTP 4xx/5xx     | Client           |
| `ErrInvalidInput`  | Malformed request URL         | Client           |

Callers classify errors with `IsFormatError()`, `IsConnectionError()`, `IsResponseError()`.

## Package Structure

```
cloudlog.go              — facade
client/
  client.go              — LokiClient, LogSender, HTTPClient interfaces
errors/
  errors.go              — sentinel errors
formatter/
  formatter.go           — Formatter interface
  entry.go               — LogEntry type
  loki_formatter.go      — JSON formatter (default)
  string_formatter.go    — human-readable formatter
logger/
  interfaces.go          — Logger, Sender interfaces
  logger.go              — logger implementation, options
  sender.go              — SyncSender
  async_sender.go        — AsyncSender
```

## Configuration

### Logger Options

| Option            | Default       | Description                        |
| ----------------- | ------------- | ---------------------------------- |
| `WithJob`         | "application" | Loki stream label                  |
| `WithFormatter`   | LokiFormatter | Content serializer                 |
| `WithMetadata`    | (none)        | Default key-value pairs            |
| `WithLabelKeys`   | (none)        | Keys to promote to stream labels   |
| `WithMinLevel`    | LevelDebug    | Minimum level to send              |

### Log Levels

```go
LevelDebug = 0  // all messages
LevelInfo  = 1
LevelWarn  = 2
LevelError = 3  // errors only
```

## AsyncSender

`AsyncSender` implements `Sender` with non-blocking, buffered delivery. A background worker batches entries and sends them to the underlying `LogSender`.

### Data Flow

```
Send(ctx, content, labels, timestamp)
  → push entry to buffer channel (non-blocking)
  → background worker:
      → accumulate entries until batchSize or flushInterval
      → group entries by full label set (including any keys added via WithLabelKeys)
      → build a single batched LokiEntry with one stream per distinct label set
      → LogSender.Send(ctx with sendTimeout, batchedEntry)
      → on error: call errorHandler
```

### Flush and Close

`Flush()` pushes a flush marker (entry with a response channel) into the buffer. When the worker encounters it, it sends the current partial batch, then closes the response channel. `Flush()` blocks until the response channel is closed.

`Close()` calls `Flush()` to drain remaining entries, then signals the worker to stop. Both `Flush()` and `Close()` return immediately if the sender is already closed.

Both methods live on `AsyncSender`, not on `Logger`.

### Error Handling

- `Send()` returns `ErrBufferFull` if the buffer channel is full (non-blocking mode)
- Background HTTP errors are passed to `errorHandler` callback (default: log to stderr)

### AsyncSender Options

| Option              | Default | Description                       |
| ------------------- | ------- | --------------------------------- |
| `WithBufferSize`    | 1000    | Buffer channel capacity           |
| `WithBatchSize`     | 100     | Max entries per HTTP request      |
| `WithFlushInterval` | 5s      | Max time between sends            |
| `WithBlockOnFull`   | false   | Block vs return ErrBufferFull     |
| `WithErrorHandler`  | stderr  | Callback for background errors    |
| `WithSendTimeout`   | 30s     | Timeout per HTTP batch send       |
