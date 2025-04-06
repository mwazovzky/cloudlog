package logger

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	clouderrors "github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
)

// MockClient is a mock implementation of the client.Interface for testing.
type MockClient struct {
	LastJob       string
	LastFormatted []byte
	ShouldError   bool
	ErrorType     string // "connection" or "response"
}

func (m *MockClient) Send(job string, formatted []byte) error {
	if m.ShouldError {
		switch m.ErrorType {
		case "connection":
			return clouderrors.ConnectionError(errors.New("mock error"), "failed to connect")
		case "response":
			return clouderrors.ResponseError(500, "server error")
		default:
			return errors.New("mock error")
		}
	}

	m.LastJob = job
	m.LastFormatted = formatted
	return nil
}

// MockFormatter is a formatter that will fail on demand
type MockFormatter struct {
	ShouldError bool
}

func (m *MockFormatter) Format(entry formatter.LogEntry) ([]byte, error) {
	if m.ShouldError {
		return nil, errors.New("format error")
	}
	data, _ := json.Marshal(entry.KeyVals)
	return data, nil
}

func TestLogger_Info(t *testing.T) {
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

func TestLogger_WithFormatter(t *testing.T) {
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

func TestLogger_WithContext(t *testing.T) {
	mockClient := &MockClient{}
	logger := New(mockClient)

	// Create logger with context
	contextLogger := logger.WithContext("requestID", "123", "userID", "456")

	err := contextLogger.Info("With context")
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}

	// Parse the JSON to verify content
	var result map[string]interface{}
	if err := json.Unmarshal(mockClient.LastFormatted, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check context was included
	if result["requestID"] != "123" {
		t.Errorf("Expected requestID=123, got %v", result["requestID"])
	}

	if result["userID"] != "456" {
		t.Errorf("Expected userID=456, got %v", result["userID"])
	}
}

func TestLogger_WithJob(t *testing.T) {
	mockClient := &MockClient{}
	logger := New(mockClient)

	// Create logger with different job
	jobLogger := logger.WithJob("custom-job")

	err := jobLogger.Info("Custom job")
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}

	if mockClient.LastJob != "custom-job" {
		t.Errorf("Expected job 'custom-job', got %s", mockClient.LastJob)
	}
}

func TestLogger_CustomOptions(t *testing.T) {
	mockClient := &MockClient{}
	logger := New(
		mockClient,
		WithJob("custom-app"),
		WithMetadata("version", "1.0"),
		WithMetadata("environment", "staging"),
	)

	err := logger.Info("Test with options")
	if err != nil {
		t.Fatalf("Info returned error: %v", err)
	}

	if mockClient.LastJob != "custom-app" {
		t.Errorf("Expected job 'custom-app', got %s", mockClient.LastJob)
	}

	// Parse the JSON to verify content
	var result map[string]interface{}
	if err := json.Unmarshal(mockClient.LastFormatted, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check default metadata was included
	if result["version"] != "1.0" {
		t.Errorf("Expected version=1.0, got %v", result["version"])
	}

	if result["environment"] != "staging" {
		t.Errorf("Expected environment=staging, got %v", result["environment"])
	}
}

func TestLogger_ClientError(t *testing.T) {
	mockClient := &MockClient{ShouldError: true, ErrorType: "connection"}
	logger := New(mockClient)

	err := logger.Info("Test message")
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	// Check that error is properly wrapped
	if !clouderrors.Is(err, clouderrors.ErrConnectionFailed) {
		t.Errorf("Expected a connection error, got: %v", err)
	}
}

func TestLogger_FormatterError(t *testing.T) {
	mockClient := &MockClient{}
	mockFormatter := &MockFormatter{ShouldError: true}
	logger := New(mockClient, WithFormatter(mockFormatter))

	err := logger.Info("Test message")
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	// Check that error is properly wrapped
	if !strings.Contains(err.Error(), "format log entry") {
		t.Errorf("Expected error message to mention formatting, got: %v", err)
	}
}
