package logger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
)

// Constants for default configuration values
const (
	// DefaultMaxRequestSize is the default maximum size for HTTP requests (5MB)
	DefaultMaxRequestSize = 5 * 1024 * 1024

	// DefaultBufferSize is the default size of the entry buffer
	DefaultBufferSize = 1000

	// DefaultBatchSize is the default number of entries to include in a batch
	DefaultBatchSize = 100

	// DefaultFlushInterval is the default interval between forced flushes
	DefaultFlushInterval = 5 * time.Second

	// DefaultWorkerCount is the default number of worker goroutines
	DefaultWorkerCount = 2

	// EstimatedEntryOverhead is the estimated overhead per entry in bytes
	EstimatedEntryOverhead = 100

	// FlushMarkerTimeout is the timeout for flush operations
	FlushMarkerTimeout = 2 * time.Second
)

// AsyncLogger implements the Logger interface with asynchronous, non-blocking behavior
type AsyncLogger struct {
	formatter      formatter.Formatter
	job            string
	metadata       map[string]interface{}
	sender         client.LogSender
	buffer         chan asyncLogEntry
	batchSize      int
	flushTicker    *time.Ticker
	workers        int
	blockOnFull    bool
	maxRequestSize int // Maximum size of HTTP requests in bytes
	wg             sync.WaitGroup
	done           chan struct{}
	closed         bool
	mu             sync.RWMutex
	errorHandler   errors.ErrorHandler
}

// asyncLogEntry represents a log entry waiting to be processed
type asyncLogEntry struct {
	job      string
	level    string
	message  string
	keyvals  []interface{}
	metadata map[string]interface{}
}

// AsyncOption defines a configuration option for AsyncLogger
type AsyncOption func(*AsyncLogger)

// NewAsync creates a new Logger instance with asynchronous behavior
func NewAsync(client client.LogSender, options ...AsyncOption) Logger {
	// Default configuration
	logger := &AsyncLogger{
		formatter:      formatter.NewLokiFormatter(),
		job:            "application",
		metadata:       make(map[string]interface{}),
		sender:         client,
		buffer:         make(chan asyncLogEntry, DefaultBufferSize), // Default buffer size 1000
		batchSize:      DefaultBatchSize,                            // Default batch size 100
		flushTicker:    time.NewTicker(DefaultFlushInterval),        // Default flush interval 5s
		workers:        DefaultWorkerCount,                          // Default 2 workers
		blockOnFull:    false,                                       // Default to non-blocking behavior
		maxRequestSize: DefaultMaxRequestSize,                       // Default 5MB size limit
		done:           make(chan struct{}),
		errorHandler:   errors.NoopErrorHandler, // Use the NoopErrorHandler from errors package
	}

	// Apply options
	for _, option := range options {
		option(logger)
	}

	// Start workers
	logger.start()

	return logger
}

// WithBufferSize sets the maximum number of log entries in the buffer
func WithBufferSize(size int) AsyncOption {
	return func(l *AsyncLogger) {
		if size > 0 {
			// Create a new channel with the specified size
			l.buffer = make(chan asyncLogEntry, size)
		}
	}
}

// WithBatchSize sets how many entries to include in each batch
func WithBatchSize(size int) AsyncOption {
	return func(l *AsyncLogger) {
		if size > 0 {
			l.batchSize = size
		}
	}
}

// WithFlushInterval sets how often to flush logs regardless of batch size
func WithFlushInterval(interval time.Duration) AsyncOption {
	return func(l *AsyncLogger) {
		if interval > 0 {
			if l.flushTicker != nil {
				l.flushTicker.Stop()
			}
			l.flushTicker = time.NewTicker(interval)
		}
	}
}

// WithWorkers sets the number of worker goroutines
func WithWorkers(count int) AsyncOption {
	return func(l *AsyncLogger) {
		if count > 0 {
			l.workers = count
		}
	}
}

// WithBlockOnFull determines behavior when buffer is full (block vs error)
func WithBlockOnFull(block bool) AsyncOption {
	return func(l *AsyncLogger) {
		l.blockOnFull = block
	}
}

// WithAsyncFormatter sets the formatter for async logger
func WithAsyncFormatter(f formatter.Formatter) AsyncOption {
	return func(l *AsyncLogger) {
		// Only set formatter if not nil
		if f != nil {
			l.formatter = f
		}
	}
}

// WithAsyncJob sets the job name for async logger
func WithAsyncJob(job string) AsyncOption {
	return func(l *AsyncLogger) {
		l.job = job
	}
}

// WithAsyncMetadata adds metadata to async logger
func WithAsyncMetadata(key string, value interface{}) AsyncOption {
	return func(l *AsyncLogger) {
		l.metadata[key] = value
	}
}

