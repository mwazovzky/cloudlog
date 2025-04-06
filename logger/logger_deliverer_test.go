package logger

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/mwazovzky/cloudlog/delivery"
)

// MockDeliverer implements delivery.LogDeliverer for testing
type MockDeliverer struct {
	LastJob       string
	LastLevel     string
	LastMessage   string
	LastFormatted []byte
	LastTimestamp time.Time
	ShouldError   bool
	CallCount     int
	CloseCount    int
	FlushCount    int
}

func (m *MockDeliverer) Deliver(job, level, message string, formatted []byte, timestamp time.Time) error {
	m.CallCount++
	m.LastJob = job
	m.LastLevel = level
	m.LastMessage = message
	m.LastFormatted = formatted
	m.LastTimestamp = timestamp

	if m.ShouldError {
		return errors.New("mock deliverer error")
	}
	return nil
}

func (m *MockDeliverer) Flush() error {
	m.FlushCount++
	if m.ShouldError {
		return errors.New("mock flush error")
	}
	return nil
}

func (m *MockDeliverer) Close() error {
	m.CloseCount++
	if m.ShouldError {
		return errors.New("mock close error")
	}
	return nil
}

func (m *MockDeliverer) Status() delivery.DeliveryStatus {
	return delivery.DeliveryStatus{
		Delivered: m.CallCount,
	}
}

func TestLoggerWithDeliverer(t *testing.T) {
	mockDeliverer := &MockDeliverer{}
	logger := NewWithDeliverer(mockDeliverer)

	// Test info log
	err := logger.Info("Test message", "key", "value")
	if err != nil {
		t.Errorf("Info returned error: %v", err)
	}

	if mockDeliverer.CallCount != 1 {
		t.Errorf("Expected 1 call to Deliver, got %d", mockDeliverer.CallCount)
	}

	if mockDeliverer.LastJob != "application" {
		t.Errorf("Expected job 'application', got '%s'", mockDeliverer.LastJob)
	}

	if mockDeliverer.LastLevel != "info" {
		t.Errorf("Expected level 'info', got '%s'", mockDeliverer.LastLevel)
	}

	if mockDeliverer.LastMessage != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", mockDeliverer.LastMessage)
	}

	// Check formatted JSON
	var result map[string]interface{}
	if err := json.Unmarshal(mockDeliverer.LastFormatted, &result); err != nil {
		t.Fatalf("Failed to parse formatted JSON: %v", err)
	}

	if result["level"] != "info" {
		t.Errorf("Expected level 'info', got '%v'", result["level"])
	}

	if result["message"] != "Test message" {
		t.Errorf("Expected message 'Test message', got '%v'", result["message"])
	}

	if result["key"] != "value" {
		t.Errorf("Expected key 'key' with value 'value', got '%v'", result["key"])
	}
}

func TestLoggerWithDelivererError(t *testing.T) {
	mockDeliverer := &MockDeliverer{ShouldError: true}
	logger := NewWithDeliverer(mockDeliverer)

	// Test error when delivering
	err := logger.Info("Test message")
	if err == nil {
		t.Error("Expected error from Info(), got nil")
	}
}

func TestLoggerFlushAndClose(t *testing.T) {
	mockDeliverer := &MockDeliverer{}
	logger := NewWithDeliverer(mockDeliverer)

	// Test Flush
	err1 := logger.Flush()
	if err1 != nil {
		t.Errorf("Flush returned error: %v", err1)
	}
	if mockDeliverer.FlushCount != 1 {
		t.Errorf("Expected 1 call to Flush, got %d", mockDeliverer.FlushCount)
	}

	// Test Close
	err2 := logger.Close()
	if err2 != nil {
		t.Errorf("Close returned error: %v", err2)
	}
	if mockDeliverer.CloseCount != 1 {
		t.Errorf("Expected 1 call to Close, got %d", mockDeliverer.CloseCount)
	}

	// Test error cases
	mockDeliverer.ShouldError = true

	err3 := logger.Flush()
	if err3 == nil {
		t.Error("Expected error from Flush(), got nil")
	}

	err4 := logger.Close()
	if err4 == nil {
		t.Error("Expected error from Close(), got nil")
	}
}

func TestLoggerWithContextWithDeliverer(t *testing.T) {
	mockDeliverer := &MockDeliverer{}
	logger := NewWithDeliverer(mockDeliverer)

	// Create logger with context
	contextLogger := logger.WithContext("request_id", "123")

	err := contextLogger.Info("With context")
	if err != nil {
		t.Errorf("Info returned error: %v", err)
	}

	// Parse the JSON to verify content
	var result map[string]interface{}
	if err := json.Unmarshal(mockDeliverer.LastFormatted, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check context was included
	if result["request_id"] != "123" {
		t.Errorf("Expected request_id=123, got %v", result["request_id"])
	}
}

func TestLoggerWithJobWithDeliverer(t *testing.T) {
	mockDeliverer := &MockDeliverer{}
	logger := NewWithDeliverer(mockDeliverer)

	// Create logger with different job
	jobLogger := logger.WithJob("custom-job")

	err := jobLogger.Info("Custom job")
	if err != nil {
		t.Errorf("Info returned error: %v", err)
	}

	if mockDeliverer.LastJob != "custom-job" {
		t.Errorf("Expected job 'custom-job', got '%s'", mockDeliverer.LastJob)
	}
}
