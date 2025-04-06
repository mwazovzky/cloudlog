package main

import (
	"fmt"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/delivery"
	"github.com/mwazovzky/cloudlog/formatter"
)

func main() {
	// Create a test client that prints to console
	consoleClient := &ConsoleClient{}

	fmt.Println("\n=== Traditional Usage (Implicit Sync Delivery) ===")

	// Traditional way of creating a logger (uses SyncDeliverer internally)
	traditionalLogger := cloudlog.New(consoleClient,
		cloudlog.WithJob("traditional"),
		cloudlog.WithMetadata("env", "production"),
	)

	traditionalLogger.Info("Hello from traditional logger")

	fmt.Println("\n=== Explicit SyncDeliverer Usage ===")

	// Create a SyncDeliverer explicitly
	syncDeliverer := delivery.NewSyncDeliverer(consoleClient)

	// Create a logger with explicit deliverer
	syncLogger := cloudlog.NewWithDeliverer(syncDeliverer,
		cloudlog.WithJob("explicit-sync"),
		cloudlog.WithFormatter(formatter.NewStringFormatter()),
	)

	syncLogger.Info("Hello from explicit sync logger")

	fmt.Println("\n=== Advanced Usage with Deliverer Chain ===")

	// Example of how you could create a more advanced delivery setup
	// (This is a simplified example - in real code, you might implement
	// a custom deliverer that does preprocessing, filtering, etc.)
	monitoredDeliverer := &MonitoringDeliverer{
		wrapped: syncDeliverer,
	}

	monitoredLogger := cloudlog.NewWithDeliverer(monitoredDeliverer,
		cloudlog.WithJob("monitored-logger"),
	)

	monitoredLogger.Info("Hello from monitored logger")
	fmt.Printf("Message logged with monitoring. Stats: %+v\n", monitoredDeliverer.stats)
}

// ConsoleClient implements client.LogSender for demo purposes
type ConsoleClient struct{}

func (c *ConsoleClient) Send(job string, formatted []byte) error {
	fmt.Printf("[%s] %s\n", job, string(formatted))
	return nil
}

// MonitoringDeliverer wraps another deliverer and adds monitoring
type MonitoringDeliverer struct {
	wrapped delivery.LogDeliverer
	stats   struct {
		InfoCount  int
		ErrorCount int
		WarnCount  int
		DebugCount int
	}
}

func (d *MonitoringDeliverer) Deliver(job, level, message string, formatted []byte, timestamp time.Time) error {
	// Count by level for monitoring
	switch level {
	case "info":
		d.stats.InfoCount++
	case "error":
		d.stats.ErrorCount++
	case "warn":
		d.stats.WarnCount++
	case "debug":
		d.stats.DebugCount++
	}

	// Pass through to the wrapped deliverer
	return d.wrapped.Deliver(job, level, message, formatted, timestamp)
}

func (d *MonitoringDeliverer) Flush() error {
	return d.wrapped.Flush()
}

func (d *MonitoringDeliverer) Close() error {
	return d.wrapped.Close()
}

func (d *MonitoringDeliverer) Status() delivery.DeliveryStatus {
	return d.wrapped.Status()
}
