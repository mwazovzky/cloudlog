package main

import (
	"fmt"
	"testing"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/logger" // Add import for logger package
	testlog "github.com/mwazovzky/cloudlog/testing"
)

// This example demonstrates how to test code that uses cloudlog.
//
// In a real application, this would be part of your test suite.
// Here we're just showing the pattern you would use.

func main() {
	// Simulate running a test
	fmt.Println("=== Running Tests ===")
	testVerifyLogMessages(&testing.T{})
	fmt.Println("\nTests completed successfully!")
}

func testVerifyLogMessages(t *testing.T) {
	// Create a test logger
	testLogger := testlog.NewTestLogger()
	logger := cloudlog.New(testLogger)

	fmt.Println("- Testing basic log verification")

	// Function under test
	processOrder(logger, "order-123")

	// Verify logs
	logs := testLogger.Logs()
	fmt.Printf("  Captured %d log messages\n", len(logs))

	// Check if specific messages were logged
	if testLogger.ContainsMessage("Processing order") {
		fmt.Println("  ✓ Found 'Processing order' message")
	} else {
		fmt.Println("  ✗ Missing 'Processing order' message")
	}

	// Check for log entry with specific fields
	if testLogger.ContainsEntry("info", "Processing order", "order_id", "order-123") {
		fmt.Println("  ✓ Found order processing log with correct order_id")
	} else {
		fmt.Println("  ✗ Missing order processing log with correct order_id")
	}

	// Check error logs
	errorLogs := testLogger.LogsOfLevel("error")
	fmt.Printf("  Found %d error logs\n", len(errorLogs))

	// Clear and verify new logs
	testLogger.Clear()
	fmt.Println("- Testing log clearing")
	fmt.Printf("  After Clear(): %d logs\n", len(testLogger.Logs()))

	// Log more messages
	failedOrder(logger, "order-456")

	// Verify error condition
	if testLogger.ContainsEntry("error", "Order processing failed", "order_id", "order-456") {
		fmt.Println("  ✓ Found error log for failed order")
	} else {
		fmt.Println("  ✗ Missing error log for failed order")
	}
}

// Example functions that would be tested

func processOrder(logger *logger.Logger, orderID string) { // Change parameter type
	// Add order context to logs
	orderLogger := logger.WithContext("order_id", orderID)

	orderLogger.Info("Processing order", "timestamp", "2023-07-15T14:22:00Z")
	orderLogger.Debug("Order details validated", "items_count", 3)
	orderLogger.Info("Order processed successfully", "processing_time_ms", 120)
}

func failedOrder(logger *logger.Logger, orderID string) { // Change parameter type
	orderLogger := logger.WithContext("order_id", orderID)

	orderLogger.Warn("Order processing delay", "reason", "high system load")
	orderLogger.Error("Order processing failed", "reason", "payment declined", "retry", false)
}
