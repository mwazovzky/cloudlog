package delivery

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Async {
		t.Error("Default config should not be async")
	}

	if config.QueueSize != 1000 {
		t.Errorf("Expected QueueSize 1000, got %d", config.QueueSize)
	}

	if config.BatchSize != 100 {
		t.Errorf("Expected BatchSize 100, got %d", config.BatchSize)
	}

	if config.Workers != 1 {
		t.Errorf("Expected Workers 1, got %d", config.Workers)
	}

	if config.FlushInterval != time.Second*5 {
		t.Errorf("Expected FlushInterval 5s, got %v", config.FlushInterval)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}

	if config.RetryInterval != time.Second*1 {
		t.Errorf("Expected RetryInterval 1s, got %v", config.RetryInterval)
	}

	if config.ShutdownTimeout != time.Second*10 {
		t.Errorf("Expected ShutdownTimeout 10s, got %v", config.ShutdownTimeout)
	}
}

func TestDeliveryStatus(t *testing.T) {
	status := DeliveryStatus{
		Buffered:  10,
		Delivered: 100,
		Failed:    5,
		Dropped:   2,
		Retried:   8,
	}

	if status.Buffered != 10 {
		t.Errorf("Expected Buffered 10, got %d", status.Buffered)
	}

	if status.Delivered != 100 {
		t.Errorf("Expected Delivered 100, got %d", status.Delivered)
	}

	if status.Failed != 5 {
		t.Errorf("Expected Failed 5, got %d", status.Failed)
	}

	if status.Dropped != 2 {
		t.Errorf("Expected Dropped 2, got %d", status.Dropped)
	}

	if status.Retried != 8 {
		t.Errorf("Expected Retried 8, got %d", status.Retried)
	}
}
