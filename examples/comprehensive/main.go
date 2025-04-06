// Example demonstrating comprehensive usage of CloudLog
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/mwazovzky/cloudlog/testing"
)

func main() {
	fmt.Println("CloudLog Comprehensive Example")
	fmt.Println("==============================")

	// PART 1: Basic Usage
	fmt.Println("\n1. Basic Usage")
	fmt.Println("-------------")
	DemonstrateBasicUsage()

	// PART 2: Advanced Formatting
	fmt.Println("\n\n2. Advanced Formatting")
	fmt.Println("---------------------")
	DemonstrateFormatting()

	// PART 3: Context and Job Handling
	fmt.Println("\n\n3. Context and Jobs")
	fmt.Println("-----------------")
	DemonstrateContext()

	// PART 4: Error Handling
	fmt.Println("\n\n4. Error Handling")
	fmt.Println("---------------")
	DemonstrateErrorHandling()

	// PART 5: Testing
	fmt.Println("\n\n5. Testing Utilities")
	fmt.Println("------------------")
	DemonstrateTesting()

	// PART 6: Direct Package Usage
	fmt.Println("\n\n6. Direct Package Usage")
	fmt.Println("--------------------")
	DemonstrateDirectPackageUsage()
}

// DemonstrateBasicUsage shows the basic logger setup and methods
func DemonstrateBasicUsage() {
	// Create a console client for demonstration purposes
	consoleClient := &ConsoleClient{}

	// Create a basic logger
	logger := cloudlog.New(consoleClient)

	// Log at different levels
	logger.Info("Application starting", "version", "1.0.0")
	logger.Debug("Configuration loaded", "env", "development")
	logger.Warn("High memory usage", "memory_mb", 1024, "threshold_mb", 1000)
	logger.Error("Database connection failed", "error", "connection refused", "retry", true)
}

// DemonstrateFormatting shows different formatting options
func DemonstrateFormatting() {
	consoleClient := &ConsoleClient{}

	// 1. Default JSON formatter
	fmt.Println("Default JSON formatter:")
	defaultLogger := cloudlog.New(consoleClient)
	defaultLogger.Info("User signed in", "user_id", 12345)

	// 2. Custom JSON formatter
	fmt.Println("\nCustom JSON formatter:")
	jsonFormatter := formatter.NewJSONFormatter(
		formatter.WithTimeFormat(time.RFC1123),
		formatter.WithTimestampField("@timestamp"),
		formatter.WithLevelField("severity"),
	)
	jsonLogger := cloudlog.New(consoleClient, cloudlog.WithFormatter(jsonFormatter))
	jsonLogger.Info("API request processed", "method", "GET", "path", "/api/v1/users", "duration_ms", 42)

	// 3. String formatter
	fmt.Println("\nString formatter:")
	stringFormatter := formatter.NewStringFormatter()
	stringLogger := cloudlog.New(consoleClient, cloudlog.WithFormatter(stringFormatter))
	stringLogger.Info("Cache updated", "items", 150, "duration_ms", 25)

	// 4. Custom string formatter
	fmt.Println("\nCustom string formatter:")
	customStringFormatter := formatter.NewStringFormatter(
		formatter.WithStringTimeFormat("2006/01/02 15:04:05"),
		formatter.WithKeyValueSeparator(": "),
		formatter.WithPairSeparator(" | "),
	)
	customStringLogger := cloudlog.New(consoleClient, cloudlog.WithFormatter(customStringFormatter))
	customStringLogger.Info("Payment processed", "amount", 99.95, "currency", "USD", "status", "approved")
}

// DemonstrateContext shows context and job name handling
func DemonstrateContext() {
	consoleClient := &ConsoleClient{}

	// Create base logger
	baseLogger := cloudlog.New(consoleClient, cloudlog.WithJob("my-service"))
	baseLogger.Info("Service started", "environment", "production")

	// Add request context
	requestLogger := baseLogger.WithContext("request_id", "req-123", "client_ip", "192.168.1.1")
	requestLogger.Info("Request received", "method", "POST", "path", "/api/orders")

	// Add user context to request context
	userLogger := requestLogger.WithContext("user_id", "user-456", "role", "admin")
	userLogger.Info("User authenticated")
	userLogger.Debug("Processing user permissions")

	// Show that context is properly nested
	userLogger.Info("Creating order", "order_id", "ord-789", "amount", 125.50)

	// Change job name
	dbLogger := baseLogger.WithJob("database-service")
	dbLogger.Info("Database connection established")
	dbLogger.Debug("Executing query", "table", "users", "operation", "SELECT")

	// Combine new job and context
	dbAdminLogger := dbLogger.WithContext("admin_operation", true)
	dbAdminLogger.Warn("Running administrative query", "impact", "high")
}

