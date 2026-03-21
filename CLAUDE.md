# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CloudLog is a Go structured logging library for Grafana Loki integration. Module path: `github.com/mwazovzky/cloudlog`. Go 1.23+.

## Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./logger/
go test ./formatter/
go test ./client/

# Run a single test
go test -run TestFunctionName ./package/

# Run tests with race detection and coverage (excludes examples/)
./coverage.sh
```

## Architecture

The root package (`cloudlog.go`) is a **facade** that re-exports types and constructors from internal packages. Users import only `github.com/mwazovzky/cloudlog`.

### Core packages

- **`logger/`** — `Logger` interface and `SyncLogger` implementation. Blocks until each log entry is sent.
- **`formatter/`** — `Formatter` interface with `LokiFormatter` (JSON, default) and `StringFormatter` (human-readable, for development).
- **`client/`** — `LogSender` interface and `LokiClient`. Sends `LokiEntry` structs as JSON to Loki's HTTP push API with basic auth.
- **`errors/`** — Sentinel errors (`ErrInvalidFormat`, `ErrConnectionFailed`, `ErrResponseError`) with `Is*` helpers.

### Data flow

`Logger.Info(msg, keyvals...)` → `formatter.Format(LogEntry)` → `client.Send(LokiEntry)` → HTTP POST to Loki

### Public API (via facade)

- **Constructors:** `NewSync`, `NewClient`, `NewLokiFormatter`
- **Logger options:** `WithJob`, `WithMetadata`, `WithFormatter`
- **Formatter options:** `WithLabelKeys`, `WithTimeFormat`
- **Error helpers:** `IsFormatError`, `IsConnectionError`, `IsResponseError`
