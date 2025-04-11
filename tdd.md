# Product Requirements Document (PRD) for Logger

## Overview

The `Logger` is a core component responsible for structured logging in the application. It provides a simple interface for logging messages with key-value pairs and integrates with the `Client` to send logs to a Grafana Loki instance.

## Goals

1. Provide a structured logging interface with support for key-value pairs.
2. Integrate seamlessly with the `Client` to send logs to Loki.
3. Ensure logs are formatted and serialized correctly before being sent.
4. Handle errors gracefully and propagate them to the caller.

## Non-Goals

1. The `Logger` will not handle advanced log processing (e.g., retries, buffering).
2. The `Logger` will not manage log storage or retrieval.

## Functional Requirements

1. **Log Levels**:

   - Support at least two log levels: `Info` and `Error`.

2. **Structured Logging**:

   - Accept key-value pairs as log metadata.
   - Ensure key-value pairs are validated (e.g., keys must be strings).

3. **Integration with Client**:

   - Use the `Client` to send logs to Loki.
   - Ensure logs are serialized to JSON before being sent.

4. **Error Handling**:
   - Return an error if the log payload cannot be serialized or sent.
   - Handle invalid key-value pairs gracefully.

## Technical Requirements

1. **Dependencies**:

   - Use the `Client` interface for sending logs.
   - Use the `encoding/json` package for JSON serialization.

2. **Performance**:

   - Minimize memory allocations during log formatting and serialization.
   - Ensure efficient handling of large log payloads.

3. **Extensibility**:
   - Allow additional log levels to be added in the future.
   - Support custom log metadata (e.g., timestamps, job names).

## Risks and Mitigations

1. **Risk**: Invalid key-value pairs may cause errors.
   - **Mitigation**: Validate key-value pairs before processing.
2. **Risk**: Network failures may cause logs to be lost.
   - **Mitigation**: Propagate errors to the caller for handling.

## Success Metrics

1. Logs are successfully sent to Loki in 99.9% of cases.
2. Errors are correctly propagated to the caller.
3. The logger is easy to use and integrates seamlessly with the `Client`.

# Technical Design Document (TDD) for CloudLog

## Overview

CloudLog provides structured logging with Grafana Loki support. It offers a simple API for creating and formatting log messages, then dispatches them via a client.

## Architecture

### Core Components

1. Logger

   - Interface and SyncLogger implementation.
   - Supports info, error, debug, warn levels.
   - Holds context (job name, metadata).

2. Client

   - Sends data to Loki.
   - Manages network operations, authentication, and errors.

3. Formatter

   - Formats LogEntry to JSON or string.
   - Supports custom formatters.

4. Errors
   - Provides sentinel errors (ErrConnectionFailed, ErrInvalidFormat, etc.).

## Key Flows

1. Creating a Logger

   - Use cloudlog.NewSync(...) with a client.
   - Provide options (WithJob, WithFormatter).

2. Logging a Message

   - Logger methods (Info, Error, Debug, Warn) create a LogEntry, apply context, format, and send it.

3. Error Propagation
   - Formatting errors return ErrInvalidFormat.
   - Network errors return ErrConnectionFailed.
   - Server errors return ErrResponseError.

## Extensibility

- Pluggable formatters (implement Formatter interface).
- Custom clients (implement LogSender).
- Customizable fields (job, level, timestamp) via options.

## Testing Strategy

1. Unit Tests

   - Mock the network client and check payload.
   - Verify log method structure.

2. Error Handling
   - Force errors.
   - Verify sentinel errors.

## Test Cases

### Client Package

1. Loki Payload Format Validation
   - Verify correct Loki payload structure.
   - Verify job name as label.
   - Verify nanosecond timestamp.

# CloudLog - Test-Driven Development Guide

## Unified Test Structure

1. **Unit Tests**: Test individual components in isolation.
2. **Integration Tests**: Test interactions between components.
3. **Error Tests**: Verify error identification and propagation.
4. **Performance Tests**: Measure throughput and memory usage.

## Assertion Patterns

- Use `assert` for soft assertions and `require` for hard assertions.
- Prefer table-driven tests for repetitive scenarios.

## Running Tests

Run all tests with:

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

The client package tests verify:

- HTTP request creation
- Authentication header setup
- Error handling for network failures
- Error handling for non-2xx responses
- Timeout handling
- Loki payload format validation, ensuring:
  - Proper streams structure
  - Correct job labels
  - Nanosecond timestamps
  - Properly structured values array

### 2. Formatter Tests

Formatter tests verify:

- Correct JSON serialization
- Field naming customization
- Time formatting
- String formatting
- Error handling for non-serializable values

### 3. Logger Tests

Logger tests verify:

- Log level handling
- Context propagation
- Metadata inclusion
- Job name setting
- Error propagation from formatters and clients

### 4. Error Package Tests

Error tests verify:

- Error type identification
- Error wrapping
- Error message construction
- Error chain traversal

### 5. Testing Package Tests

Testing utilities tests verify:

- Log capture
- Filtering by level and content
- Searching by field values

## Mocking Strategy

CloudLog uses explicit mock implementations for testing:

- **MockClient**: Implements the LogSender interface
- **MockFormatter**: Implements the Formatter interface
- **TestLogger**: Special implementation for capturing logs

