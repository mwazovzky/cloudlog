package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mwazovzky/cloudlog"
)

func main() {
	// Create an HTTP client with a reasonable timeout
	httpClient := &http.Client{
		Timeout: time.Second * 5,
	}

	// Create a Loki client
	client := cloudlog.NewClient(
		"http://localhost:3100/loki/api/v1/push",
		"username",
		"token",
		httpClient,
	)

	fmt.Println("=== Basic Usage (with Sync Deliverer) ===")
	fmt.Println("Creating logger with default sync deliverer...")

	// Create a logger using the default constructor which uses
	// a sync deliverer under the hood
	logger := cloudlog.New(client, cloudlog.WithJob("sync-example"))

	// Log some messages
	logger.Info("Application started", "version", "1.0.0")
	logger.Debug("Configuration loaded", "env", "development")
	logger.Warn("Resource usage high", "cpu_percent", 85)
	logger.Error("Database connection failed", "retry", true)

	fmt.Println("Logs sent synchronously")

	// Note: No need to call flush or close for sync logger,
	// but it's good practice to add them for code that might
	// be switched to async later
	logger.Flush()
	logger.Close()
}
