package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/formatter"
)

func main() {
	ctx := context.Background()

	lokiURL := getEnvOrDefault("LOKI_URL", "")
	lokiUsername := getEnvOrDefault("LOKI_USERNAME", "")
	lokiToken := getEnvOrDefault("LOKI_AUTH_TOKEN", "")

	var syncLogger cloudlog.Logger

	if lokiURL != "" && lokiUsername != "" && lokiToken != "" {
		httpClient := &http.Client{
			Timeout: 5 * time.Second,
		}
		c := cloudlog.NewClient(lokiURL, lokiUsername, lokiToken, httpClient)
		fmt.Println("Sending logs to Loki instance at:", lokiURL)
		syncLogger = cloudlog.NewSync(c, cloudlog.WithJob("sync-example-service"))
	} else {
		fmt.Println("Loki credentials not provided, logging to console instead")
		consoleClient := &consoleClient{}
		syncLogger = cloudlog.NewSync(
			consoleClient,
			cloudlog.WithJob("sync-example-service"),
			cloudlog.WithFormatter(formatter.NewStringFormatter()),
		)
	}

	if err := syncLogger.Info(ctx, "Application started", "version", "1.0.0"); err != nil {
		fmt.Println("Error logging:", err)
	}

	userLogger := syncLogger.With("user_id", "123456")
	if err := userLogger.Info(ctx, "User logged in", "login_method", "password"); err != nil {
		fmt.Println("Error logging:", err)
	}

	if err := syncLogger.Debug(ctx, "Debug information", "memory_usage", "128MB"); err != nil {
		fmt.Println("Error logging:", err)
		os.Exit(1)
	}

	if err := syncLogger.Warn(ctx, "Resource usage high", "cpu", "80%"); err != nil {
		fmt.Println("Error logging:", err)
		os.Exit(1)
	}

	if err := syncLogger.Error(ctx, "Operation failed", "error", "connection timeout"); err != nil {
		fmt.Println("Error logging:", err)
		os.Exit(1)
	}

	// Console logging example
	consoleClient := &consoleClient{}
	consoleLogger := cloudlog.NewSync(consoleClient,
		cloudlog.WithJob("console-example"),
		cloudlog.WithFormatter(formatter.NewStringFormatter()),
	)
	consoleLogger.Info(ctx, "Logging to console", "example", true)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// consoleClient implements the LogSender interface to print logs to console
type consoleClient struct{}

func (c *consoleClient) Send(_ context.Context, entry client.LokiEntry) error {
	for _, stream := range entry.Streams {
		for _, value := range stream.Values {
			if len(value) >= 2 {
				fmt.Printf("[%s] %s\n", stream.Stream["job"], value[1])
			}
		}
	}
	return nil
}
