package delivery

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
)

// BatchDeliverer implements LogDeliverer with batch processing
// It's a more advanced version of the AsyncDeliverer that batches logs
// together before sending them to the backend
type BatchDeliverer struct {
	sender       client.LogSender
	queue        chan LogEntry
	batchTicker  *time.Ticker
	done         chan struct{}
	wg           sync.WaitGroup
	config       Config
	currentBatch []LogEntry

	// Statistics
	buffered  int64
	delivered uint64
	failed    uint64
	dropped   uint64
	retried   uint64
	batches   uint64

	mu sync.Mutex
}

// NewBatchDeliverer creates a new batch deliverer
func NewBatchDeliverer(sender client.LogSender, config Config) *BatchDeliverer {
	if !config.Async {
		config.Async = true
	}

	d := &BatchDeliverer{
		sender:       sender,
		queue:        make(chan LogEntry, config.QueueSize),
		done:         make(chan struct{}),
		config:       config,
		currentBatch: make([]LogEntry, 0, config.BatchSize),
	}

	// Start processors
	d.startProcessing()

	return d
}

// Deliver queues a message for batch delivery
func (d *BatchDeliverer) Deliver(job string, level string, message string, formatted []byte, timestamp time.Time) error {
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

// startProcessing starts the batch processing routine
func (d *BatchDeliverer) startProcessing() {
	// Start a ticker for regular flushing
	d.batchTicker = time.NewTicker(d.config.FlushInterval)

	// Start the batch processor
	d.wg.Add(1)
	go d.batchProcessor()
}

// batchProcessor manages the batching and sending of log entries
func (d *BatchDeliverer) batchProcessor() {
	defer d.wg.Done()

	for {
		select {
		case entry := <-d.queue:
			// Got a log entry, add to batch
			atomic.AddInt64(&d.buffered, -1)
			d.mu.Lock()
			d.currentBatch = append(d.currentBatch, entry)

			// If batch is full, send it
			if len(d.currentBatch) >= d.config.BatchSize {
				d.sendCurrentBatchLocked()
			}
			d.mu.Unlock()

		case <-d.batchTicker.C:
			// Time to flush the batch
			d.mu.Lock()
			if len(d.currentBatch) > 0 {
				d.sendCurrentBatchLocked()
			}
			d.mu.Unlock()

		case <-d.done:
			// Exit signal received
			d.batchTicker.Stop()

			// Final flush
			d.mu.Lock()
			if len(d.currentBatch) > 0 {
				d.sendCurrentBatchLocked()
			}
			d.mu.Unlock()
			return
		}
	}
}

// sendCurrentBatchLocked sends the current batch
// Caller must hold the mutex
func (d *BatchDeliverer) sendCurrentBatchLocked() {
	// Get the current batch and create a new one
	batch := d.currentBatch
	d.currentBatch = make([]LogEntry, 0, d.config.BatchSize)

	// Send the batch in a separate goroutine to not block the processor
	go d.sendBatch(batch)
}

// sendBatch sends a batch of log entries with retries
func (d *BatchDeliverer) sendBatch(batch []LogEntry) {
	if len(batch) == 0 {
		return
	}

	atomic.AddUint64(&d.batches, 1)
	success := 0

	// Process each entry in the batch
	for _, entry := range batch {
		var err error

		// Try with retries
		for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
			if attempt > 0 {
				atomic.AddUint64(&d.retried, 1)
				time.Sleep(d.config.RetryInterval)
			}

			err = d.sender.Send(entry.Job, entry.Formatted)
			if err == nil {
				success++
				break
			}
		}

		// Count failures
		if err != nil {
			atomic.AddUint64(&d.failed, 1)
		}
	}

	// Count successful deliveries
	atomic.AddUint64(&d.delivered, uint64(success))
}

// Flush waits for all queued messages to be processed
func (d *BatchDeliverer) Flush() error {
	// Create a temporary channel to signal flush completion
	flushDone := make(chan struct{})

	go func() {
		// First wait until queue is empty
		for {
			if atomic.LoadInt64(&d.buffered) == 0 {
				// Queue is empty, but ensure current batch is flushed
				d.mu.Lock()
				batchEmpty := len(d.currentBatch) == 0
				d.mu.Unlock()

				if batchEmpty {
					close(flushDone)
					return
				}

				// Force flush the current batch
				d.mu.Lock()
				if len(d.currentBatch) > 0 {
					d.sendCurrentBatchLocked()
				}
				d.mu.Unlock()

				// Give it a moment to process
				time.Sleep(50 * time.Millisecond)
			} else {
				// Still have messages in queue
				time.Sleep(10 * time.Millisecond)
			}
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
func (d *BatchDeliverer) Close() error {
	// First flush all pending messages
	flushErr := d.Flush()

	// Signal processor to stop
	close(d.done)

	// Wait for the processor to finish with timeout
	c := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(c)
	}()

	var closeErr error
	select {
	case <-c:
		closeErr = nil
	case <-time.After(d.config.ShutdownTimeout):
		closeErr = errors.ShutdownError(nil)
	}

	// Return flush error if that failed, otherwise return close error
	if flushErr != nil {
		return flushErr
	}
	return closeErr
}

// Status returns current delivery statistics
func (d *BatchDeliverer) Status() DeliveryStatus {
	return DeliveryStatus{
		Buffered:  int(atomic.LoadInt64(&d.buffered)),
		Delivered: int(atomic.LoadUint64(&d.delivered)),
		Failed:    int(atomic.LoadUint64(&d.failed)),
		Dropped:   int(atomic.LoadUint64(&d.dropped)),
		Retried:   int(atomic.LoadUint64(&d.retried)),
	}
}
