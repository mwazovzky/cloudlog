# Technical Design Document (TDD) for CloudLog

## Overview

CloudLog provides structured logging with Grafana Loki support. It offers a simple API for creating and formatting log messages, then dispatches them via a client.

## Architecture Overview

CloudLog follows a modular design with clear separation of concerns between logging, formatting, and transport. The system uses composition over inheritance and emphasizes interface-based design.

### Core Components

1. **Logger Interface**

   - Defines the core logging operations (Info, Error, Debug, Warn)
   - Specifies context propagation methods (WithContext, WithJob)
   - Includes lifecycle methods (Flush, Close)

2. **SyncLogger Implementation**

   - Implements Logger with synchronous, blocking behavior
   - Formats and sends each log entry immediately
   - Returns errors directly to caller

3. **Client**

   - Implements protocol-specific transport (Loki HTTP push API)
   - Manages network communication
   - Handles authentication and error wrapping

4. **Formatter**

   - Transforms log entries into wire format
   - Supports Loki protocol and human-readable formats
   - Enables customization via functional options

5. **Errors**
   - Provides sentinel error values
   - Implements type-safe error checking
   - Wraps underlying errors with context

## Data Flow

1. Application calls logger.Info/Error/Debug/Warn
2. SyncLogger creates log entry with timestamp and level
3. SyncLogger formats entry using configured formatter
4. SyncLogger sends formatted log via client
5. Error (if any) is returned immediately to caller

## Interface Contracts

### Logger Interface

```go
type Logger interface {
  Info(message string, keyvals ...interface{}) error
  Error(message string, keyvals ...interface{}) error
  Debug(message string, keyvals ...interface{}) error
  Warn(message string, keyvals ...interface{}) error
  WithContext(keyvals ...interface{}) Logger
  WithJob(job string) Logger
  Flush() error
  Close() error
}
```

### LogSender Interface

```go
type LogSender interface {
  Send(entry LokiEntry) error
}
```

### Formatter Interface

```go
type Formatter interface {
  Format(entry LogEntry) (client.LokiEntry, error)
}
```

## Extension Points

1. **Custom Formatters**

   - Implement the Formatter interface
   - Register with logger via WithFormatter option

2. **Custom Log Senders**

   - Implement the LogSender interface
   - Pass directly to NewSync constructor

3. **Configuration Options**
   - Functional options pattern for all components
   - Namespaced options prevent naming conflicts

## Error Handling Strategy

1. **Sentinel Errors**

   - Predefined error values for specific failure types
   - Helper functions for type checking (IsConnectionError, IsFormatError, etc.)

2. **Error Propagation**
   - All errors returned to caller

## Testing Strategy

1. **Unit Tests**

   - Component isolation through mocking
   - Behavior verification with testify assertions
   - Edge case coverage

2. **Integration Tests**

   - Example-based testing in examples/ directory
   - Optionally connects to real Loki instance
   - Verifies end-to-end flows

## Running Tests

```bash
go test ./... -v
```

Generate coverage reports with:

```bash
go test ./... -coverprofile=cover.out
go tool cover -html=cover.out
```

Aim for at least 85% coverage.

## Test Categories

### 1. Client Tests

- HTTP request creation
- Authentication header setup
- Error handling for network failures
- Error handling for non-2xx responses
- Timeout handling
- Loki payload format validation

### 2. Formatter Tests

- Correct JSON serialization
- Field naming customization
- Time formatting
- String formatting

### 3. Logger Tests

- Log level handling
- Context propagation
- Metadata inclusion
- Job name setting
- Error propagation from formatters and clients

### 4. Error Package Tests

- Error type identification
- Error wrapping
- Error message construction
- Error chain traversal

## Mocking Strategy

CloudLog uses explicit mock implementations for testing:

- **MockClient**: Implements the LogSender interface
- **MockFormatter**: Implements the Formatter interface
