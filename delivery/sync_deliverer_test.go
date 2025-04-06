package delivery

import (
	"errors"
	"testing"
	"time"
)

// MockSender implements client.LogSender for testing
type MockSender struct {
	LastJob       string
	LastFormatted []byte
	ShouldError   bool
	CallCount     int
}

func (m *MockSender) Send(job string, formatted []byte) error {
	m.CallCount++
	m.LastJob = job
	m.LastFormatted = formatted

	if m.ShouldError {
		return errors.New("mock error")
	}
	return nil
}

func TestSyncDeliverer_Deliver(t *testing.T) {
	// Test successful delivery
	t.Run("Successful delivery", func(t *testing.T) {
		mockSender := &MockSender{}
		deliverer := NewSyncDeliverer(mockSender)

		err := deliverer.Deliver(
			"test-job",
			"info",
			"Test message",
			[]byte(`{"message":"Test message"}`),
			time.Now(),
		)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if mockSender.CallCount != 1 {
			t.Errorf("Expected 1 call to Send, got %d", mockSender.CallCount)
		}

		if mockSender.LastJob != "test-job" {
			t.Errorf("Expected job 'test-job', got '%s'", mockSender.LastJob)
		}

		expected := `{"message":"Test message"}`
		if string(mockSender.LastFormatted) != expected {
			t.Errorf("Expected formatted message '%s', got '%s'", expected, string(mockSender.LastFormatted))
		}

		// Check stats
		stats := deliverer.Status()
		if stats.Delivered != 1 {
			t.Errorf("Expected 1 delivered message, got %d", stats.Delivered)
		}
		if stats.Failed != 0 {
			t.Errorf("Expected 0 failed messages, got %d", stats.Failed)
		}
	})

	// Test failed delivery
	t.Run("Failed delivery", func(t *testing.T) {
		mockSender := &MockSender{ShouldError: true}
		deliverer := NewSyncDeliverer(mockSender)

		err := deliverer.Deliver(
			"test-job",
			"info",
			"Test message",
			[]byte(`{"message":"Test message"}`),
			time.Now(),
		)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		// Check stats
		stats := deliverer.Status()
		if stats.Delivered != 0 {
			t.Errorf("Expected 0 delivered messages, got %d", stats.Delivered)
		}
		if stats.Failed != 1 {
			t.Errorf("Expected 1 failed message, got %d", stats.Failed)
		}
	})
}

func TestSyncDeliverer_FlushAndClose(t *testing.T) {
	mockSender := &MockSender{}
	deliverer := NewSyncDeliverer(mockSender)

	// Both should be no-ops and return nil for sync deliverer
	err1 := deliverer.Flush()
	if err1 != nil {
		t.Errorf("Expected nil from Flush(), got: %v", err1)
	}

	err2 := deliverer.Close()
	if err2 != nil {
		t.Errorf("Expected nil from Close(), got: %v", err2)
	}
}
