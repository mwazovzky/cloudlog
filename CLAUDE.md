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

- **`logger/`** — `Logger` and `Sender` interfaces, `SyncSender`, and logger implementation. Log methods accept `context.Context`.
- **`formatter/`** — `Formatter` interface returns `[]byte`. `LokiFormatter` (JSON) and `StringFormatter` (human-readable).
- **`client/`** — `LogSender` interface and `LokiClient`. Sends `LokiEntry` to Loki's HTTP push API.
- **`errors/`** — Sentinel errors with `Is*` helpers.

### Data flow

```
Logger.Info(ctx, msg, kv...)
  → formatter.Format(LogEntry) → []byte
  → Sender.Send(ctx, content, labels, timestamp)
  → [SyncSender] builds LokiEntry → LogSender.Send(ctx, LokiEntry) → HTTP POST
```

### Key interfaces

- **`Logger`** — Info/Error/Debug/Warn + With/WithJob
- **`Sender`** — Send(ctx, content, labels, timestamp). SyncSender sends immediately. AsyncSender (future) will buffer.
- **`Formatter`** — Format(LogEntry) → []byte
- **`LogSender`** — Send(ctx, LokiEntry). Low-level HTTP transport.

### Public API

- **Constructors:** `New`, `NewSyncSender`, `NewClient`, `NewLokiFormatter`
- **Logger options:** `WithJob`, `WithMetadata`, `WithFormatter`, `WithLabelKeys`, `WithMinLevel`
- **Level constants:** `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError`
- **Error helpers:** `IsFormatError`, `IsConnectionError`, `IsResponseError`
