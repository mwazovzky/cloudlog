package main

import (
	"fmt"
	"time"

	"github.com/mwazovzky/cloudlog"
	"github.com/mwazovzky/cloudlog/delivery"
)

func main() {
	// Create a test client that prints to console
	consoleClient := &ConsoleClient{}

	fmt.Println("\n=== Synchronous Logging ===")

	// Synchronous logging example
	syncLogger := cloudlog.New(consoleClient,
		cloudlog.WithJob("sync"),
		cloudlog.WithMetadata("env", "production"),
	)
	syncLogger.Info("Hello from synchronous logger")

	fmt.Println("\n=== Asynchronous Logging ===")

	// Asynchronous logging example
	asyncDeliverer := delivery.NewAsyncDeliverer(consoleClient)
	asyncLogger := cloudlog.NewWithDeliverer(asyncDeliverer,
		cloudlog.WithJob("async"),
	)
	asyncLogger.Info("Hello from asynchronous logger")
	asyncDeliverer.Close() // Ensure all logs are delivered before exiting

	fmt.Println("\n=== Batch Logging ===")

	// Batch logging example
	batchDeliverer := delivery.NewBatchDeliverer(consoleClient, delivery.WithBatchSize(10), delivery.WithFlushInterval(5*time.Second))
	batchLogger := cloudlog.NewWithDeliverer(batchDeliverer,
		cloudlog.WithJob("batch"),
	)
	batchLogger.Info("Hello from batch logger")
	batchDeliverer.Close() // Ensure all logs are delivered before exiting
}

// ConsoleClient implements client.LogSender for demo purposes
type ConsoleClient struct{}

func (c *ConsoleClient) Send(job string, formatted []byte) error {
	fmt.Printf("[%s] %s\n", job, string(formatted))
	return nil
}
