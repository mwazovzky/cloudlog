package logger

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	clouderrors "github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
	cloudtesting "github.com/mwazovzky/cloudlog/testing"
	"github.com/stretchr/testify/assert"
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

func TestErrorDebugWarn(t *testing.T) {
	// Test Error, Debug, and Warn methods
	testClient := cloudtesting.NewTestLogger()
	logger := New(testClient)

	// Test Error method
	err := logger.Error("error message", "key", "value")
	assert.NoError(t, err)
	assert.True(t, testClient.ContainsEntry("error", "error message"))

	// Clear logs before next test
	testClient.Clear()

	// Test Debug method
	err = logger.Debug("debug message", "key", "value")
	assert.NoError(t, err)
	assert.True(t, testClient.ContainsEntry("debug", "debug message"))

	// Clear logs before next test
	testClient.Clear()

	// Test Warn method
	err = logger.Warn("warn message", "key", "value")
	assert.NoError(t, err)
	assert.True(t, testClient.ContainsEntry("warn", "warn message"))
}

func TestWithContextEdgeCases(t *testing.T) {
	// Test WithContext with additional edge cases
	testClient := cloudtesting.NewTestLogger()
	logger := New(testClient)

	// Test with nil context
	contextLogger := logger.WithContext(nil)
	assert.Equal(t, logger, contextLogger)

	// Test with empty context
	emptyCtx := context.Background()
	contextLogger = logger.WithContext(emptyCtx)
	// The behavior should be the same as the original logger
	// when no values are in the context
	assert.NotNil(t, contextLogger)
}

func TestWithJobVariations(t *testing.T) {
	testClient := cloudtesting.NewTestLogger()
	logger := New(testClient, WithJob("application"))

	// Verify job name is set correctly
	assert.Equal(t, "application", logger.job)

	// Test with empty job name
	jobLogger := logger.WithJob("")
	assert.Equal(t, "", jobLogger.job)

	// Test with non-empty job
	jobLogger = logger.WithJob("test-job")
	assert.Equal(t, "test-job", jobLogger.job)
}

func TestWithContextComplexScenarios(t *testing.T) {
	testClient := cloudtesting.NewTestLogger()
	logger := New(testClient)

	contextLogger := logger.WithContext("key1", "value1", "key2", 42)
	assert.NotNil(t, contextLogger)

	err := contextLogger.Info("Test with context values")
	assert.NoError(t, err)

	logs := testClient.Logs()
	assert.Equal(t, 1, len(logs))

	data := logs[0].Data
	assert.Equal(t, "value1", data["key1"], "Context key1 should be in the log with correct value")

	numVal, ok := data["key2"].(float64)
	assert.True(t, ok, "key2 should be a numeric value")
	assert.Equal(t, float64(42), numVal, "Context key2 should be in the log with correct value")
}

func TestWithJobAdditionalCases(t *testing.T) {
	// Test additional cases for WithJob
	testClient := cloudtesting.NewTestLogger()

	// Test the default job (typically "application")
	logger := New(testClient)

	// Test the job inheritance in method calls
	err := logger.Info("Test with default job")
	assert.NoError(t, err)

	logs := testClient.Logs()
	assert.Equal(t, 1, len(logs))
	assert.True(t, testClient.ContainsEntry("info", "Test with default job"))

	testClient.Clear()

	// Test with special characters in job name
	specialJobLogger := logger.WithJob("special@job#123")
	err = specialJobLogger.Info("Test with special job name")
	assert.NoError(t, err)

	logs = testClient.Logs()
	assert.Equal(t, 1, len(logs))
	assert.True(t, testClient.ContainsEntry("info", "Test with special job name"))
}

// TestBackwardCompatibility verifies the logger works the same way with delivery layer
func TestBackwardCompatibility(t *testing.T) {
	mockClient := &MockClient{}
	logger := New(mockClient) // Using the updated New that wraps with SyncDeliverer

	// Test the basic functionality
	err := logger.Info("Test message", "key1", "value1")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if mockClient.LastJob != "application" {
		t.Errorf("Expected job 'application', got: %s", mockClient.LastJob)
	}

	// Verify the message format hasn't changed
	var parsed map[string]interface{}
	err = json.Unmarshal(mockClient.LastFormatted, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["level"] != "info" {
		t.Errorf("Expected level 'info', got: %v", parsed["level"])
	}

	if parsed["message"] != "Test message" {
		t.Errorf("Expected message 'Test message', got: %v", parsed["message"])
	}

	if parsed["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got: %v", parsed["key1"])
	}
}

// TestErrorPropagation tests that client errors are still propagated correctly
func TestErrorPropagation(t *testing.T) {
	mockClient := &MockClient{ShouldError: true}
	logger := New(mockClient)

	err := logger.Info("Test message")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestWithCustomFormatter verifies formatter options still work
func TestWithCustomFormatter(t *testing.T) {
	mockClient := &MockClient{}
	stringFormatter := formatter.NewStringFormatter()
	logger := New(mockClient, WithFormatter(stringFormatter))

	err := logger.Info("Test message", "key1", "value1")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// String formatter produces output like "timestamp=2023-01-01T12:34:56Z job=application level=info message=Test message key1=value1"
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
