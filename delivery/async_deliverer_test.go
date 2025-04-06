package delivery

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	clouderrors "github.com/mwazovzky/cloudlog/errors"
)

// MockSlowSender simulates a slow or unreliable sender for testing
type MockSlowSender struct {
	mu          sync.Mutex
	calls       []MockCall
	ShouldError bool
	ErrorUntil  int // Error until this many calls
	Delay       time.Duration
	PanicOnCall int // Panic on this call number (0 to disable)
	callCount   int32
}

type MockCall struct {
	Job       string
	Formatted []byte
}

func (m *MockSlowSender) Send(job string, formatted []byte) error {
	currentCall := atomic.AddInt32(&m.callCount, 1)

	// Optional panic for testing worker recovery
	if (m.PanicOnCall > 0) && (int(currentCall) == m.PanicOnCall) {
		panic("mock panic")
	}

	// Simulate processing delay
	if m.Delay > 0 {
		time.Sleep(m.Delay)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, MockCall{Job: job, Formatted: formatted})

	// Simulate errors for the first N calls if configured
	if m.ShouldError || (m.ErrorUntil > 0 && len(m.calls) <= m.ErrorUntil) {
		return errors.New("mock slow sender error")
	}

	return nil
}

func (m *MockSlowSender) Calls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.calls))
	copy(result, m.calls)
	return result
}

func (m *MockSlowSender) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func TestAsyncDeliverer_BasicOperation(t *testing.T) {
	mockSender := &MockSlowSender{}

	config := DefaultConfig()
	config.Async = true
	config.QueueSize = 10
	config.Workers = 1

	deliverer := NewAsyncDeliverer(mockSender, config)

	// Deliver a message
	err := deliverer.Deliver("test-job", "info", "Test message", []byte("test"), time.Now())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Allow some time for processing
	time.Sleep(50 * time.Millisecond)

	calls := mockSender.Calls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 call to Send, got %d", len(calls))
	}

	if len(calls) > 0 && calls[0].Job != "test-job" {
		t.Errorf("Expected job 'test-job', got '%s'", calls[0].Job)
	}

	// Check stats
	stats := deliverer.Status()
	if stats.Buffered != 0 {
		t.Errorf("Expected 0 buffered messages, got %d", stats.Buffered)
	}
	if stats.Delivered != 1 {
		t.Errorf("Expected 1 delivered message, got %d", stats.Delivered)
	}

	// Clean up
	err = deliverer.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

func TestAsyncDeliverer_QueueOverflow(t *testing.T) {
	// Use a very slow sender and a small queue
	mockSender := &MockSlowSender{Delay: 100 * time.Millisecond}

	config := DefaultConfig()
	config.Async = true
	config.QueueSize = 2 // Very small queue
	config.Workers = 1

	deliverer := NewAsyncDeliverer(mockSender, config)

	// Fill the queue and buffer
	err1 := deliverer.Deliver("job1", "info", "msg1", []byte("test1"), time.Now())
	err2 := deliverer.Deliver("job2", "info", "msg2", []byte("test2"), time.Now())
	err3 := deliverer.Deliver("job3", "info", "msg3", []byte("test3"), time.Now())

	// First two should succeed, third should fail with buffer full error
	if err1 != nil {
		t.Errorf("First deliver should succeed, got: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second deliver should succeed, got: %v", err2)
	}
	if err3 == nil {
		t.Error("Third deliver should fail with buffer full error")
	} else if !clouderrors.Is(err3, clouderrors.ErrBufferFull) {
		t.Errorf("Expected buffer full error, got: %v", err3)
	}

	// Check stats after overflow
	stats := deliverer.Status()
	if stats.Dropped != 1 {
		t.Errorf("Expected 1 dropped message, got %d", stats.Dropped)
	}

	// Allow processing to complete
	time.Sleep(250 * time.Millisecond)

	// Clean up
	deliverer.Close()
}

func TestAsyncDeliverer_Retry(t *testing.T) {
	// Sender that will fail the first attempt
	mockSender := &MockSlowSender{ErrorUntil: 1}

	config := DefaultConfig()
	config.Async = true
	config.MaxRetries = 3
	config.RetryInterval = 10 * time.Millisecond

	deliverer := NewAsyncDeliverer(mockSender, config)

	// Send a message that will need to be retried
	err := deliverer.Deliver("retry-job", "info", "Retry message", []byte("test"), time.Now())
	if err != nil {
		t.Errorf("Deliver should not return error, got: %v", err)
	}

	// Allow time for retries and processing
	time.Sleep(100 * time.Millisecond)

	// Should eventually succeed
	stats := deliverer.Status()
	if stats.Delivered != 1 {
		t.Errorf("Expected 1 delivered message after retry, got %d", stats.Delivered)
	}
	if stats.Retried < 1 {
		t.Errorf("Expected retry count >= 1, got %d", stats.Retried)
	}

	// Call count should be > 1 due to retries
	callCount := mockSender.CallCount()
	if callCount <= 1 {
		t.Errorf("Expected multiple calls due to retries, got %d", callCount)
	}

	deliverer.Close()
}