// WithMaxRequestSize sets the maximum size of HTTP requests in bytes
func WithMaxRequestSize(sizeBytes int) AsyncOption {
	return func(l *AsyncLogger) {
		if sizeBytes > 0 {
			l.maxRequestSize = sizeBytes
		}
	}
}

// WithErrorHandler sets a custom handler for internal errors
func WithErrorHandler(handler errors.ErrorHandler) AsyncOption {
	return func(l *AsyncLogger) {
		if handler != nil {
			l.errorHandler = handler
		}
	}
}

// start initializes and starts the worker goroutines
func (l *AsyncLogger) start() {
	for i := 0; i < l.workers; i++ {
		l.wg.Add(1)
		go l.processLogs()
	}

	// Start a dedicated worker for handling flush markers
	l.wg.Add(1)
	go l.handleFlushMarkers()

	// Start a goroutine to listen for flush ticks
	l.wg.Add(1)
	go l.flushRoutine()
}

// handleFlushMarkers processes flush markers separately to avoid delays in regular log processing
func (l *AsyncLogger) handleFlushMarkers() {
	defer l.wg.Done()
	for {
		select {
		case entry, more := <-l.buffer:
			if !more {
				return // Buffer closed
			}

			if l.isFlushMarker(entry) {
				l.processFlushMarker(entry)
				continue
			}

			// Try to requeue non-flush entries
			if !l.shouldRequeueEntry() {
				// If we're shutting down, just drop the entry
				continue
			}

			l.requeueEntry(entry)

		case <-l.done:
			return
		}
	}
}

func (l *AsyncLogger) shouldRequeueEntry() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.closed || l.buffer == nil {
		return false
	}

	select {
	case <-l.done:
		return false
	default:
		return true
	}
}

func (l *AsyncLogger) requeueEntry(entry asyncLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Don't send on a closed channel
	if !l.closed {
		select {
		case l.buffer <- entry:
			// Successfully sent
		default:
			// Channel is full, log a warning if possible
			// This is a fallback if we can't requeue
			// No need to panic, just drop the entry
		}
	}
}

// flushRoutine periodically triggers a flush based on the ticker
func (l *AsyncLogger) flushRoutine() {
	defer l.wg.Done()
	for {
		select {
		case <-l.flushTicker.C:
			// On each tick, do nothing - the worker goroutines will check
			// the time themselves and flush if needed
		case <-l.done:
			return
		}
	}
}

// processBatch sends a batch of log entries to Loki
// It groups entries by job name for efficiency, keeps track of batch size to avoid
// exceeding HTTP request limits, and splits batches when they grow too large.
//
// The process:
// 1. Group entries by job name
// 2. Format each group of entries using the batch formatter
// 3. If the batch grows too large, send what's accumulated so far
// 4. Add the newly formatted entries to the batch
// 5. Send any remaining entries at the end
func (l *AsyncLogger) processBatch(batch []asyncLogEntry) {
	// Group entries by job to optimize Loki streams
	jobBatches := make(map[string][]asyncLogEntry)

	for _, entry := range batch {
		jobBatches[entry.job] = append(jobBatches[entry.job], entry)
	}

	// Create a combined LokiEntry with streams for all jobs
	combinedEntry := client.LokiEntry{
		Streams: []client.LokiStream{},
	}

	// Track estimated size to prevent exceeding HTTP request limits
	estimatedSize := 0

	// Process each job batch
	for job, entries := range jobBatches {
		if len(entries) == 0 {
			continue
		}

		// Format entries for this job
		formattedEntries := make([]formatter.LogEntry, 0, len(entries))
		for _, entry := range entries {
			formattedEntry := l.createFormattedEntry(entry)
			formattedEntries = append(formattedEntries, formattedEntry)
		}

		// Use batch formatting for this job
		jobEntry, err := l.formatter.FormatBatch(job, formattedEntries)
		if err != nil {
			// If batch formatting fails, report error and process entries individually
			l.errorHandler(fmt.Errorf("%w: batch formatting failed: %v", errors.ErrProcessingFailed, err))
			l.processSingleEntries(formattedEntries)
			continue
		}

		// Estimate the size of the batch
		estimatedSize += l.estimateBatchSize(jobEntry)

		// If adding this batch would exceed size limits, send the current batch first
		if estimatedSize > l.maxRequestSize && len(combinedEntry.Streams) > 0 {
			// Send the current combined batch
			if err := l.sender.Send(combinedEntry); err != nil {
				l.errorHandler(fmt.Errorf("%w: sending batch failed: %v", errors.ErrProcessingFailed, err))
			}

			// Reset for a new batch
			combinedEntry = client.LokiEntry{
				Streams: []client.LokiStream{},
			}
			estimatedSize = 0
		}

		// Add this job's streams to the combined entry
		combinedEntry.Streams = append(combinedEntry.Streams, jobEntry.Streams...)
	}

	// Send the final combined entry with all remaining job batches
	if len(combinedEntry.Streams) > 0 {
		if err := l.sender.Send(combinedEntry); err != nil {
			l.errorHandler(fmt.Errorf("%w: sending final batch failed: %v", errors.ErrProcessingFailed, err))
		}
	}
}

