package logger

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
)

// entry represents a log entry waiting to be sent, or a flush marker.
type entry struct {
	content   []byte
	labels    map[string]string
	timestamp time.Time
	flushCh   chan struct{} // non-nil for flush markers
}

// AsyncSender implements Sender with non-blocking, buffered delivery.
// A background worker batches entries and sends them to the underlying LogSender.
type AsyncSender struct {
	client        client.LogSender
	buffer        chan entry
	batchSize     int
	flushInterval time.Duration
	blockOnFull   bool
	done          chan struct{}
	wg            sync.WaitGroup
	errorHandler  func(error)
	closed        bool
	mu            sync.Mutex
}

// AsyncSenderOption configures an AsyncSender.
type AsyncSenderOption func(*AsyncSender)

// NewAsyncSender creates a buffered sender that delivers log entries asynchronously.
func NewAsyncSender(client client.LogSender, options ...AsyncSenderOption) *AsyncSender {
	s := &AsyncSender{
		client:        client,
		buffer:        make(chan entry, 1000),
		batchSize:     100,
		flushInterval: 5 * time.Second,
		blockOnFull:   false,
		done:          make(chan struct{}),
		errorHandler:  func(err error) { log.Printf("cloudlog: send error: %v", err) },
	}

	for _, opt := range options {
		opt(s)
	}

	s.wg.Add(1)
	go s.worker()

	return s
}

// Send buffers an entry for async delivery. Non-blocking by default.
func (s *AsyncSender) Send(_ context.Context, content []byte, labels map[string]string, timestamp time.Time) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("%w: sender is closed", errors.ErrBufferFull)
	}
	s.mu.Unlock()

	e := entry{
		content:   content,
		labels:    labels,
		timestamp: timestamp,
	}

	if s.blockOnFull {
		select {
		case s.buffer <- e:
			return nil
		case <-s.done:
			return fmt.Errorf("%w: sender is closed", errors.ErrBufferFull)
		}
	}

	select {
	case s.buffer <- e:
		return nil
	default:
		return fmt.Errorf("%w", errors.ErrBufferFull)
	}
}

// Flush blocks until all buffered entries have been sent.
func (s *AsyncSender) Flush() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	flushCh := make(chan struct{})
	s.buffer <- entry{flushCh: flushCh}
	<-flushCh
}

// Close flushes remaining entries and stops the background worker.
func (s *AsyncSender) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	s.Flush()
	close(s.done)
	s.wg.Wait()
}

// worker is the background goroutine that pulls entries, batches them, and sends.
func (s *AsyncSender) worker() {
	defer s.wg.Done()

	batch := make([]entry, 0, s.batchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case e := <-s.buffer:
			if e.flushCh != nil {
				s.sendBatch(batch)
				batch = batch[:0]
				close(e.flushCh)
				continue
			}

			batch = append(batch, e)
			if len(batch) >= s.batchSize {
				s.sendBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.sendBatch(batch)
				batch = batch[:0]
			}

		case <-s.done:
			// Drain remaining entries from buffer
			for {
				select {
				case e := <-s.buffer:
					if e.flushCh != nil {
						close(e.flushCh)
						continue
					}
					batch = append(batch, e)
				default:
					if len(batch) > 0 {
						s.sendBatch(batch)
					}
					return
				}
			}
		}
	}
}

// sendBatch groups entries by job label and sends one LokiEntry per job.
func (s *AsyncSender) sendBatch(batch []entry) {
	if len(batch) == 0 {
		return
	}

	// Group entries by job label
	byJob := make(map[string][]entry)
	for _, e := range batch {
		job := e.labels["job"]
		byJob[job] = append(byJob[job], e)
	}

	for job, entries := range byJob {
		values := make([][]string, 0, len(entries))
		for _, e := range entries {
			values = append(values, []string{
				fmt.Sprintf("%d", e.timestamp.UnixNano()),
				string(e.content),
			})
		}

		// Use labels from first entry, ensure job is set
		labels := make(map[string]string)
		for k, v := range entries[0].labels {
			labels[k] = v
		}
		labels["job"] = job

		lokiEntry := client.LokiEntry{
			Streams: []client.LokiStream{
				{
					Stream: labels,
					Values: values,
				},
			},
		}

		if err := s.client.Send(context.Background(), lokiEntry); err != nil {
			s.errorHandler(err)
		}
	}
}

// Option constructors for AsyncSender

func WithBufferSize(size int) AsyncSenderOption {
	return func(s *AsyncSender) {
		if size > 0 {
			s.buffer = make(chan entry, size)
		}
	}
}

func WithBatchSize(size int) AsyncSenderOption {
	return func(s *AsyncSender) {
		if size > 0 {
			s.batchSize = size
		}
	}
}

func WithFlushInterval(d time.Duration) AsyncSenderOption {
	return func(s *AsyncSender) {
		if d > 0 {
			s.flushInterval = d
		}
	}
}

func WithBlockOnFull(block bool) AsyncSenderOption {
	return func(s *AsyncSender) {
		s.blockOnFull = block
	}
}

func WithErrorHandler(handler func(error)) AsyncSenderOption {
	return func(s *AsyncSender) {
		if handler != nil {
			s.errorHandler = handler
		}
	}
}