func TestAsyncDeliverer_FlushAndClose(t *testing.T) {
	// Use a slow sender to test flush behavior
	mockSender := &MockSlowSender{Delay: 50 * time.Millisecond}

	config := DefaultConfig()
	config.Async = true
	config.QueueSize = 5
	config.Workers = 1
	config.ShutdownTimeout = 500 * time.Millisecond

	deliverer := NewAsyncDeliverer(mockSender, config)

	// Queue several messages
	for i := 0; i < 3; i++ {
		deliverer.Deliver("job", "info", "Message", []byte("test"), time.Now())
	}

	// Wait a bit to ensure at least some messages are processed
	// before we call Flush (but not all of them)
	time.Sleep(75 * time.Millisecond)

	// Flush should wait for all remaining queued messages
	start := time.Now()
	err := deliverer.Flush()
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Flush returned error: %v", err)
	}

	// Flush should have waited for messages to be processed
	if duration < 50*time.Millisecond {
		t.Errorf("Flush returned too quickly, expected to wait for processing")
	}

	// Wait a bit more for messages to be fully recorded in the mock
	time.Sleep(50 * time.Millisecond)

	calls := mockSender.Calls()
	if len(calls) != 3 {
		t.Errorf("Expected 3 messages to be processed after flush, got %d", len(calls))
	}

	// Test Close
	err = deliverer.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// After close, should not accept more messages
	err = deliverer.Deliver("job", "info", "After close", []byte("test"), time.Now())
	if err != nil {
		// This is expected, channel should be closed
	}
}

func TestAsyncDeliverer_MultipleWorkers(t *testing.T) {
	mockSender := &MockSlowSender{Delay: 20 * time.Millisecond}

	config := DefaultConfig()
	config.Async = true
	config.QueueSize = 10
	config.Workers = 3 // Multiple workers

	deliverer := NewAsyncDeliverer(mockSender, config)

	// Send multiple messages
	for i := 0; i < 6; i++ {
		err := deliverer.Deliver("job", "info", "Message", []byte("test"), time.Now())
		if err != nil {
			t.Errorf("Deliver returned error: %v", err)
		}
	}

	// With multiple workers, processing should be faster
	start := time.Now()

	// Give workers enough time to start processing before flush
	time.Sleep(10 * time.Millisecond)

	// Now flush and ensure all messages are processed
	err := deliverer.Flush()
	if err != nil {
		t.Errorf("Flush returned error: %v", err)
	}

	duration := time.Since(start)

	// With 3 workers and 6 messages at 20ms each, should take about 40ms (2 batches)
	// But give a little extra time for safety
	if duration > 100*time.Millisecond {
		t.Errorf("Multiple workers should process faster, took %v", duration)
	}

	// Give a bit more time for mockSender to record all calls
	time.Sleep(50 * time.Millisecond)

	calls := mockSender.Calls()
	if len(calls) != 6 {
		t.Errorf("Expected 6 messages to be processed, got %d", len(calls))
	}

	deliverer.Close()
}

func TestAsyncDeliverer_PermanentFailure(t *testing.T) {
	// Sender that will always fail
	mockSender := &MockSlowSender{ShouldError: true}

	config := DefaultConfig()
	config.Async = true
	config.MaxRetries = 2
	config.RetryInterval = 10 * time.Millisecond

	deliverer := NewAsyncDeliverer(mockSender, config)

	// Send a message that will ultimately fail
	err := deliverer.Deliver("fail-job", "error", "Will fail", []byte("test"), time.Now())
	if err != nil {
		t.Errorf("Deliver should not return error, got: %v", err)
	}

	// Allow time for retries and processing
	time.Sleep(100 * time.Millisecond)

	// Should eventually be marked as failed
	stats := deliverer.Status()
	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed message, got %d", stats.Failed)
	}

	// Should have retried according to config
	callCount := mockSender.CallCount()
	expectedCalls := 1 + config.MaxRetries // Initial try + retries
	if callCount != expectedCalls {
		t.Errorf("Expected %d calls due to retries, got %d", expectedCalls, callCount)
	}

	deliverer.Close()
}
