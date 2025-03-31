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

# Technical Design Document (TDD) for Logger

## Overview

The `Logger` is a core component responsible for structured logging in the application. It provides a simple interface for logging messages with key-value pairs and integrates with the `Client` to send logs to a Grafana Loki instance.

---

## Architecture

### 1. **Core Components**

- **`Logger`**:
  - The main struct responsible for logging messages.
  - Encapsulates a `Client` instance for sending logs.
- **`Client`**:
  - Handles communication with the Grafana Loki instance.
  - Provides methods for sending log payloads.
- **`LogPayload`**:
  - A struct representing the log payload sent to Loki.
- **`LogStream`**:
  - A struct representing a single log stream within the payload.

---

## Methods

### 1. **NewLogger**

- **Purpose**: Creates a new instance of the `Logger`.
- **Signature**:
  ```go
  func NewLogger(client ClientInterface, service string) *Logger
  ```

### 2. **Log**

- **Purpose**: Logs a message with a specified log level.
- **Signature**:
  ```go
  func (l *Logger) Log(level string, message string, keyValues ...interface{}) error
  ```

### 3. **Info**

- **Purpose**: Logs an informational message.
- **Signature**:
  ```go
  func (l *Logger) Info(message string, keyValues ...interface{}) error
  ```

### 4. **Error**

- **Purpose**: Logs an error message.
- **Signature**:
  ```go
  func (l *Logger) Error(message string, keyValues ...interface{}) error
  ```

### 5. **Debug**

- **Purpose**: Logs a debug message.
- **Signature**:
  ```go
  func (l *Logger) Debug(message string, keyValues ...interface{}) error
  ```

---

## Error Handling

### 1. **Invalid Key-Value Pairs**

- If the key-value pairs are invalid (e.g., odd number of arguments, non-string keys), the `Log` method returns an error.

### 2. **Serialization Errors**

- If the log payload cannot be serialized to JSON, the `Log` method returns an error.

### 3. **Client Errors**

- If the `Client` fails to send the log payload, the `Log` method returns the error from the `Client`.

---

## Extensibility

### 1. **Additional Log Levels**

- The `Log` method can be extended to support additional log levels (e.g., `Warn`, `Fatal`).

### 2. **Custom Metadata**

- The `LogPayload` struct can be extended to include custom metadata (e.g., application version, environment).

---

## Testing Strategy

### 1. **Unit Tests**

- Mock the `Client` to simulate various scenarios (e.g., successful sends, serialization errors, client errors).
- Verify that the `Log`, `Info`, `Error`, and `Debug` methods behave as expected in each scenario.

### 2. **Integration Tests**

- Use a real `Client` to send logs to a test Loki instance.
- Verify that the logs are received and processed correctly.

---

## Test Cases

### 1. Successful Log with String Arguments

- **Description**: Verify that the `Log` method correctly handles string arguments.
- **Expected Outcome**: The `Log` method forms a valid `LogEntry` and sends it successfully.

### 2. Successful Log with Integer Arguments

- **Description**: Verify that the `Log` method correctly handles integer arguments.
- **Expected Outcome**: The `Log` method converts integers to strings and forms a valid `LogEntry`.

### 3. Successful Log with Float Arguments

- **Description**: Verify that the `Log` method correctly handles float arguments.
- **Expected Outcome**: The `Log` method converts floats to strings and forms a valid `LogEntry`.

### 4. Successful Log with Struct Arguments

- **Description**: Verify that the `Log` method correctly handles struct arguments.
- **Expected Outcome**: The `Log` method converts structs to their string representation and forms a valid `LogEntry`.

### 5. Successful Log with Slice Arguments

- **Description**: Verify that the `Log` method correctly handles slice arguments.
- **Expected Outcome**: The `Log` method converts slices to their string representation and forms a valid `LogEntry`.

### 6. Invalid Key-Value Pairs

- **Description**: Verify that the `Log` method returns an error for invalid key-value pairs.
- **Expected Outcome**: The `Log` method returns an error indicating the issue.

### 7. Serialization Error

- **Description**: Verify that the `Log` method returns an error if the log payload cannot be serialized.
- **Expected Outcome**: The `Log` method returns a serialization error.

### 8. Client Error

- **Description**: Verify that the `Log` method returns an error if the `Client` fails to send the log payload.
- **Expected Outcome**: The `Log` method returns the client error.

---

## Tools and Frameworks

### 1. **Testing Framework**

- Use `testify` for assertions and mocking.

### 2. **Mocking**

- Use a `MockClient` to simulate client behavior.

### 3. **Code Coverage**

- Ensure at least 90% test coverage for the `Logger` and `Client` components.

---

## Example Usage

```go
client := NewClient("http://loki-instance", "username", "token", httpClient)
logger := NewLogger(client, "test-service")

logger.Info("Info message", "key1", "value1")
logger.Error("Error message", "key1", "value1")
logger.Debug("Debug message", "key1", "value1")
```
