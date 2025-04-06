/*
Package delivery provides mechanisms for delivering log entries to logging backends.

The delivery package offers two main delivery strategies:

 1. Synchronous delivery (SyncDeliverer): Each log entry is immediately sent to the
    backend, and the calling code blocks until delivery completes or fails. This is
    the simplest approach and provides immediate feedback, but can impact application
    performance if the logging backend is slow or unreliable.

 2. Asynchronous delivery (AsyncDeliverer): Log entries are queued for delivery in
    a background goroutine. The calling code can continue execution immediately.
    This improves application performance but requires proper shutdown handling to
    ensure logs are not lost.

Both deliverers implement the LogDeliverer interface, allowing them to be used
interchangeably in code that depends on log delivery.

Example usage of synchronous delivery:

	// Create a sender
	sender := client.NewLokiClient("http://loki:3100", "user", "token", httpClient)

	// Create a synchronous deliverer
	deliverer := delivery.NewSyncDeliverer(sender)

	// Deliver a log entry
	err := deliverer.Deliver(
		"my-service",
		"info",
		"Application started",
		[]byte(`{"level":"info","message":"Application started"}`),
		time.Now(),
	)

For asynchronous delivery, see the AsyncDeliverer documentation.
*/
package delivery
