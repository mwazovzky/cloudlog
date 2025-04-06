package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/delivery"
	"github.com/mwazovzky/cloudlog/formatter"
)

// This example demonstrates basic usage of the CloudLog library.
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

	var syncLogger *cloudlog.Logger

	if lokiURL != "" && lokiUsername != "" && lokiToken != "" {
		// Create a new HTTP client with reasonable timeout
		httpClient := &http.Client{
			Timeout: 5 * time.Second,
		}
		// Create Loki client
		client := cloudlog.NewClient(lokiURL, lokiUsername, lokiToken, httpClient)
		fmt.Println("Sending logs to Loki instance at:", lokiURL)
		// Create explicit sync logger
		syncDeliverer := delivery.NewSyncDeliverer(client)
		syncLogger = cloudlog.NewWithDeliverer(syncDeliverer, cloudlog.WithJob("sync-example-service"))
	} else {
		// Create console output logger if Loki credentials aren't provided
		fmt.Println("Loki credentials not provided, logging to console instead")
		consoleClient := &consoleClient{}
		// Create explicit sync logger
		syncDeliverer := delivery.NewSyncDeliverer(consoleClient)
		syncLogger = cloudlog.NewWithDeliverer(
			syncDeliverer,
			cloudlog.WithJob("sync-example-service"),
			cloudlog.WithFormatter(formatter.NewStringFormatter()),
		)
	}
	err := syncLogger.Info("Application started", "version", "1.0.0")
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
	}
	// Logging with additional context
	userLogger := syncLogger.WithContext("user_id", "123456")
	err = userLogger.Info("User logged in", "login_method", "password")
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
	}
	// Logging different levels
	err = syncLogger.Debug("Debug information", "memory_usage", "128MB")
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
		os.Exit(1)
	}
	err = syncLogger.Warn("Resource usage high", "cpu", "80%")
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
		os.Exit(1)
	}
	err = syncLogger.Error("Operation failed", "error", "connection timeout", "retry", true)
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
		os.Exit(1)
	}
	// Good practice to call Flush and Close even for sync loggers
	syncLogger.Flush()
	syncLogger.Close()

	// Console logging example
	consoleClient := &consoleClient{}
	consoleLogger := cloudlog.New(consoleClient,
		cloudlog.WithJob("console-example"),
		cloudlog.WithFormatter(formatter.NewStringFormatter()),
	)
	consoleLogger.Info("Logging to console", "example", true)
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

func (c *consoleClient) Send(_ string, formatted []byte) error {
	fmt.Println(string(formatted))
	return nil
}
