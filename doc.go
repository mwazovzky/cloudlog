/*
Package cloudlog provides a structured logging system designed for integration with logging backends
like Grafana Loki. It offers multiple logging strategies including synchronous, asynchronous, and
batch delivery modes.

# Basic Usage

The simplest way to use CloudLog is with the synchronous logger:

	// Create an HTTP client with a reasonable timeout
	httpClient := &http.Client{Timeout: 5 * time.Second}

	// Create a client for your log backend
	client := cloudlog.NewClient("http://loki:3100/loki/api/v1/push", "username", "token", httpClient)

	// Create a logger with synchronous delivery (blocks until logs are sent)
	logger := cloudlog.New(client, cloudlog.WithJob("my-service"))

	// Log a message with structured data
	logger.Info("User logged in", "user_id", "12345", "login_method", "oauth")

	// Log an error
	err := someOperation()
	if err != nil {
		logger.Error("Operation failed", "error", err.Error())
	}

# Asynchronous Logging

For high-throughput applications, use asynchronous logging to avoid blocking:

	// Create an async logger with default settings
	asyncLogger := cloudlog.NewAsync(client, cloudlog.WithJob("my-service"))

	// Log messages without blocking
	for i := 0; i < 1000; i++ {
		asyncLogger.Info("Processing item", "item_id", i)
	}

	// Ensure logs are sent before shutdown
	asyncLogger.Flush()
	asyncLogger.Close()

# Batch Logging

For optimal performance when logging high volumes, use batch logging:

	// Create a batch logger with custom configuration
	config := cloudlog.DefaultDeliveryConfig()
	config.BatchSize = 200
	config.FlushInterval = 2 * time.Second

	batchLogger := cloudlog.NewBatchLoggerWithConfig(client, config)

	// Log messages will be collected in batches
	for i := 0; i < 1000; i++ {
		batchLogger.Info("Processing item", "item_id", i)
	}

	// Ensure logs are sent before shutdown
	batchLogger.Flush()
	batchLogger.Close()

# Adding Context

You can add context to your logs:

	// Create a logger with context
	userLogger := logger.WithContext("user_id", "12345", "request_id", "abc-123")

	// All logs from this logger will include the context
	userLogger.Info("User action performed")

# Error Handling

CloudLog provides detailed error information:

	err := logger.Info("Test message")
	if err != nil {
		if cloudlog.IsConnectionError(err) {
			// Handle connection error
		} else if cloudlog.IsBufferFullError(err) {
			// Handle buffer full error
		}
	}

# Advanced Configuration

For advanced use cases, you can create custom delivery configurations:

	config := cloudlog.DefaultDeliveryConfig()
	config.Async = true
	config.QueueSize = 5000
	config.Workers = 4
	config.MaxRetries = 3
	config.RetryInterval = 500 * time.Millisecond

	customLogger := cloudlog.NewAsyncWithConfig(client, config)

# Custom Deliverers

You can implement your own delivery strategies by implementing the LogDeliverer interface:

	type CustomDeliverer struct {
		// Your implementation details
	}

	// Implement the LogDeliverer interface methods

	// Use your custom deliverer
	logger := cloudlog.NewWithDeliverer(customDeliverer)

# Best Practices

1. Always call Flush() and Close() before shutting down async or batch loggers
2. Set appropriate QueueSize to avoid buffer overflow errors
3. Handle logging errors appropriately for your application
4. Use structured logging (key-value pairs) instead of embedding values in messages
5. Set meaningful job names to identify log sources

# Asynchronous and Batch Logging

CloudLog supports asynchronous and batch logging to optimize performance:

	// Create an async logger
	asyncLogger := cloudlog.NewAsync(client, cloudlog.WithJob("async-logger"))

	// Create a batch logger
	batchConfig := cloudlog.DefaultDeliveryConfig()
	batchConfig.BatchSize = 3
	batchLogger := cloudlog.NewBatchLoggerWithConfig(client, batchConfig)

	// Ensure logs are flushed before shutdown
	asyncLogger.Flush()
	batchLogger.Flush()
*/
package cloudlog
