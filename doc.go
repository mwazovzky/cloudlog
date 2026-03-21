/*
Package cloudlog provides a structured logging system designed for integration with Grafana Loki
and other logging backends. It features key-value pair logging, context propagation, and
flexible formatting options.

# Key Components

1. Logger: Interface that defines logging operations (Info, Error, Debug, Warn)
2. Client: Implementation for sending logs to backends like Loki
3. Formatter: Transforms log entries into proper format (JSON, string, Loki protocol)
4. SyncLogger: Synchronous implementation that blocks until logs are sent

# Basic Usage

Create a client and synchronous logger:

	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := cloudlog.NewClient("http://loki-instance/api/v1/push", "username", "token", httpClient)
	logger := cloudlog.NewSync(client, cloudlog.WithJob("my-service"))

Log a message with key-value pairs:

	logger.Info("User logged in",
		"user_id", "12345",
		"method", "oauth",
		"ip", "192.168.1.1")

# Context Propagation

Add persistent context to a logger:

	// Create a context-specific logger
	userLogger := logger.WithContext("user_id", "12345", "session_id", "abc123")

	// All logs from this logger will include the context
	userLogger.Info("Profile updated")
	userLogger.Warn("Password change attempted")

# Formatting Options

Configure formatting:

	// Loki formatter with custom label keys
	lokiLogger := cloudlog.NewSync(
		client,
		cloudlog.WithFormatter(cloudlog.NewLokiFormatter(
			cloudlog.WithLabelKeys("request_id", "user_id"),
		)),
	)

# Error Handling

Check and handle specific error types:

	err := logger.Info("Operation performed", "status", "success")
	if err != nil {
		switch {
		case cloudlog.IsConnectionError(err):
			// Handle connection problem
		case cloudlog.IsFormatError(err):
			// Handle formatting issue
		default:
			// Handle other errors
		}
	}

# Graceful Shutdown

Ensure all logs are processed before exiting:

	logger.Close()
*/
package cloudlog
