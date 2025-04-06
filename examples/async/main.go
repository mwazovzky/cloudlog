package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mwazovzky/cloudlog"
)

func main() {
	// Get configuration from environment or use defaults
	lokiURL := getEnvOrDefault("LOKI_URL", "")
	lokiUsername := getEnvOrDefault("LOKI_USERNAME", "")
	lokiToken := getEnvOrDefault("LOKI_AUTH_TOKEN", "")
	// Create an HTTP client for Loki
	httpClient := &http.Client{Timeout: 5 * time.Second}
	// Create Loki client
	lokiClient := cloudlog.NewClient(lokiURL, lokiUsername, lokiToken, httpClient)
	// Create an asynchronous logger
	asyncLogger := cloudlog.NewAsync(lokiClient,
		cloudlog.WithJob("basic-async-example"),
		cloudlog.WithMetadata("env", "production"),
	)
	// Log some messages
	for i := 0; i < 20; i++ {
		asyncLogger.Info("Async log message", "index", i)
	}
	// Ensure all logs are delivered before exiting
	asyncLogger.Flush()
	asyncLogger.Close()
	fmt.Println("All logs sent to Loki.")
}

// getEnvOrDefault gets environment variable or returns default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
