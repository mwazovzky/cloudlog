// Package client provides implementations for sending logs to different backends.
// The primary implementation is LokiClient which sends logs to a Grafana Loki instance.
package client

// LogSender defines methods that any log client implementation should provide.
// It represents a component capable of sending formatted logs to a backend service.
type LogSender interface {
	// Send transmits formatted log data with the given job name to a logging backend.
	// Parameters:
	//   - job: Identifies the source or category of the log (e.g., "api-service")
	//   - formatted: The formatted log entry as bytes, typically JSON or string
	// Returns an error if the sending operation fails.
	Send(job string, formatted []byte) error
}
