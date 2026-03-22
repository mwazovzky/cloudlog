/*
Package cloudlog provides a structured logging system designed for integration with Grafana Loki.
It features key-value pair logging, context propagation, and flexible formatting options.

# Key Components

1. Logger: Interface that defines logging operations (Info, Error, Debug, Warn)
2. Client: Sends logs to Loki via HTTP push API
3. Formatter: Transforms log entries into bytes (JSON or human-readable string)

# Basic Usage

Create a client, sender, and logger:

	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := cloudlog.NewClient("http://loki-instance/api/v1/push", "username", "token", httpClient)
	sender := cloudlog.NewSyncSender(client)
	logger := cloudlog.New(sender, cloudlog.WithJob("my-service"))

Log a message with key-value pairs:

	ctx := context.Background()
	logger.Info(ctx, "User logged in",
		"user_id", "12345",
		"method", "oauth")

# Metadata

Add persistent metadata to a logger:

	userLogger := logger.With("user_id", "12345", "session_id", "abc123")

	userLogger.Info(ctx, "Profile updated")
	userLogger.Warn(ctx, "Password change attempted")

# Loki Labels

Promote keys to Loki stream labels:

	logger := cloudlog.New(sender,
		cloudlog.WithJob("my-service"),
		cloudlog.WithLabelKeys("request_id", "user_id"),
	)

# Level Filtering

Set minimum log level:

	logger := cloudlog.New(sender,
		cloudlog.WithMinLevel(cloudlog.LevelWarn),
	)

# Error Handling

	err := logger.Info(ctx, "Operation performed")
	if err != nil {
		switch {
		case cloudlog.IsConnectionError(err):
			// Handle connection problem
		case cloudlog.IsFormatError(err):
			// Handle formatting issue
		}
	}
*/
package cloudlog
