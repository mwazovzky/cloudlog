# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CloudLog is a Go structured logging library for Grafana Loki integration. Module path: `github.com/mwazovzky/cloudlog`. Go 1.23+.

## Commands

```bash
go test ./...                              # Run all tests
go test ./logger/                          # Run tests for a specific package
go test -run TestFunctionName ./package/   # Run a single test
./coverage.sh                              # Race detection + coverage
```

## Architecture

The root package (`cloudlog.go`) is a **facade** re-exporting from internal packages.

### Core packages

- **`logger/`** — `Logger` interface and `SyncLogger`. Log methods accept `context.Context` as first arg.
- **`formatter/`** — `Formatter` interface returns `[]byte`. `LokiFormatter` (JSON) and `StringFormatter` (human-readable).
- **`client/`** — `LogSender` interface and `LokiClient`. Sends `LokiEntry` to Loki's HTTP push API. `Send` accepts `context.Context`.
- **`errors/`** — Sentinel errors with `Is*` helpers.

### Data flow

`Logger.Info(ctx, msg, kv...)` → `formatter.Format(LogEntry) → []byte` → logger builds `LokiEntry` with labels → `client.Send(ctx, LokiEntry)` → HTTP POST

### Public API

- **Constructors:** `NewSync`, `NewClient`, `NewLokiFormatter`
- **Logger options:** `WithJob`, `WithMetadata`, `WithFormatter`, `WithLabelKeys`, `WithMinLevel`
- **Formatter options:** `WithTimeFormat`
- **Error helpers:** `IsFormatError`, `IsConnectionError`, `IsResponseError`
- **Level constants:** `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError`
