package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/delivery"
	"github.com/mwazovzky/cloudlog/formatter"
)

func main() {
	// Create a console client for demonstration
	consoleClient := &ConsoleClient{}

	fmt.Println("=== CloudLog Async Delivery Example ===")

	// Create loggers with different configurations
	fmt.Println("\n1. Creating Standard Synchronous Logger")
	syncLogger := cloudlog.New(consoleClient,
		cloudlog.WithJob("sync-logger"),
		cloudlog.WithFormatter(formatter.NewStringFormatter()),
	)

	fmt.Println("\n2. Creating Asynchronous Logger (Default Settings)")
	asyncLogger := cloudlog.NewAsync(consoleClient,
		cloudlog.WithJob("async-logger"),
	)

	fmt.Println("\n3. Creating Asynchronous Logger with Custom Settings")
	customConfig := delivery.DefaultConfig()
	customConfig.Async = true
	customConfig.QueueSize = 1000
	customConfig.Workers = 2
	customConfig.MaxRetries = 3
	customConfig.RetryInterval = time.Second * 1

	customAsyncLogger := cloudlog.NewAsyncWithConfig(consoleClient, customConfig,
		cloudlog.WithJob("custom-async"),
		cloudlog.WithMetadata("logger_type", "custom_async"),
	)

	fmt.Println("\n4. Creating Batch Logger")
	batchConfig := delivery.DefaultConfig()
	batchConfig.Async = true
	batchConfig.BatchSize = 3
	batchConfig.FlushInterval = time.Second * 2

	batchLogger := cloudlog.NewBatchLoggerWithConfig(consoleClient, batchConfig,
		cloudlog.WithJob("batch-logger"),
	)

	// Demonstrate basic logging with each type
	fmt.Println("\n=== Basic Logging with Each Logger Type ===")

	syncLogger.Info("Sync logger message", "sync", true)
	asyncLogger.Info("Async logger message", "async", true)
	customAsyncLogger.Info("Custom async logger message", "custom", true)
	batchLogger.Info("Batch logger message 1", "batch", 1)

	// Demonstrate high-volume async logging
	fmt.Println("\n=== High-Volume Async Logging (100 messages) ===")

	start := time.Now()
	for i := 0; i < 100; i++ {
		asyncLogger.Info("High volume async message", "index", i)
	}
	fmt.Printf("Time to queue 100 async messages: %v\n", time.Since(start))

	// Demonstrate batch logging behavior
	fmt.Println("\n=== Batch Logging (will send after 3 messages or interval) ===")

	batchLogger.Info("Batch logger message 2", "batch", 2)
	fmt.Println("Added second batch message, waiting for one more to trigger batch send...")
	batchLogger.Info("Batch logger message 3", "batch", 3) // This should trigger the batch

	// Demonstrate error handling for buffer overflow
	fmt.Println("\n=== Buffer Overflow Handling ===")

	// Create a logger with a tiny buffer
	tinyConfig := delivery.DefaultConfig()
	tinyConfig.Async = true
	tinyConfig.QueueSize = 2
	tinyConfig.Workers = 1

	// Add a delay to the console client to make overflow more likely
	slowConsoleClient := &SlowConsoleClient{delay: 50 * time.Millisecond}

	overflowLogger := cloudlog.NewAsyncWithConfig(slowConsoleClient, tinyConfig,
		cloudlog.WithJob("overflow-test"),
	)

	fmt.Println("Sending messages to a logger with a small buffer and slow processing...")

	var overflow *sync.WaitGroup = &sync.WaitGroup{}
	overflow.Add(1)

	go func() {
		defer overflow.Done()

		for i := 0; i < 10; i++ {
			err := overflowLogger.Info("Attempting to log", "index", i)
			if err != nil {
				if cloudlog.IsBufferFullError(err) {
					fmt.Printf("Buffer full error detected at message %d\n", i)
					break
				} else {
					fmt.Printf("Unexpected error: %v\n", err)
					break
				}
			}
		}
	}()

	overflow.Wait()

	// Demonstrate graceful shutdown
	fmt.Println("\n=== Graceful Shutdown ===")
	fmt.Println("Sending final messages before shutdown...")

	for i := 0; i < 10; i++ {
		asyncLogger.Info("Final message before shutdown", "index", i)
	}

	fmt.Println("Flushing async loggers...")
	asyncLogger.Flush()
	customAsyncLogger.Flush()
	batchLogger.Flush()
	overflowLogger.Flush()

	fmt.Println("Closing all loggers...")
	syncLogger.Close()
	asyncLogger.Close()
	customAsyncLogger.Close()
	batchLogger.Close()
	overflowLogger.Close()

	fmt.Println("\nExample complete. In a real application, always call Flush() and Close()")
	fmt.Println("before shutdown to ensure all log messages are delivered.")
}

// ConsoleClient implements client.LogSender for demonstration
type ConsoleClient struct{}

func (c *ConsoleClient) Send(job string, formatted []byte) error {
	fmt.Printf("[%s] %s\n", job, string(formatted))
	return nil
}

// SlowConsoleClient simulates a slow backend for demonstrating buffer behavior
type SlowConsoleClient struct {
	delay time.Duration
}

func (c *SlowConsoleClient) Send(job string, formatted []byte) error {
	time.Sleep(c.delay)
	fmt.Printf("[%s] %s\n", job, string(formatted))
	return nil
}

// Example of a real-world application setup with proper signal handling
func realWorldExample() {
	// Create HTTP client for Loki
	httpClient := &http.Client{Timeout: 5 * time.Second}

	// Create Loki client
	lokiClient := cloudlog.NewClient(
		"http://loki:3100/api/v1/push",
		"username",
		"token",
		httpClient,
	)

	// Create async logger
	config := delivery.DefaultConfig()
	config.Async = true
	config.Workers = 2

	logger := cloudlog.NewAsyncWithConfig(lokiClient, config,
		cloudlog.WithJob("my-service"),
		cloudlog.WithMetadata("version", "1.0.0"),
		cloudlog.WithMetadata("env", "production"),
	)

	// Setup signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		fmt.Println("Shutdown signal received, flushing logs...")

		// Ensure all logs are delivered before exiting
		logger.Flush()
		logger.Close()

		fmt.Println("Logger shutdown complete")
		os.Exit(0)
	}()

	// Application code...
	logger.Info("Application started")

	// Main loop...
}
