package main

import (
	"fmt"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/formatter"
	// Add import for logger package
)

// This example demonstrates different formatters available in CloudLog.
//
// Run: go run main.go

func main() {
	// Create a console client for easy visualization
	consoleClient := &consoleClient{}

	fmt.Println("=== JSON Formatter (Default) ===")
	jsonLogger := cloudlog.New(consoleClient)

	// Basic JSON logging
	jsonLogger.Info("User profile updated",
		"user_id", 12345,
		"fields_changed", []string{"name", "email"},
		"timestamp", time.Now().Unix())

	fmt.Println("\n=== JSON Formatter with Custom Options ===")
	customJSON := formatter.NewJSONFormatter(
		formatter.WithTimeFormat(time.RFC1123),
		formatter.WithTimestampField("@timestamp"),
		formatter.WithLevelField("severity"),
		formatter.WithJobField("service"),
	)

	customJSONLogger := cloudlog.New(consoleClient, cloudlog.WithFormatter(customJSON))
	customJSONLogger.Info("API request received", "method", "GET", "path", "/users")

	fmt.Println("\n=== String Formatter ===")
	stringLogger := cloudlog.New(consoleClient,
		cloudlog.WithFormatter(formatter.NewStringFormatter()))

	stringLogger.Info("Database connection established",
		"db", "users",
		"pool_size", 10)

	fmt.Println("\n=== String Formatter with Custom Options ===")
	customString := formatter.NewStringFormatter(
		formatter.WithStringTimeFormat("2006/01/02 15:04:05"),
		formatter.WithKeyValueSeparator(": "),
		formatter.WithPairSeparator(" | "),
	)

	customStringLogger := cloudlog.New(consoleClient, cloudlog.WithFormatter(customString))
	customStringLogger.Warn("Memory usage high", "usage_mb", 1024, "threshold_mb", 1000)
}

// consoleClient implements the LogSender interface to print logs to console
type consoleClient struct{}

func (c *consoleClient) Send(_ string, formatted []byte) error {
	fmt.Println(string(formatted))
	return nil
}
