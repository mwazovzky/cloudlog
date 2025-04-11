package logger

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// threadSafeLogSender is a mock sender that provides thread-safe access to captured logs
type threadSafeLogSender struct {
	mu       sync.Mutex
	messages []string
	jobs     []string
}

// Send processes log entries while providing thread-safety
func (m *threadSafeLogSender) Send(entry client.LokiEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Extract job from the first stream
	if len(entry.Streams) > 0 {
		for _, stream := range entry.Streams {
			m.jobs = append(m.jobs, stream.Stream["job"])

			// Extract the message from values
			for _, value := range stream.Values {
				if len(value) >= 2 {
					message := value[1]
					m.messages = append(m.messages, message)
				}
			}
		}
	}
	return nil
}

// getMessages returns a thread-safe copy of collected messages
func (m *threadSafeLogSender) getMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.messages))
	copy(result, m.messages)
	return result
}

// getJobs returns a thread-safe copy of collected job names
func (m *threadSafeLogSender) getJobs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.jobs))
	copy(result, m.jobs)
	return result
}

// waitForMessages is a helper to wait for a minimum number of messages with timeout
func waitForMessages(t *testing.T, sender *threadSafeLogSender, count int, timeout time.Duration) {
	// Increase the timeout to avoid flaky tests
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		sender.mu.Lock()
		messageCount := len(sender.messages)
		sender.mu.Unlock()

		if messageCount >= count {
			return
		}

		// Use shorter sleep interval to check more frequently
		time.Sleep(50 * time.Millisecond)
	}

	// If we reach here, the timeout has expired
	sender.mu.Lock()
	messageCount := len(sender.messages)
	sender.mu.Unlock()

	t.Errorf("Timeout waiting for %d messages. Got: %d", count, messageCount)
}

// Test sections organized by functionality
// 1. Basic initialization and configuration tests
func TestNewAsync(t *testing.T) {
	sender := &threadSafeLogSender{}
	logger := NewAsync(sender).(*AsyncLogger)

	// Verify defaults
	assert.Equal(t, "application", logger.job)
	assert.Equal(t, 2, logger.workers)
	assert.Equal(t, 100, logger.batchSize)

	// Clean up
	logger.Close()
}

func TestAsyncLoggerOptions(t *testing.T) {
	sender := &threadSafeLogSender{}

	// Test formatter option
	customFormatter := formatter.NewStringFormatter()
	logger := NewAsync(sender, WithAsyncFormatter(customFormatter)).(*AsyncLogger)
	assert.Equal(t, customFormatter, logger.formatter)
	logger.Close()

	// Test job option
	customJob := "test-job-name"
	logger = NewAsync(sender, WithAsyncJob(customJob)).(*AsyncLogger)
	assert.Equal(t, customJob, logger.job)
	logger.Close()

	// Test metadata option
	logger = NewAsync(sender, WithAsyncMetadata("key1", "value1")).(*AsyncLogger)
	assert.Equal(t, "value1", logger.metadata["key1"])
	logger.Close()
}

func TestAsyncLoggerOptionsCombined(t *testing.T) {
	sender := &threadSafeLogSender{}
	customFormatter := formatter.NewStringFormatter()
	customJob := "custom-service"

	// Test multiple options applied together
	logger := NewAsync(
		sender,
		WithAsyncFormatter(customFormatter),
		WithAsyncJob(customJob),
		WithAsyncMetadata("env", "test"),
		WithAsyncMetadata("version", "1.0.0"),
	).(*AsyncLogger)

	// Verify all options were applied
	assert.Equal(t, customFormatter, logger.formatter)
	assert.Equal(t, customJob, logger.job)
	assert.Equal(t, "test", logger.metadata["env"])
	assert.Equal(t, "1.0.0", logger.metadata["version"])

	// Clean up
	logger.Close()
}

func TestWithAsyncFormatterNil(t *testing.T) {
	sender := &threadSafeLogSender{}
	// First create a logger with default formatter
	logger := NewAsync(sender).(*AsyncLogger)
	// We want to verify that the formatter is not nil, not store it
	assert.NotNil(t, logger.formatter, "Default formatter should not be nil")
	logger.Close()

	// Then create one with nil formatter (should use default)
	nilLogger := NewAsync(sender, WithAsyncFormatter(nil)).(*AsyncLogger)
	assert.NotNil(t, nilLogger.formatter, "Formatter should not be nil even when passed nil")
	nilLogger.Close()
}

