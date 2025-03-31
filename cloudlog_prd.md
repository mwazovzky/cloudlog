# Product Requirements Document (PRD) for Logger

## Overview

The `Logger` is responsible for structured logging in the application. It formats log messages with key-value pairs, serializes them to JSON, and sends them to a Grafana Loki instance using the `Client`.

---

## Goals

1. Provide a structured logging interface with support for key-value pairs.
2. Integrate seamlessly with the `Client` to send logs to Loki.
3. Ensure logs are formatted and serialized correctly before being sent.
4. Handle errors gracefully and propagate them to the caller.

---

## Non-Goals

1. The `Logger` will not handle advanced log processing (e.g., retries, buffering).
2. The `Logger` will not manage log storage or retrieval.

---

## Functional Requirements

1. **Log Levels**:

   - Support at least three log levels: `Info`, `Error`, and `Debug`.

2. **Structured Logging**:

   - Accept key-value pairs as log metadata.
   - Ensure key-value pairs are validated (e.g., keys must be strings).

3. **Integration with Client**:

   - Use the `Client` to send logs to Loki.
   - Ensure logs are serialized to JSON before being sent.

4. **Error Handling**:
   - Return an error if the log payload cannot be serialized or sent.
   - Handle invalid key-value pairs gracefully.

---

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

---

## Risks and Mitigations

1. **Risk**: Invalid key-value pairs may cause errors.
   - **Mitigation**: Validate key-value pairs before processing.
2. **Risk**: Network failures may cause logs to be lost.
   - **Mitigation**: Propagate errors to the caller for handling.

---

## Success Metrics

1. Logs are successfully sent to Loki in 99.9% of cases.
2. Errors are correctly propagated to the caller.
3. The logger is easy to use and integrates seamlessly with the `Client`.

---

## Example Usage

```go
client := cloudlog.NewClient("http://loki-instance", "username", "token", httpClient)
logger := cloudlog.NewLogger(client)

logger.Info("Info message", "key1", "value1")
logger.Error("Error message", "key1", "value1")
logger.Debug("Debug message", "key1", "value1")
```
