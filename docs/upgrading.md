# Upgrading to CloudLog v1.1.0

CloudLog v1.1.0 introduces significant improvements to the logging architecture with the addition of asynchronous and batch logging capabilities. This guide will help you understand the changes and how to upgrade from previous versions.

## Key New Features

1. **Asynchronous Logging**: Log messages can be queued and processed in the background
2. **Batch Processing**: Multiple logs can be grouped for efficient delivery
3. **Retry Mechanism**: Failed log attempts can be automatically retried
4. **Delivery Interface**: A new abstraction layer for implementing custom delivery strategies

## Compatibility

v1.1.0 is fully backward compatible with previous versions. All existing code using CloudLog will continue to work without modification.

## Basic Upgrade Steps

### 1. Update the Import

```go
import "github.com/mwazovzky/cloudlog"
```

### 2. Migrate to Async Logging (Optional)

If you want to use the new asynchronous logging features, you can change:

```go
// Before
logger := cloudlog.New(client)

// After - Async Logger
logger := cloudlog.NewAsync(client)
```

### 3. Add Proper Shutdown (Recommended for Async)

When using asynchronous loggers, ensure proper shutdown:

```go
// Before shutdown
logger.Flush() // Ensure pending logs are sent
logger.Close() // Close resources
```

## Advanced Usage

### Custom Configuration

```go
config := cloudlog.DefaultDeliveryConfig()
config.Async = true
config.QueueSize = 1000
config.Workers = 2
config.MaxRetries = 3

logger := cloudlog.NewAsyncWithConfig(client, config)
```

### Batch Logging

```go
config := cloudlog.DefaultDeliveryConfig()
config.BatchSize = 50
config.FlushInterval = time.Second * 5

logger := cloudlog.NewBatchLoggerWithConfig(client, config)
```

### Error Handling

New error types are available:

```go
err := logger.Info("Message")
if err != nil {
    if cloudlog.IsBufferFullError(err) {
        // Handle buffer overflow
    } else if cloudlog.IsTimeoutError(err) {
        // Handle timeout
    } else if cloudlog.IsShutdownError(err) {
        // Handle shutdown errors
    }
}
```

## Performance Considerations

1. **Sync Logging**: Blocks until the log is delivered

   - Good for critical logs where you need confirmation
   - May slow down application during network issues

2. **Async Logging**: Returns immediately, processes logs in background

   - Better performance, doesn't block application code
   - Buffer may fill up during high load or backend issues
   - Requires proper shutdown to avoid losing logs

3. **Batch Logging**: Groups logs for efficiency
   - Best performance for high volume logging
   - Introduces a delay in delivery (based on batch size/interval)
   - Most efficient network utilization

## Best Practices

1. Use synchronous logging for critical operations where confirmation is needed
2. Use asynchronous logging for most general purpose logging
3. Use batch logging for high-volume logging scenarios
4. Always call `Flush()` and `Close()` before shutdown
5. Handle buffer overflow errors appropriately
6. Configure queue size based on expected log volume and memory constraints
7. Set appropriate worker count based on available CPU resources
