package cloudlog

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	clouderrors "github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
)

// MockClient is a mock client implementation for testing
type MockClient struct {
	LastJob       string
	LastFormatted []byte
	ShouldError   bool
}

func (m *MockClient) Send(job string, formatted []byte) error {
	if m.ShouldError {
		return errors.New("mock error")
	}

	m.LastJob = job
	m.LastFormatted = formatted
	return nil
}

func TestNewClient(t *testing.T) {
	// This is a simple test to ensure the facade function works
	// The actual implementation is tested in the client package
	client := NewClient("http://test", "user", "token", &http.Client{})

	// Just check that we get a non-nil client
	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestNew(t *testing.T) {
	mockClient := &MockClient{}
	logger := New(mockClient)

	err := logger.Info("Test message", "key1", "value1", "key2", 42)
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}

	if mockClient.LastJob != "application" {
		t.Errorf("Expected job 'application', got %s", mockClient.LastJob)
	}

	// Parse the JSON to verify content
	var result map[string]interface{}
	if err := json.Unmarshal(mockClient.LastFormatted, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["level"] != "info" {
		t.Errorf("Expected level info, got %s", result["level"])
	}

	if result["message"] != "Test message" {
		t.Errorf("Expected message 'Test message', got %v", result["message"])
	}

	if result["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", result["key1"])
	}

	if result["key2"].(float64) != 42 {
		t.Errorf("Expected key2=42, got %v", result["key2"])
	}
}

func TestWithFormatter(t *testing.T) {
	mockClient := &MockClient{}
	stringFormatter := formatter.NewStringFormatter()
	logger := New(mockClient, WithFormatter(stringFormatter))

	err := logger.Info("Test message", "key1", "value1")
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}

	output := string(mockClient.LastFormatted)

	// Check that it contains expected parts
	if !strings.Contains(output, "job=application") {
		t.Errorf("Expected output to contain 'job=application', got: %s", output)
	}

	if !strings.Contains(output, "level=info") {
		t.Errorf("Expected output to contain 'level=info', got: %s", output)
	}

	if !strings.Contains(output, "message=Test message") {
		t.Errorf("Expected output to contain 'message=Test message', got: %s", output)
	}

	if !strings.Contains(output, "key1=value1") {
		t.Errorf("Expected output to contain 'key1=value1', got: %s", output)
	}
}

func TestWithJob(t *testing.T) {
	mockClient := &MockClient{}
	logger := New(mockClient, WithJob("custom-job"))

	err := logger.Info("Test message")
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}

	if mockClient.LastJob != "custom-job" {
		t.Errorf("Expected job 'custom-job', got %s", mockClient.LastJob)
	}
}

func TestWithMetadata(t *testing.T) {
	mockClient := &MockClient{}
	logger := New(mockClient, WithMetadata("version", "1.0"))

	err := logger.Info("Test message")
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}

	// Parse the JSON to verify content
	var result map[string]interface{}
	if err := json.Unmarshal(mockClient.LastFormatted, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["version"] != "1.0" {
		t.Errorf("Expected version=1.0, got %v", result["version"])
	}
}

func TestErrorTypeFunctions(t *testing.T) {
	// Test IsConnectionError
	connectionErr := clouderrors.ConnectionError(errors.New("network down"), "failed to connect")
	if !IsConnectionError(connectionErr) {
		t.Error("IsConnectionError should return true for connection errors")
	}

	// Test IsResponseError
	responseErr := clouderrors.ResponseError(403, "forbidden")
	if !IsResponseError(responseErr) {
		t.Error("IsResponseError should return true for response errors")
	}

	// Test IsFormatError
	formatErr := clouderrors.FormatError(errors.New("bad format"), "invalid format")
	if !IsFormatError(formatErr) {
		t.Error("IsFormatError should return true for format errors")
	}

	// Test IsInputError
	inputErr := clouderrors.InputError("missing required field")
	if !IsInputError(inputErr) {
		t.Error("IsInputError should return true for input errors")
	}

	// Test with generic error (should all return false)
	genericErr := errors.New("generic error")
	if IsConnectionError(genericErr) || IsResponseError(genericErr) ||
		IsFormatError(genericErr) || IsInputError(genericErr) {
		t.Error("Error type checks should return false for unrelated errors")
	}
}

// Tests for NewTestLogger and NewTestingLogger were removed as these functions
// were moved to the testing package