// DemonstrateErrorHandling shows error handling
func DemonstrateErrorHandling() {
	// Create a client that will fail
	failingClient := &FailingClient{FailWith: "connection"}
	logger := cloudlog.New(failingClient)

	// Try to log a message
	fmt.Println("Attempting to log with a failing client...")
	err := logger.Info("This will fail", "reason", "client is configured to fail")

	if err != nil {
		fmt.Printf("Got error: %v\n", err)

		// Demonstrate error type checking
		if cloudlog.IsConnectionError(err) {
			fmt.Println("✓ Correctly identified as a connection error")
		} else {
			fmt.Println("✗ Failed to identify as a connection error")
		}
	}

	// Test response error
	failingClient.FailWith = "response"
	err = logger.Info("This will fail with response error")

	if err != nil {
		fmt.Printf("Got error: %v\n", err)

		if cloudlog.IsResponseError(err) {
			fmt.Println("✓ Correctly identified as a response error")
		} else {
			fmt.Println("✗ Failed to identify as a response error")
		}
	}
}

// DemonstrateTesting shows how to use testing utilities
func DemonstrateTesting() {
	// Create a test logger
	testLogger := testing.NewTestLogger()
	logger := cloudlog.New(testLogger)

	// Log some messages
	logger.Info("User login", "user_id", "user-123")
	logger.Debug("Processing request", "request_id", "req-456")
	logger.Error("Payment failed", "order_id", "ord-789", "reason", "insufficient funds")

	// Demonstrate finding logs
	logs := testLogger.Logs()
	fmt.Printf("Captured %d logs\n", len(logs))

	// Find specific log types
	errorLogs := testLogger.LogsOfLevel("error")
	fmt.Printf("Found %d error logs\n", len(errorLogs))

	// Check for specific messages
	if testLogger.ContainsMessage("User login") {
		fmt.Println("✓ Found 'User login' message")
	}

	// Check for specific entries
	if testLogger.ContainsEntry("error", "Payment failed", "reason", "insufficient funds") {
		fmt.Println("✓ Found payment failure with correct reason")
	}

	// Clear logs and verify
	testLogger.Clear()
	fmt.Printf("After clearing: %d logs\n", len(testLogger.Logs()))
}

// DemonstrateDirectPackageUsage shows how to use packages directly
func DemonstrateDirectPackageUsage() {
	// Create components directly
	httpClient := &http.Client{Timeout: 5 * time.Second}
	lokiClient := client.NewLokiClient(
		"http://example-loki.com/api/v1/push",
		"username",
		"token",
		httpClient,
	)

	// Use custom formatter
	jsonFormatter := formatter.NewJSONFormatter(
		formatter.WithTimeFormat(time.RFC3339Nano),
		formatter.WithTimestampField("timestamp"),
	)

	// Create logger directly from components
	logger := cloudlog.New(
		lokiClient,
		cloudlog.WithJob("direct-example"),
		cloudlog.WithFormatter(jsonFormatter),
	)

	// Just print information about the created components
	fmt.Printf("Created Loki client: %T\n", lokiClient)
	fmt.Printf("Created JSON formatter: %T\n", jsonFormatter)
	fmt.Printf("Created logger with direct components: %T\n", logger)
}

// ConsoleClient is a simple implementation of client.LogSender that prints to console
type ConsoleClient struct{}

func (c *ConsoleClient) Send(job string, formatted []byte) error {
	fmt.Printf("[%s] %s\n", job, string(formatted))
	return nil
}

// FailingClient is a client that fails in different ways for demonstration
type FailingClient struct {
	FailWith string
}

func (c *FailingClient) Send(_ string, _ []byte) error {
	switch c.FailWith {
	case "connection":
		return errors.ConnectionError(fmt.Errorf("network error"), "failed to connect")
	case "response":
		return errors.ResponseError(500, "internal server error")
	default:
		return fmt.Errorf("unknown error")
	}
}