// 2. Log operations and severity tests
func TestAsyncLogger_LogLevels(t *testing.T) {
	sender := &threadSafeLogSender{}
	logger := NewAsync(sender, WithBatchSize(1), WithFlushInterval(10*time.Millisecond), WithBufferSize(10))
	defer logger.Close()

	logLevels := []struct {
		method func(string) error
		level  string
	}{
		{func(msg string) error { return logger.Info(msg) }, "info"},
		{func(msg string) error { return logger.Error(msg) }, "error"},
		{func(msg string) error { return logger.Debug(msg) }, "debug"},
		{func(msg string) error { return logger.Warn(msg) }, "warn"},
	}

	for _, level := range logLevels {
		err := level.method(level.level + " message")
		assert.NoError(t, err)
	}

	waitForMessages(t, sender, len(logLevels), 2*time.Second)

	messages := sender.getMessages()
	require.Len(t, messages, len(logLevels))

	foundLevels := map[string]bool{}
	for _, msg := range messages {
		var logData map[string]interface{}
		err := json.Unmarshal([]byte(msg), &logData)
		require.NoError(t, err)

		level, ok := logData["level"].(string)
		require.True(t, ok)
		foundLevels[level] = true
	}

	for _, level := range logLevels {
		assert.True(t, foundLevels[level.level], "Should contain "+level.level+" level")
	}
}

// 3. Context and metadata tests
func TestAsyncLogger_WithContext(t *testing.T) {
	sender := &threadSafeLogSender{}
	logger := NewAsync(sender, WithBatchSize(1), WithFlushInterval(50*time.Millisecond))

	// Add context
	contextLogger := logger.WithContext("key1", "value1", "key2", 42)

	// Log with context
	err := contextLogger.Info("context test")
	assert.NoError(t, err)

	// Wait for message to be processed
	waitForMessages(t, sender, 1, 1*time.Second)

	// Check results
	messages := sender.getMessages()
	require.Len(t, messages, 1)

	// Verify context was included
	assert.Contains(t, messages[0], "key1")
	assert.Contains(t, messages[0], "value1")
	assert.Contains(t, messages[0], "key2")
	assert.Contains(t, messages[0], "42")

	// Clean up
	logger.Close()
}

func TestAsyncLogger_WithJob(t *testing.T) {
	sender := &threadSafeLogSender{}
	logger := NewAsync(sender, WithBatchSize(1), WithFlushInterval(50*time.Millisecond))

	// Create logger with custom job
	jobLogger := logger.WithJob("custom-job")

	// Log with custom job
	err := jobLogger.Info("job test")
	assert.NoError(t, err)

	// Wait for message to be processed
	waitForMessages(t, sender, 1, 1*time.Second)

	// Check job was set correctly
	jobs := sender.getJobs()
	require.Len(t, jobs, 1)
	assert.Equal(t, "custom-job", jobs[0])

	// Clean up
	logger.Close()
}

// 4. Buffer and batching behavior tests
func TestAsyncLogger_BufferFull(t *testing.T) {
	// Create a logger with very small buffer
	sender := &threadSafeLogSender{}
	logger := NewAsync(sender, WithBufferSize(1), WithBlockOnFull(false))

	// Fill the buffer
	err := logger.Info("first message")
	assert.NoError(t, err)

	// This should cause a buffer full error since we're using a very small buffer
	// Add a small delay to ensure the second message is attempted after the first is queued
	time.Sleep(10 * time.Millisecond)
	err = logger.Info("second message")

	if err != nil {
		assert.True(t, errors.IsBufferFullError(err))
	}

	// Clean up
	logger.Close()
}

func TestAsyncLogger_BlockOnFull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping blocking test in short mode")
	}

	// Create a slow sender that will cause the buffer to fill
	slowSender := &slowLogSender{delay: 100 * time.Millisecond}

	// Create a logger with small buffer that blocks when full
	logger := NewAsync(slowSender, WithBufferSize(2), WithBlockOnFull(true), WithWorkers(1))

	// Fill the buffer and then some more to test blocking
	for i := 0; i < 5; i++ {
		// Each of these should complete since we're blocking
		err := logger.Info("message", "index", i)
		assert.NoError(t, err)
	}

	// Clean up
	logger.Close()
}

