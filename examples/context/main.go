package main

import (
	"fmt"

	"github.com/mwazovzky/cloudlog"
	// Add import for logger package
)

// This example demonstrates how to use context and job names
// to organize and structure your logs.
//
// Run: go run main.go

func main() {
	// Create a console client for easy visualization
	consoleClient := &consoleClient{}

	// Create a base logger
	baseLogger := cloudlog.New(consoleClient, cloudlog.WithJob("app"))

	fmt.Println("=== Basic Logging ===")
	baseLogger.Info("Application started", "version", "1.0.0")

	fmt.Println("\n=== Logging with Context ===")
	// Add request context
	requestLogger := baseLogger.WithContext(
		"request_id", "req-abc-123",
		"client_ip", "192.168.1.1",
	)

	requestLogger.Info("Request received", "method", "GET", "path", "/api/users")
	requestLogger.Debug("Processing request parameters", "limit", 20, "offset", 0)

	// Add user context to the request context
	userLogger := requestLogger.WithContext("user_id", "user-456", "role", "admin")
	userLogger.Info("User authenticated")

	// Log an error with all the context
	userLogger.Error("Permission denied", "resource", "reports", "action", "download")

	fmt.Println("\n=== Logging with Different Job Names ===")
	// Create loggers for different components
	apiLogger := baseLogger.WithJob("api-service")
	dbLogger := baseLogger.WithJob("db-service")
	authLogger := baseLogger.WithJob("auth-service")

	apiLogger.Info("API server started", "port", 8080)
	dbLogger.Info("Database connection established", "host", "db.example.com")
	authLogger.Info("Auth provider initialized", "provider", "oauth2")

	// Combining job names and context
	dbRequestLogger := dbLogger.WithContext("request_id", "req-abc-123")
	dbRequestLogger.Debug("Executing database query", "table", "users", "operation", "SELECT")
}

// consoleClient implements the LogSender interface to print logs to console
type consoleClient struct{}

func (c *consoleClient) Send(job string, formatted []byte) error {
	fmt.Printf("[%s] %s\n", job, string(formatted))
	return nil
}
