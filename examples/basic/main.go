package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mwazovzky/cloudlog"
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

	var logger *cloudlog.Logger

	if lokiURL != "" && lokiUsername != "" && lokiToken != "" {
		// Create a new HTTP client with reasonable timeout
		httpClient := &http.Client{
			Timeout: 5 * time.Second,
		}

		// Create Loki client and logger
		client := cloudlog.NewClient(lokiURL, lokiUsername, lokiToken, httpClient)
		logger = cloudlog.New(client, cloudlog.WithJob("example-service"))
		fmt.Println("Sending logs to Loki instance at:", lokiURL)
	} else {
		// Create console output logger if Loki credentials aren't provided
		fmt.Println("Loki credentials not provided, logging to console instead")
		logger = createConsoleLogger()
	}

	// Basic logging
	err := logger.Info("Application started", "version", "1.0.0")
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
	}

	// Logging with additional context
	userLogger := logger.WithContext("user_id", "123456")
	err = userLogger.Info("User logged in", "login_method", "password")
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
	}

	// Logging different levels
	err = logger.Debug("Debug information", "memory_usage", "128MB")
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
		os.Exit(1)
	}

	err = logger.Warn("Resource usage high", "cpu", "80%")
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
		os.Exit(1)
	}

	err = logger.Error("Operation failed", "error", "connection timeout", "retry", true)
	if err != nil {
		fmt.Println("Error logging to Loki:", err)
		os.Exit(1)
	}

	fmt.Println("Example completed successfully!")
}

// getEnvOrDefault gets environment variable or returns default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// createConsoleLogger creates a logger that prints to console
func createConsoleLogger() *cloudlog.Logger {
	return cloudlog.New(&consoleClient{},
		cloudlog.WithFormatter(formatter.NewStringFormatter()))
}

// consoleClient implements the LogSender interface to print logs to console
type consoleClient struct{}

func (c *consoleClient) Send(_ string, formatted []byte) error {
	fmt.Println(string(formatted))
	return nil
}
