// Package delivery provides mechanisms for delivering log entries to backends.
// It includes both synchronous and asynchronous delivery strategies.
package delivery

import (
	"time"
)

// LogEntry represents a log entry to be delivered
type LogEntry struct {
	Job       string
	Level     string
	Message   string
	Formatted []byte
	Timestamp time.Time
}

// LogDeliverer defines the interface for any log delivery mechanism
type LogDeliverer interface {
	// Deliver attempts to deliver a log entry
	// For synchronous deliverers, this blocks until delivery completes
	// For asynchronous deliverers, this may queue the message and return immediately
	Deliver(job string, level string, message string, formatted []byte, timestamp time.Time) error

	// Flush forces delivery of any buffered messages
	Flush() error

	// Close gracefully shuts down the deliverer
	Close() error

	// Status returns the current delivery statistics
	Status() DeliveryStatus
}

// DeliveryStatus contains statistics about the delivery service
type DeliveryStatus struct {
	Buffered  int
	Delivered int
	Failed    int
	Dropped   int
	Retried   int
}

// Config defines configuration options for delivery services
type Config struct {
	Async           bool
	QueueSize       int
	BatchSize       int
	Workers         int
	FlushInterval   time.Duration
	MaxRetries      int
	RetryInterval   time.Duration
	ShutdownTimeout time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Async:           false,
		QueueSize:       1000,
		BatchSize:       100,
		Workers:         1,
		FlushInterval:   time.Second * 5,
		MaxRetries:      3,
		RetryInterval:   time.Second * 1,
		ShutdownTimeout: time.Second * 10,
	}
}