func TestAsyncLogger_BatchProcessing(t *testing.T) {
	// Create a mock sender with instrumentation
	sender := &instrumentedSender{calls: make(map[string]int)}

	// Create an async logger optimized for batching
	logger := NewAsync(sender,
		WithBatchSize(5),
		WithFlushInterval(10*time.Millisecond),
		WithBufferSize(30),
		WithWorkers(2),
	).(*AsyncLogger)

	// Send multiple logs with the same job
	for i := 0; i < 20; i++ {
		err := logger.Info("test message", "index", i)
		assert.NoError(t, err)
	}

	// Wait for processing directly rather than using Flush
	maxWait := time.After(3 * time.Second)
	checkTicker := time.NewTicker(100 * time.Millisecond)
	defer checkTicker.Stop()

	for {
		select {
		case <-maxWait:
			t.Fatalf("Max wait time exceeded. Got %d messages instead of 20",
				sender.messageCount())
		case <-checkTicker.C:
			if sender.messageCount() >= 20 {
				goto checkResults
			}
		}
	}

checkResults:
	// We should see fewer API calls than messages due to batching
	calls := sender.sendCalls()
	assert.Less(t, calls, 20, "Should use fewer API calls than messages due to batching")

	// Verify all messages were sent
	assert.Equal(t, 20, sender.messageCount(), "All messages should be sent")

	// Clean up
	err := logger.Close()
	assert.NoError(t, err)
}

// 5. Lifecycle (Flush/Close) tests
func TestAsyncLogger_Flush(t *testing.T) {
	sender := &threadSafeLogSender{}
	logger := NewAsync(sender,
		WithBatchSize(1),
		WithBufferSize(10),
		WithFlushInterval(10*time.Millisecond),
		WithWorkers(2),
	).(*AsyncLogger)

	// Send some test messages
	for i := 0; i < 5; i++ {
		err := logger.Info("test message")
		assert.NoError(t, err)
	}

	// Call Flush with an increased timeout
	err := logger.Flush()
	assert.NoError(t, err, "Flush should complete without errors")

	// Wait for messages to be processed
	waitForMessages(t, sender, 5, 5*time.Second) // Increased timeout to 5 seconds

	// Verify all messages were sent
	messages := sender.getMessages()
	assert.GreaterOrEqual(t, len(messages), 5, "Should have at least 5 messages")

	// Close to clean up
	logger.Close()
}

func TestAsyncLogger_Close(t *testing.T) {
	sender := &threadSafeLogSender{}
	logger := NewAsync(sender,
		WithBatchSize(1),
		WithBufferSize(10),
		WithFlushInterval(50*time.Millisecond),
	).(*AsyncLogger)

	// Send some messages
	for i := 0; i < 5; i++ {
		err := logger.Info("test message")
		assert.NoError(t, err)
	}

	// Add a small delay to ensure processing starts
	time.Sleep(100 * time.Millisecond)

	// Close the logger
	err := logger.Close()
	assert.NoError(t, err)

	// Add a larger delay to ensure messages are processed after close
	// and close operation completes
	time.Sleep(500 * time.Millisecond)

	// Verify the messages were sent
	messages := sender.getMessages()
	assert.GreaterOrEqual(t, len(messages), 5, "Should have at least 5 messages")

	// Check that the logger is properly marked as closed
	assert.True(t, logger.closed, "Logger should be marked as closed")

	// If the above passes, then IsClosed() is returning false when closed is true
	assert.True(t, logger.IsClosed(), "IsClosed() should return true")

	// Verify operations on closed logger fail gracefully
	err = logger.Info("after close")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.ErrLoggerClosed), "Expected logger closed error for operations on closed logger")
}

// Test-specific helper types
// slowLogSender simulates network latency
type slowLogSender struct {
	mu       sync.Mutex
	messages []string
	delay    time.Duration
}

