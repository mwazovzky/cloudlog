package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/formatter"
)

// This example demonstrates asynchronous logging with CloudLog.
//
// To run with a real Loki instance:
// 1. Create a .env file with your Loki credentials
// 2. Run: go run main.go
//
// Required environment variables:
//   - LOKI_URL: URL of your Loki instance
//   - LOKI_USERNAME: Username for authentication
//   - LOKI_AUTH_TOKEN: Authentication token
//
// If these environment variables aren't set, the example will log to console instead.

func main() {
	// Get configuration from environment or use defaults
	lokiURL := getEnvOrDefault("LOKI_URL", "")
	lokiUsername := getEnvOrDefault("LOKI_USERNAME", "")
	lokiToken := getEnvOrDefault("LOKI_AUTH_TOKEN", "")

	var asyncLogger cloudlog.Logger

	if lokiURL != "" && lokiUsername != "" && lokiToken != "" {
		// Create a new HTTP client with reasonable timeout
		httpClient := &http.Client{
			Timeout: 5 * time.Second,
		}
		// Create Loki client
		client := cloudlog.NewClient(lokiURL, lokiUsername, lokiToken, httpClient)
		fmt.Println("Sending logs to Loki instance at:", lokiURL)

		// Create async logger with custom configuration
		asyncLogger = cloudlog.NewAsync(
			client,
			cloudlog.WithAsyncJob("async-example-service"),
			cloudlog.WithBufferSize(10000),            // Buffer up to 10,000 logs
			cloudlog.WithBatchSize(100),               // Send in batches of 100
			cloudlog.WithFlushInterval(1*time.Second), // Flush at least every second
			cloudlog.WithWorkers(4),                   // Use 4 worker goroutines
			cloudlog.WithBlockOnFull(false),           // Don't block when buffer full
		)
	} else {
		// Create console output logger if Loki credentials aren't provided
		fmt.Println("Loki credentials not provided, logging to console instead")
		consoleClient := &consoleClient{}

		// Create async logger for console
		asyncLogger = cloudlog.NewAsync(
			consoleClient,
			cloudlog.WithAsyncJob("async-example-service"),
			cloudlog.WithAsyncFormatter(formatter.NewStringFormatter()),
			cloudlog.WithBufferSize(1000),
			cloudlog.WithBatchSize(10),
			cloudlog.WithFlushInterval(200*time.Millisecond),
		)
	}

	// Initial application log
	if err := asyncLogger.Info("Application started", "version", "1.0.0"); err != nil {
		fmt.Printf("Error logging: %v\n", err)
	}

	// Simulate high-volume concurrent logging
	var wg sync.WaitGroup
	numRoutines := 10
	logsPerRoutine := 100

	fmt.Printf("Generating %d logs from %d goroutines...\n", numRoutines*logsPerRoutine, numRoutines)

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			// Create a context-specific logger for this routine
			routineLogger := asyncLogger.WithContext("routine_id", routineID)

			for j := 0; j < logsPerRoutine; j++ {
				// Add some randomization to simulate real-world patterns
				level := randomLevel()

				var err error
				switch level {
				case "info":
					err = routineLogger.Info("Processing item",
						"item_id", fmt.Sprintf("item-%d-%d", routineID, j),
						"duration_ms", rand.Intn(500),
					)
				case "warn":
					err = routineLogger.Warn("Slow operation detected",
						"item_id", fmt.Sprintf("item-%d-%d", routineID, j),
						"duration_ms", 500+rand.Intn(1000),
					)
				case "error":
					err = routineLogger.Error("Operation failed",
						"item_id", fmt.Sprintf("item-%d-%d", routineID, j),
						"error_code", 500+rand.Intn(100),
						"retryable", rand.Intn(2) == 1,
					)
				case "debug":
					err = routineLogger.Debug("Processing details",
						"item_id", fmt.Sprintf("item-%d-%d", routineID, j),
						"memory_used", rand.Intn(1024),
						"cpu_time_ms", rand.Intn(100),
					)
				}

				if err != nil {
					fmt.Printf("Error logging: %v\n", err)
				}

				// Simulate varying workload
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}(i)
	}

	// Wait for all logging goroutines to finish
	wg.Wait()

	fmt.Println("All logs generated, flushing...")

	// Ensure all logs are sent before exiting
	if err := asyncLogger.Flush(); err != nil {
		fmt.Printf("Error flushing logs: %v\n", err)
	}

	// Gracefully close the logger
	if err := asyncLogger.Close(); err != nil {
		fmt.Printf("Error closing logger: %v\n", err)
	}

	fmt.Println("Async logging example completed successfully")
}

// randomLevel returns a random log level with realistic distribution
func randomLevel() string {
	r := rand.Intn(100)
	switch {
	case r < 70: // 70% info
		return "info"
	case r < 85: // 15% debug
		return "debug"
	case r < 95: // 10% warn
		return "warn"
	default: // 5% error
		return "error"
	}
}

// getEnvOrDefault gets environment variable or returns default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// consoleClient implements the LogSender interface to print logs to console
type consoleClient struct{}

func (c *consoleClient) Send(entry client.LokiEntry) error {
	for _, stream := range entry.Streams {
		for _, value := range stream.Values {
			if len(value) >= 2 {
				fmt.Printf("[%s] %s\n", stream.Stream["job"], value[1])
			}
		}
	}
	return nil
}
