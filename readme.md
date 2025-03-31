# Logger Package

The `logger` package provides a structured logging interface for sending logs to Grafana Loki.

## Installation

```bash
go get github.com/mwazovzky/cloudlog/logger
```

## Usage

```go
import (
    "github.com/mwazovzky/cloudlog/logger"
    "github.com/mwazovzky/cloudlog/grafanaclient"
)

func main() {
    client := grafanaclient.NewClient(grafanaclient.Config{
        LokiURL:      "http://loki-instance",
        LokiUsername: "example-username",
        LokiAuthToken: "example-token",
    }, &http.Client{})

    log := logger.NewLogger(client)

    log.Info("example-job", "key1", "value1", "key2", "value2")
}
```

## Features

- Structured logging with key-value pairs.
- Integration with Grafana Loki.
- Easy-to-use interface.

## Logger Implementation

### Overview

The `logger` package uses synchronous API calls to send logs to Grafana Loki. While this approach is simple and effective for many use cases, it introduces certain risks and limitations that should be considered.

### Issues and Risks

#### 1. Blocking Behavior

- **Description**: Each log message is sent to Loki in a blocking manner. If the Loki server is slow to respond or there are network issues, the application may experience delays.
- **Impact**:
  - Increased latency in critical application paths.
  - Potential bottlenecks in high-throughput scenarios.

#### 2. Network Failures

- **Description**: If the network connection to Loki is unstable or unavailable, the `logger` will return an error for each failed log attempt.
- **Impact**:
  - Logs may be lost if errors are not handled properly.
  - Repeated failures can degrade application performance.

#### 3. No Retry Mechanism

- **Description**: The current implementation does not include retries for failed log attempts.
- **Impact**:
  - Temporary network issues or server unavailability can result in lost logs.
  - Applications requiring guaranteed log delivery may need to implement retries externally.

#### 4. High Volume Logging

- **Description**: In scenarios with a high volume of logs, the synchronous nature of the API calls can overwhelm the Loki server or the network.
- **Impact**:
  - Increased risk of throttling or rate-limiting by the Loki server.
  - Higher resource consumption (CPU, memory, and network bandwidth).

### Feature requests

#### 1. Asynchronous Logging

- Implement an asynchronous logging mechanism where log messages are queued and sent to Loki in the background. This reduces blocking behavior and improves application performance.

#### 2. Batching Logs

- Group multiple log messages into a single payload and send them to Loki in batches. This reduces the number of API calls and improves efficiency.

#### 3. Retry Mechanism

- Add a retry mechanism with exponential backoff for failed log attempts. This improves reliability during temporary network issues.

#### 4. Circuit Breaker

- Implement a circuit breaker pattern to temporarily stop sending logs if the Loki server is unavailable or experiencing issues. This prevents overwhelming the server and reduces resource consumption.

#### 5. Monitoring and Metrics

- Add monitoring and metrics to track the performance and reliability of the `logger`. This helps identify bottlenecks and issues in production.

### Conclusion

While the current synchronous implementation of the `logger` is simple and effective for many use cases, it may not be suitable for high-throughput or latency-sensitive applications. By addressing the issues and risks outlined above, the `logger` can be made more robust and scalable.
