# CloudLog - Product Requirements Document

## Overview

CloudLog is a structured logging library designed to seamlessly integrate with Grafana Loki and other log aggregation services. It provides key-value-based structured logging with context propagation, flexible formatting options, and both synchronous and asynchronous logging capabilities.

## Goals

1. Provide a simple, consistent API for structured logging.
2. Support Grafana Loki integration with minimal configuration.
3. Enable flexible formatting options (JSON and human-readable).
4. Implement efficient context propagation.
5. Offer both blocking and non-blocking logging implementations.
6. Ensure robust error handling and reporting.
7. Keep external dependencies to a minimum.

## Non-Goals

1. Supporting every possible logging backend (focusing on Loki)
2. Implementing complex log aggregation or analysis features
3. Managing log storage or retrieval mechanisms

## User Personas

### Application Developer

- Needs clean, simple logging API
- Wants type-safe structured logging
- Prefers minimal configuration

### DevOps Engineer

- Needs central log aggregation
- Wants configurable log fields and formats
- Requires proper labeling for efficient querying

### High-Traffic Service Owner

- Needs non-blocking logging to maintain performance
- Wants batch processing for efficiency
- Requires graceful handling of high log volumes

### QA Engineer

- Needs predictable log outputs
- Wants to verify logging in tests

## Requirements

### Functional Requirements

1. **Structured Logging**

   - Support multiple log levels (info, error, debug, warn)
   - Enable key-value pair logging
   - Allow timestamp format customization

2. **Loki Integration**

   - Implement proper Loki push API protocol
   - Support authentication
   - Enable stream labeling
   - Provide nanosecond timestamp precision

3. **Formatting**

   - Support JSON formatting for machine consumption
   - Support human-readable string formatting for console output
   - Allow field name customization

4. **Context & Metadata**

   - Support context propagation across logger instances
   - Enable additional metadata for all logs
   - Provide job/service identification

5. **Synchronous Logging**

   - Implement blocking behavior for simple use cases
   - Return per-log errors for immediate handling

6. **Asynchronous Logging**

   - Implement non-blocking behavior for high-throughput scenarios
   - Use a buffered queue for log entries
   - Support configurable batching for efficient processing
   - Use worker goroutines to process logs in the background
   - Handle flush markers to ensure all logs are processed during flush or shutdown

7. **Error Handling**
   - Return typed errors for specific failure scenarios
   - Provide specific errors for buffer full and shutdown scenarios

### Non-Functional Requirements

1. **Performance**

   - Minimize allocations
   - Optimize serialization
   - Batch processing for high volumes
   - Non-blocking API option

2. **Reliability**

   - No runtime panics
   - Clean closure of resources
   - Buffer overflow protection

3. **Usability**

   - Intuitive API design
   - Comprehensive documentation
   - Clear examples for both sync and async use cases

4. **Extensibility**
   - Support custom formatters
   - Enable custom log senders
   - Allow flexible configuration

## Success Metrics

1. 99.9% successful log delivery to Loki
2. > 85% code coverage in tests
3. Minimal overhead (<1ms per log entry for sync, <0.1ms for async)
4. Clear and consistent error reporting
5. Zero impact on application performance in async mode
