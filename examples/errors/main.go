package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mwazovzky/cloudlog"
)

// This example demonstrates how to handle different types of errors
// that can occur when using the cloudlog library.
//
// Run: go run main.go

func main() {
	// Create a client with an invalid URL to demonstrate connection errors
	badClient := cloudlog.NewClient("http://nonexistent-loki-server", "user", "token", &http.Client{
		Timeout: 2 * time.Second,
	})
	logger := cloudlog.New(badClient)

	// Try to log a message
	fmt.Println("Attempting to log to a non-existent server...")
	err := logger.Info("This will fail", "reason", "server doesn't exist")

	if err != nil {
		fmt.Printf("Error encountered: %v\n\n", err)

		// Demonstrate error type checking
		if cloudlog.IsConnectionError(err) {
			fmt.Println("✓ Correctly identified as a connection error")
		} else {
			fmt.Println("✗ Failed to identify as a connection error")
		}
	}

	// Create a custom type that will cause JSON marshaling to fail
	type BadType struct {
		Channel chan int // Channels can't be JSON marshaled
	}

	// Try to log a value that can't be marshaled to JSON
	fmt.Println("\nAttempting to log a value that can't be marshaled to JSON...")
	badValue := BadType{Channel: make(chan int)}

	err = logger.Info("This will fail", "bad_value", badValue)

	if err != nil {
		fmt.Printf("Error encountered: %v\n\n", err)

		// Demonstrate error type checking
		if cloudlog.IsFormatError(err) {
			fmt.Println("✓ Correctly identified as a format error")
		} else {
			fmt.Println("✗ Failed to identify as a format error")
		}
	}

	fmt.Println("\nExample completed.")
}