// New helper method to estimate batch size
func (l *AsyncLogger) estimateBatchSize(jobEntry client.LokiEntry) int {
	estimatedSize := 0
	for _, stream := range jobEntry.Streams {
		for _, value := range stream.Values {
			if len(value) >= 2 {
				// Add the size of this log entry plus some overhead
				estimatedSize += len(value[1]) + EstimatedEntryOverhead // 100 bytes for overhead
			}
		}
	}
	return estimatedSize
}

// processSingleEntries falls back to processing entries one by one if batch formatting fails
func (l *AsyncLogger) processSingleEntries(entries []formatter.LogEntry) {
	for _, entry := range entries {
		// Format and send each entry
		formatted, err := l.formatter.Format(entry)
		if err != nil {
			l.errorHandler(fmt.Errorf("%w: individual formatting failed: %v", errors.ErrProcessingFailed, err))
			continue
		}

		// Send the formatted entry
		if err = l.sender.Send(formatted); err != nil {
			l.errorHandler(fmt.Errorf("%w: sending individual entry failed: %v", errors.ErrProcessingFailed, err))
		}
	}
}

// isFlushMarker checks if an entry is a flush marker
func (l *AsyncLogger) isFlushMarker(entry asyncLogEntry) bool {
	for i, kv := range entry.keyvals {
		if i+1 < len(entry.keyvals) {
			if k, ok := kv.(string); ok && k == "_flush_marker" {
				return true
			}
		}
	}
	return false
}

// processFlushMarker handles a flush marker entry
func (l *AsyncLogger) processFlushMarker(entry asyncLogEntry) {
	for i, kv := range entry.keyvals {
		if i+1 < len(entry.keyvals) {
			if k, ok := kv.(string); ok && k == "_flush_marker" {
				if ch, ok := entry.keyvals[i+1].(chan struct{}); ok {
					// Signal that flush is complete
					close(ch)
				}
			}
		}
	}
}

// createFormattedEntry converts an asyncLogEntry to a formatter.LogEntry
func (l *AsyncLogger) createFormattedEntry(entry asyncLogEntry) formatter.LogEntry {
	// Combine passed key-values with default metadata
	allKeyVals := make([]interface{}, 0, len(entry.keyvals)+2)
	// Add message first
	allKeyVals = append(allKeyVals, "message", entry.message)

	// Add provided key-values
	allKeyVals = append(allKeyVals, entry.keyvals...)

	// Add metadata
	for k, v := range entry.metadata {
		allKeyVals = append(allKeyVals, k, v)
	}

	// Create a formatter.LogEntry
	return formatter.NewLogEntry(entry.job, entry.level, allKeyVals...)
}

// Info logs an info level message asynchronously
func (l *AsyncLogger) Info(message string, keyvals ...interface{}) error {
	return l.log("info", message, keyvals...)
}

// Error logs an error level message asynchronously
func (l *AsyncLogger) Error(message string, keyvals ...interface{}) error {
	return l.log("error", message, keyvals...)
}

// Debug logs a debug level message asynchronously
func (l *AsyncLogger) Debug(message string, keyvals ...interface{}) error {
	return l.log("debug", message, keyvals...)
}

// Warn logs a warning level message asynchronously
func (l *AsyncLogger) Warn(message string, keyvals ...interface{}) error {
	return l.log("warn", message, keyvals...)
}

// log is the internal logging function that sends to the processing queue
func (l *AsyncLogger) log(level string, message string, keyvals ...interface{}) error {
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return fmt.Errorf("%w: cannot log to closed logger", errors.ErrLoggerClosed)
	}
	l.mu.RUnlock()

	// Create the entry
	entry := asyncLogEntry{
		job:      l.job,
		level:    level,
		message:  message,
		keyvals:  keyvals,
		metadata: l.copyMetadata(), // Copy metadata to avoid race conditions
	}

	// If configured to block, try sending with potential blocking
	if l.blockOnFull {
		select {
		case l.buffer <- entry:
			return nil
		case <-l.done:
			return fmt.Errorf("%w: logger is shutting down", errors.ErrShutdown)
		}
	}

	// Non-blocking send - will fail if buffer is full
	select {
	case l.buffer <- entry:
		return nil
	default:
		return fmt.Errorf("%w: log buffer is full", errors.ErrBufferFull)
	}
}