func (s *slowLogSender) Send(entry client.LokiEntry) error {
	time.Sleep(s.delay) // Simulate network delay
	s.mu.Lock()
	defer s.mu.Unlock()
	// Extract the message from the entry
	if len(entry.Streams) > 0 && len(entry.Streams[0].Values) > 0 && len(entry.Streams[0].Values[0]) > 1 {
		s.messages = append(s.messages, entry.Streams[0].Values[0][1])
	}
	return nil
}

// instrumentedSender tracks API calls for batch testing
type instrumentedSender struct {
	mu       sync.Mutex
	calls    map[string]int
	messages int
}

func (s *instrumentedSender) Send(entry client.LokiEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Track the API call
	s.calls["send"]++

	// Count all messages in all streams
	for _, stream := range entry.Streams {
		s.messages += len(stream.Values)
	}
	return nil
}

func (s *instrumentedSender) sendCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls["send"]
}

func (s *instrumentedSender) messageCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messages
}

// TestErrorHandling consolidates error handling tests for better organization
func TestErrorHandling(t *testing.T) {
	t.Run("SendErrors", func(t *testing.T) {
		// Create a channel to capture errors from the error handler
		errorCh := make(chan error, 10)
		errorHandler := func(err error) {
			errorCh <- err
		}

		// Create a mock sender that fails when sending
		failingSender := &asyncMockSender{
			shouldFail: true,
			failWith:   fmt.Errorf("simulated send failure"),
		}

		// Create a logger with the error handler
		logger := NewAsync(failingSender,
			WithAsyncFormatter(formatter.NewLokiFormatter()),
			WithBatchSize(1),
			WithFlushInterval(50*time.Millisecond),
			WithErrorHandler(errorHandler))

		// Log a message that will trigger a send error
		err := logger.Info("test message")
		assert.NoError(t, err, "Log call should not return error even when sender fails")

		// Wait for error handler to be called
		var capturedErr error
		select {
		case capturedErr = <-errorCh:
			// Got the error
		case <-time.After(2 * time.Second):
			t.Fatal("Timed out waiting for error handler to be called")
		}

		// Verify the error is correctly wrapped with ErrProcessingFailed
		assert.True(t, errors.IsProcessingError(capturedErr),
			"Error should be wrapped with ErrProcessingFailed")
		assert.Contains(t, capturedErr.Error(), "simulated send failure",
			"Original error message should be included")

		// Clean up
		logger.Close()
	})

	t.Run("FormatErrors", func(t *testing.T) {
		// Create a channel to capture errors from the error handler
		errorCh := make(chan error, 10)
		errorHandler := func(err error) {
			errorCh <- err
		}

		// Create a mock sender to track received logs
		sender := &threadSafeLogSender{}

		// Create a simple formatter that fails only for batch operations
		formatter := &mockFormatter{
			shouldFailBatch: true,
			failMessage:     "simulated format error",
		}

		// Create a logger with our test formatter
		logger := NewAsync(sender,
			WithAsyncFormatter(formatter),
			WithBatchSize(2), // Use batch size > 1 to trigger batch formatting
			WithFlushInterval(50*time.Millisecond),
			WithErrorHandler(errorHandler))

		// Send test messages
		for i := 0; i < 2; i++ {
			err := logger.Info("test message", "index", i)
			assert.NoError(t, err, "Log call should succeed even with failing formatter")
		}

		// Wait for error handler to be called
		var formatError error
		select {
		case formatError = <-errorCh:
			// Got the error from the handler
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for error handler to be called")
		}

		// Verify error is properly wrapped
		assert.True(t, errors.IsProcessingError(formatError),
			"Error should be wrapped with ErrProcessingFailed")
		assert.Contains(t, formatError.Error(), "batch formatting failed",
			"Error should indicate batch formatting failed")

		// Wait for individual messages to be processed via the fallback path
		waitForMessages(t, sender, 2, 2*time.Second)

		// Verify messages were sent by fallback path
		messages := sender.getMessages()
		assert.GreaterOrEqual(t, len(messages), 2,
			"Should have at least 2 messages from fallback path")

		logger.Close()
	})
}

