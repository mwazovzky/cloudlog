package logger

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
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
	sendTimeout   time.Duration
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
		sendTimeout:   30 * time.Second,
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

// labelKey returns a string key for grouping entries by their full label set.
func labelKey(labels map[string]string) string {
	// Build a deterministic key from sorted label pairs
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(labels[k])
	}
	return b.String()
}

// sendBatch groups entries by their full label set and sends one LokiEntry per group.
func (s *AsyncSender) sendBatch(batch []entry) {
	if len(batch) == 0 {
		return
	}

	type streamGroup struct {
		labels map[string]string
		values [][]string
	}

	groups := make(map[string]*streamGroup)
	for _, e := range batch {
		key := labelKey(e.labels)
		g, ok := groups[key]
		if !ok {
			g = &streamGroup{labels: e.labels}
			groups[key] = g
		}
		g.values = append(g.values, []string{
			fmt.Sprintf("%d", e.timestamp.UnixNano()),
			string(e.content),
		})
	}

	lokiEntry := client.LokiEntry{
		Streams: make([]client.LokiStream, 0, len(groups)),
	}
	for _, g := range groups {
		lokiEntry.Streams = append(lokiEntry.Streams, client.LokiStream{
			Stream: g.labels,
			Values: g.values,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.sendTimeout)
	defer cancel()

	if err := s.client.Send(ctx, lokiEntry); err != nil {
		s.errorHandler(err)
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

func WithSendTimeout(d time.Duration) AsyncSenderOption {
	return func(s *AsyncSender) {
		if d > 0 {
			s.sendTimeout = d
		}
	}
}
