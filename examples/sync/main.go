package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/formatter"
)

func main() {
	ctx := context.Background()

	lokiURL := getEnvOrDefault("LOKI_URL", "")
	lokiUsername := getEnvOrDefault("LOKI_USERNAME", "")
	lokiToken := getEnvOrDefault("LOKI_AUTH_TOKEN", "")

	var log cloudlog.Logger

	if lokiURL != "" && lokiUsername != "" && lokiToken != "" {
		httpClient := &http.Client{Timeout: 5 * time.Second}
		c := cloudlog.NewClient(lokiURL, lokiUsername, lokiToken, httpClient)
		sender := cloudlog.NewSyncSender(c)
		fmt.Println("Sending logs to Loki instance at:", lokiURL)
		log = cloudlog.New(sender, cloudlog.WithJob("example-service"))
	} else {
		fmt.Println("Loki credentials not provided, logging to console instead")
		sender := &consoleSender{}
		log = cloudlog.New(
			sender,
			cloudlog.WithJob("example-service"),
			cloudlog.WithFormatter(formatter.NewStringFormatter()),
		)
	}

	if err := log.Info(ctx, "Application started", "version", "1.0.0"); err != nil {
		fmt.Println("Error logging:", err)
	}

	userLogger := log.With("user_id", "123456")
	if err := userLogger.Info(ctx, "User logged in", "login_method", "password"); err != nil {
		fmt.Println("Error logging:", err)
	}

	if err := log.Debug(ctx, "Debug information", "memory_usage", "128MB"); err != nil {
		fmt.Println("Error logging:", err)
		os.Exit(1)
	}

	if err := log.Warn(ctx, "Resource usage high", "cpu", "80%"); err != nil {
		fmt.Println("Error logging:", err)
		os.Exit(1)
	}

	if err := log.Error(ctx, "Operation failed", "error", "connection timeout"); err != nil {
		fmt.Println("Error logging:", err)
		os.Exit(1)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// consoleSender implements Sender to print log content to console
type consoleSender struct{}

func (c *consoleSender) Send(_ context.Context, content []byte, labels map[string]string, _ time.Time) error {
	fmt.Printf("[%s] %s\n", labels["job"], string(content))
	return nil
}