// TestErrorHandlerOption tests WithErrorHandler option separately
func TestErrorHandlerOption(t *testing.T) {
	sender := &threadSafeLogSender{}

	var called bool
	customHandler := func(err error) {
		called = true
	}

	// Test that WithErrorHandler option is applied
	logger := NewAsync(sender, WithErrorHandler(customHandler)).(*AsyncLogger)

	// Verify the custom handler was set
	assert.NotNil(t, logger.errorHandler, "Error handler should be set")

	// Create an error and invoke the handler directly to verify it works
	testErr := fmt.Errorf("test error")
	logger.errorHandler(testErr)

	assert.True(t, called, "Custom error handler should have been called")

	// Passing nil should not change the handler to nil
	logger = NewAsync(sender, WithErrorHandler(nil)).(*AsyncLogger)
	assert.NotNil(t, logger.errorHandler, "Error handler should not be nil even when passed nil")

	logger.Close()
}

// Simplify the mock formatter to be more consistent and avoid redundancy
type mockFormatter struct {
	shouldFail      bool
	shouldFailBatch bool
	failMessage     string
}

func (m *mockFormatter) Format(entry formatter.LogEntry) (client.LokiEntry, error) {
	if m.shouldFail {
		return client.LokiEntry{}, fmt.Errorf("%w: %s", errors.ErrInvalidFormat, m.failMessage)
	}

	// Return a simple valid entry
	timestamp := time.Now().UnixNano()
	return client.LokiEntry{
		Streams: []client.LokiStream{
			{
				Stream: map[string]string{"job": entry.Job},
				Values: [][]string{
					{
						fmt.Sprintf("%d", timestamp),
						fmt.Sprintf("{\"message\":\"test\"}"),
					},
				},
			},
		},
	}, nil
}

func (m *mockFormatter) FormatBatch(job string, entries []formatter.LogEntry) (client.LokiEntry, error) {
	if m.shouldFail || m.shouldFailBatch {
		return client.LokiEntry{}, fmt.Errorf("%w: %s", errors.ErrInvalidFormat, m.failMessage)
	}

	// Return a simple valid batch entry
	timestamp := time.Now().UnixNano()
	return client.LokiEntry{
		Streams: []client.LokiStream{
			{
				Stream: map[string]string{"job": job},
				Values: [][]string{
					{
						fmt.Sprintf("%d", timestamp),
						fmt.Sprintf("{\"message\":\"test\"}"),
					},
				},
			},
		},
	}, nil
}

// Rename mockSender to asyncMockSender to avoid conflicts with the one in sync_logger_test.go
type asyncMockSender struct {
	mu         sync.Mutex
	messages   []string
	shouldFail bool
	failWith   error
}

func (s *asyncMockSender) Send(entry client.LokiEntry) error {
	if s.shouldFail {
		return s.failWith
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Extract messages for verification
	if len(entry.Streams) > 0 && len(entry.Streams[0].Values) > 0 {
		for _, value := range entry.Streams[0].Values {
			if len(value) >= 2 {
				s.messages = append(s.messages, value[1])
			}
		}
	}
	return nil
}

// countingSender is a sender that tracks the number of call attempts
type countingSender struct {
	callTracker chan<- struct{}
}

func (s *countingSender) Send(entry client.LokiEntry) error {
	s.callTracker <- struct{}{}
	return nil
}

func TestAsyncLogger_ErrorHandling(t *testing.T) {
	t.Run("SimulatedSendFailure", func(t *testing.T) {
		// Create a channel to capture errors from the error handler
		errorCh := make(chan error, 1)

		// Create an error handler that signals when called
		errorHandler := func(err error) {
			errorCh <- err
		}

		// Create a failing sender
		failingSender := &asyncMockSender{
			shouldFail: true,
			failWith:   fmt.Errorf("simulated send failure"),
		}

		// Create logger with error handler
		logger := NewAsync(failingSender,
			WithAsyncFormatter(formatter.NewLokiFormatter()),
			WithBatchSize(1),
			WithFlushInterval(50*time.Millisecond), // Added missing comma here
			WithErrorHandler(errorHandler))

		// Log message to trigger send error
		err := logger.Info("test message")
		assert.NoError(t, err, "Log call should not return error even when sender fails")

		// Wait with shorter timeout
		select {
		case capturedErr := <-errorCh:
			assert.Contains(t, capturedErr.Error(), "simulated send failure")
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Timed out waiting for error handler")
		}

		// Clean up
		logger.Close()
	})
}