## Assertion Patterns

For consistent test writing, use these patterns:

### Error Testing

```go
err := someOperation()
if err != nil {
    t.Errorf("Expected no error, got: %v", err)
}
```

### Specific Error Type Testing

```go
if !errors.Is(err, errors.ErrConnectionFailed) {
    t.Errorf("Expected connection error, got: %v", err)
}
```

### Value Testing

```go
if result["level"] != "info" {
    t.Errorf("Expected level 'info', got: %v", result["level"])
}
```

## Test Cases

### Client Package Test Cases

#### 1. Loki Payload Format Validation

- **Description**: Verify that the client creates a correctly formatted Loki payload
- **Expected Outcome**: The payload should have the proper structure with streams, labels, and values
- **Test Steps**:
  1. Create a client and send a log message
  2. Capture the HTTP request payload
  3. Validate the payload structure matches Loki's API requirements
  4. Verify job name is included as a label
  5. Verify timestamp is properly formatted as a nanosecond timestamp

# Technical Design Document for CloudLog

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

3. **AsyncLogger Implementation**

   - Implements Logger with asynchronous, non-blocking behavior
   - Buffers log entries for batch processing
   - Uses worker goroutines to process logs in background
   - Only returns errors when buffer is full or logger is closed

4. **Client**

   - Implements protocol-specific transport
   - Manages network communication
   - Handles authentication and error wrapping

5. **Formatter**

   - Transforms log entries into wire format
   - Supports Loki protocol and human-readable formats
   - Enables customization via functional options

6. **Errors**
   - Provides sentinel error values
   - Implements type-safe error checking
   - Wraps underlying errors with context

## Data Flow

### Synchronous Flow

1. Application calls logger.Info/Error/Debug/Warn
2. SyncLogger creates log entry with timestamp and level
3. SyncLogger formats entry using configured formatter
4. SyncLogger sends formatted log via client
5. Error (if any) is returned immediately to caller

### Asynchronous Flow

1. Application calls logger.Info/Error/Debug/Warn
2. AsyncLogger creates async log entry containing job, level, message, and metadata
3. AsyncLogger sends entry to internal buffer channel
4. Worker goroutines pull entries from buffer
5. Workers accumulate entries until batch size or interval reached
6. Workers format and send batches via client
7. Errors during processing are handled internally (not returned to caller)

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
  Send(job string, formatted []byte) error
}
```

### Formatter Interface

```go
type Formatter interface {
  Format(entry LogEntry) ([]byte, error)
}
```

## Extension Points

1. **Custom Formatters**

   - Implement the Formatter interface
   - Register with logger via WithFormatter option

2. **Custom Log Senders**

   - Implement the LogSender interface
   - Pass directly to NewSync or NewAsync constructor

3. **Configuration Options**
   - Functional options pattern for all components
   - Namespaced options prevent naming conflicts
   - AsyncLogger-specific options for buffer and batch control

## AsyncLogger Design Details

### Buffer Management

The AsyncLogger uses a Go channel as its internal buffer, which provides:

- Thread-safe access from multiple goroutines
- Efficient producer-consumer pattern
- Blocking behavior when configured to do so

### Batch Processing

1. Worker goroutines consume entries from the buffer
2. Entries are accumulated until one of:
   - Batch size is reached
   - Flush interval elapses
   - Flush is explicitly called
   - Logger is closed
3. Batches are grouped by job name for efficient sending
4. Each batch is formatted and sent via the client

### Error Handling

- Buffer full: Returns ErrBufferFull or blocks depending on configuration
- Logger closed: Returns ErrShutdown
- Send errors: Logged or dropped (no way to return to caller)
- Format errors: Logged or dropped (no way to return to caller)

### Concurrency Safety

- WithContext/WithJob return new loggers that reference shared buffer
- All loggers from a single NewAsync share worker pool and buffer
- Close() is synchronized to prevent concurrent shutdown issues
- Metadata is copied to avoid race conditions

## Error Handling Strategy

1. **Sentinel Errors**

   - Predefined error values for specific failure types
   - Helper functions for type checking (IsConnectionError, etc.)
   - New AsyncLogger-specific errors like ErrBufferFull

2. **Error Propagation**
   - Synchronous logger: All errors returned to caller
   - Asynchronous logger: Only buffer-full or closed errors returned to caller

## Testing Strategy

The package follows extensive test-driven development with:

1. **Unit Tests**

   - Component isolation through mocking
   - Behavior verification with testify assertions
   - Edge case coverage
   - AsyncLogger-specific tests for buffering, batching, and worker behavior

2. **Integration Tests**

   - Example-based testing in examples/ directory
   - Optionally connects to real Loki instance
   - Verifies end-to-end flows

3. **Performance Tests**
   - Benchmarks for both synchronous and asynchronous loggers
   - Measurement of throughput under load
   - Memory allocation patterns

## Performance Considerations

1. **Synchronous Logger**

   - Simple and direct, good for low-volume logging
   - Each log call blocks until complete
   - Network latency impacts application performance

2. **Asynchronous Logger**
   - Non-blocking for high-throughput scenarios
   - Buffers messages to smooth out processing
   - Batches requests for better network efficiency
   - Minimal impact on application latency
   - Consider memory usage when configuring buffer size
