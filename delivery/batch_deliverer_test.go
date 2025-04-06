package delivery

import (
	"testing"
	"time"
)

func TestBatchDeliverer_BasicOperation(t *testing.T) {
	mockSender := &MockSlowSender{}

	config := DefaultConfig()
	config.Async = true
	config.BatchSize = 3
	config.FlushInterval = 100 * time.Millisecond

	deliverer := NewBatchDeliverer(mockSender, config)

	// Deliver some messages
	deliverer.Deliver("job1", "info", "Message 1", []byte("test1"), time.Now())
	deliverer.Deliver("job2", "info", "Message 2", []byte("test2"), time.Now())

	// These should not be sent yet since batch size is 3
	time.Sleep(10 * time.Millisecond)
	if mockSender.CallCount() > 0 {
		t.Errorf("Messages should not be sent until batch size reached or flush interval")
	}

	// Add one more message to trigger batch send
	deliverer.Deliver("job3", "info", "Message 3", []byte("test3"), time.Now())

	// Wait a bit for the batch to be processed
	time.Sleep(50 * time.Millisecond)

	calls := mockSender.Calls()
	if len(calls) != 3 {
		t.Errorf("Expected 3 calls after batch size reached, got %d", len(calls))
	}

	// Clean up
	deliverer.Close()
}

func TestBatchDeliverer_FlushInterval(t *testing.T) {
	mockSender := &MockSlowSender{}

	config := DefaultConfig()
	config.Async = true
	config.BatchSize = 10                        // Large batch size
	config.FlushInterval = 50 * time.Millisecond // Short flush interval

	deliverer := NewBatchDeliverer(mockSender, config)

	// Deliver some messages
	deliverer.Deliver("job1", "info", "Message 1", []byte("test1"), time.Now())
	deliverer.Deliver("job2", "info", "Message 2", []byte("test2"), time.Now())

	// These should not be sent immediately
	time.Sleep(10 * time.Millisecond)
	if mockSender.CallCount() > 0 {
		t.Errorf("Messages should not be sent until flush interval elapsed")
	}

	// Wait for flush interval
	time.Sleep(60 * time.Millisecond)

	// Messages should be sent now due to flush interval
	calls := mockSender.Calls()
	if len(calls) != 2 {
		t.Errorf("Expected 2 calls after flush interval, got %d", len(calls))
	}

	// Clean up
	deliverer.Close()
}

func TestBatchDeliverer_ManualFlush(t *testing.T) {
	mockSender := &MockSlowSender{}

	config := DefaultConfig()
	config.Async = true
	config.BatchSize = 10                          // Large batch size
	config.FlushInterval = 1000 * time.Millisecond // Long flush interval

	deliverer := NewBatchDeliverer(mockSender, config)

	// Deliver some messages
	deliverer.Deliver("job1", "info", "Message 1", []byte("test1"), time.Now())
	deliverer.Deliver("job2", "info", "Message 2", []byte("test2"), time.Now())

	// Manually flush
	err := deliverer.Flush()
	if err != nil {
		t.Errorf("Flush returned error: %v", err)
	}

	// Allow time for flush to complete
	time.Sleep(50 * time.Millisecond)

	// Messages should be sent now due to manual flush
	calls := mockSender.Calls()
	if len(calls) != 2 {
		t.Errorf("Expected 2 calls after manual flush, got %d", len(calls))
	}

	// Clean up
	deliverer.Close()
}

func TestBatchDeliverer_CloseWithPendingMessages(t *testing.T) {
	mockSender := &MockSlowSender{}

	config := DefaultConfig()
	config.Async = true
	config.BatchSize = 10
	config.FlushInterval = 1000 * time.Millisecond
	// Set shorter shutdown timeout to not slow down tests too much
	config.ShutdownTimeout = 300 * time.Millisecond

	deliverer := NewBatchDeliverer(mockSender, config)

	// Deliver some messages
	deliverer.Deliver("job1", "info", "Message 1", []byte("test1"), time.Now())
	deliverer.Deliver("job2", "info", "Message 2", []byte("test2"), time.Now())

	// Give some time for messages to be queued properly
	time.Sleep(10 * time.Millisecond)

	// Close should first flush pending messages and then shut down
	err := deliverer.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Wait a bit after close to allow for processing
	time.Sleep(100 * time.Millisecond)

	// Check that all messages were sent
	calls := mockSender.Calls()
	if len(calls) != 2 {
		t.Errorf("Expected 2 calls after close with pending messages, got %d", len(calls))
		// Print details to help with debugging
		for i, call := range calls {
			t.Logf("Call %d: job=%s", i, call.Job)
		}
	}
}