// WithContext returns a new logger with additional context
func (l *AsyncLogger) WithContext(keyvals ...interface{}) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.closed {
		// If closed, return self to avoid creating more resources
		return l
	}

	newLogger := &AsyncLogger{
		formatter:      l.formatter,
		job:            l.job,
		metadata:       l.copyMetadata(),
		sender:         l.sender,
		buffer:         l.buffer,
		batchSize:      l.batchSize,
		flushTicker:    l.flushTicker,
		workers:        l.workers,
		blockOnFull:    l.blockOnFull,
		maxRequestSize: l.maxRequestSize,
		done:           l.done,
		closed:         l.closed,
	}

	// Process new context
	processKeyvals(newLogger.metadata, keyvals...)

	return newLogger
}

// WithJob returns a new logger with a different job name
func (l *AsyncLogger) WithJob(job string) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.closed {
		// If closed, return self to avoid creating more resources
		return l
	}

	newLogger := &AsyncLogger{
		formatter:      l.formatter,
		job:            job,
		metadata:       l.copyMetadata(),
		sender:         l.sender,
		buffer:         l.buffer,
		batchSize:      l.batchSize,
		flushTicker:    l.flushTicker,
		workers:        l.workers,
		blockOnFull:    l.blockOnFull,
		maxRequestSize: l.maxRequestSize,
		done:           l.done,
		closed:         l.closed,
	}

	return newLogger
}

// copyMetadata safely copies the logger's metadata
func (l *AsyncLogger) copyMetadata() map[string]interface{} {
	newMetadata := make(map[string]interface{}, len(l.metadata))
	for k, v := range l.metadata {
		newMetadata[k] = v
	}
	return newMetadata
}

// Flush waits for all queued logs to be processed
func (l *AsyncLogger) Flush() error {
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return errors.ErrLoggerClosed
	}
	l.mu.RUnlock()

	// Create a temporary channel to signal completion of flush
	flushDone := make(chan struct{})
	flushCtx, cancel := context.WithTimeout(context.Background(), FlushMarkerTimeout)
	defer cancel()

	// Add a special marker entry that will signal when processing reaches this point
	select {
	case l.buffer <- asyncLogEntry{
		job:      l.job,
		level:    "info",
		message:  "flush-marker",
		keyvals:  []interface{}{"_flush_marker", flushDone},
		metadata: map[string]interface{}{},
	}:
		// Successfully added flush marker
	case <-flushCtx.Done():
		return fmt.Errorf("%w: timed out waiting to add flush marker", errors.ErrTimeout)
	}

	// Wait for the marker to be processed with timeout
	select {
	case <-flushDone:
		return nil
	case <-flushCtx.Done():
		return fmt.Errorf("%w: timed out waiting for flush to complete", errors.ErrTimeout)
	}
}

// Close shuts down the logger and waits for all pending logs to be processed
func (l *AsyncLogger) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return errors.ErrLoggerClosed
	}

	// Mark as closed first to prevent new logs
	l.closed = true
	l.mu.Unlock()

	// Flush remaining logs
	if err := l.Flush(); err != nil && !errors.Is(err, errors.ErrLoggerClosed) {
		return fmt.Errorf("%w: failed to flush logs during close", err)
	}

	// Stop the ticker
	if l.flushTicker != nil {
		l.flushTicker.Stop()
	}

	// Signal all workers to stop
	close(l.done)

	// Close the buffer channel
	close(l.buffer)

	// Wait for all workers to finish
	l.wg.Wait()

	return nil
}

// IsClosed returns whether the logger has been closed
func (l *AsyncLogger) IsClosed() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.closed
}

// processLogs is the worker routine that collects and sends batches of logs
func (l *AsyncLogger) processLogs() {
	defer l.wg.Done()
	var batch []asyncLogEntry

	for {
		select {
		case entry, more := <-l.buffer:
			if !more {
				// Channel was closed, process remaining batch and exit
				if len(batch) > 0 {
					l.processBatch(batch)
				}
				return
			}

			// Add entry to batch
			batch = append(batch, entry)

			// If batch is full, process it
			if len(batch) >= l.batchSize {
				l.processBatch(batch)
				batch = batch[:0] // Clear batch
			}

		case <-l.flushTicker.C:
			// If there are any entries in the batch, process them
			if len(batch) > 0 {
				l.processBatch(batch)
				batch = batch[:0] // Clear batch
			}

		case <-l.done:
			// Process any remaining entries and exit
			if len(batch) > 0 {
				l.processBatch(batch)
			}
			return
		}
	}
}
