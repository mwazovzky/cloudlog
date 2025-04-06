package delivery

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
)

// AsyncDeliverer implements LogDeliverer for asynchronous delivery
type AsyncDeliverer struct {
	sender client.LogSender
	queue  chan LogEntry
	done   chan struct{}
	wg     sync.WaitGroup
	config Config

	// Statistics
	buffered  int64
	delivered uint64
	failed    uint64
	dropped   uint64
	retried   uint64
}

// NewAsyncDeliverer creates a new asynchronous deliverer
func NewAsyncDeliverer(sender client.LogSender, config Config) *AsyncDeliverer {
	if !config.Async {
		config.Async = true
	}

	d := &AsyncDeliverer{
		sender: sender,
		queue:  make(chan LogEntry, config.QueueSize),
		done:   make(chan struct{}),
		config: config,
	}

	// Start worker goroutines
	d.startWorkers()

	return d
}

// Deliver queues a message for delivery
func (d *AsyncDeliverer) Deliver(job string, level string, message string, formatted []byte, timestamp time.Time) error {
	entry := LogEntry{
		Job:       job,
		Level:     level,
		Message:   message,
		Formatted: formatted,
		Timestamp: timestamp,
	}

	// Try to enqueue non-blocking
	select {
	case d.queue <- entry:
		atomic.AddInt64(&d.buffered, 1)
		return nil
	default:
		// Queue full, message dropped
		atomic.AddUint64(&d.dropped, 1)
		return errors.BufferFullError(nil)
	}
}

// startWorkers creates worker goroutines to process log entries
func (d *AsyncDeliverer) startWorkers() {
	for i := 0; i < d.config.Workers; i++ {
		d.wg.Add(1)
		go d.worker()
	}
}

// worker processes log entries from the queue
func (d *AsyncDeliverer) worker() {
	defer d.wg.Done()

	for {
		select {
		case entry := <-d.queue:
			atomic.AddInt64(&d.buffered, -1)
			d.processEntry(entry)
		case <-d.done:
			return
		}
	}
}

// processEntry attempts to deliver a single log entry with retries
func (d *AsyncDeliverer) processEntry(entry LogEntry) {
	var err error

	// Implement retry logic
	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		if attempt > 0 {
			atomic.AddUint64(&d.retried, 1)
			time.Sleep(d.config.RetryInterval)
		}

		err = d.sender.Send(entry.Job, entry.Formatted)
		if err == nil {
			atomic.AddUint64(&d.delivered, 1)
			return
		}
	}

	// Failed after all retries
	atomic.AddUint64(&d.failed, 1)
}

// Flush waits for all queued messages to be processed
func (d *AsyncDeliverer) Flush() error {
	// Create a temporary channel to signal flush completion
	flushDone := make(chan struct{})

	go func() {
		// Wait until queue is empty and all messages are processed
		for {
			// Check current queue size and processing count
			currentBuffered := atomic.LoadInt64(&d.buffered)

			// Wait until queue is completely empty
			if currentBuffered == 0 {
				// Add a small delay to ensure entries being actively processed have time to complete
				time.Sleep(20 * time.Millisecond)

				// Double-check that the buffered count is still 0 (no new entries came in)
				if atomic.LoadInt64(&d.buffered) == 0 {
					close(flushDone)
					return
				}
			}

			// Wait a bit before checking again
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait for flush to complete or timeout
	select {
	case <-flushDone:
		return nil
	case <-time.After(d.config.ShutdownTimeout):
		return errors.TimeoutError(nil)
	}
}

// Close gracefully shuts down the deliverer
func (d *AsyncDeliverer) Close() error {
	close(d.done)

	// Wait for workers to finish with timeout
	c := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		return nil
	case <-time.After(d.config.ShutdownTimeout):
		return errors.ShutdownError(nil)
	}
}

// Status returns current delivery statistics
func (d *AsyncDeliverer) Status() DeliveryStatus {
	return DeliveryStatus{
		Buffered:  int(atomic.LoadInt64(&d.buffered)),
		Delivered: int(atomic.LoadUint64(&d.delivered)),
		Failed:    int(atomic.LoadUint64(&d.failed)),
		Dropped:   int(atomic.LoadUint64(&d.dropped)),
		Retried:   int(atomic.LoadUint64(&d.retried)),
	}
}
