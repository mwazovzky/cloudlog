package cloudlog

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	clouderrors "github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/stretchr/testify/assert"
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
	client := NewClient("http://test", "user", "token", &http.Client{})

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

	var result map[string]interface{}
	if err := json.Unmarshal(mockClient.LastFormatted, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["version"] != "1.0" {
		t.Errorf("Expected version=1.0, got %v", result["version"])
	}
}

func TestErrorTypeFunctions(t *testing.T) {
	connectionErr := clouderrors.ConnectionError(errors.New("network down"), "failed to connect")
	if !IsConnectionError(connectionErr) {
		t.Error("IsConnectionError should return true for connection errors")
	}

	responseErr := clouderrors.ResponseError(403, "forbidden")
	if !IsResponseError(responseErr) {
		t.Error("IsResponseError should return true for response errors")
	}

	formatErr := clouderrors.FormatError(errors.New("bad format"), "invalid format")
	if !IsFormatError(formatErr) {
		t.Error("IsFormatError should return true for format errors")
	}

	inputErr := clouderrors.InputError("missing required field")
	if !IsInputError(inputErr) {
		t.Error("IsInputError should return true for input errors")
	}

	genericErr := errors.New("generic error")
	if IsConnectionError(genericErr) || IsResponseError(genericErr) ||
		IsFormatError(genericErr) || IsInputError(genericErr) {
		t.Error("Error type checks should return false for unrelated errors")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	url := "http://loki.example.com"
	user := "test-user"
	token := "test-token"
	httpClient := &http.Client{}

	client := NewClientWithOptions(url, user, token, httpClient)

	assert.NotNil(t, client)
}

func TestIsError(t *testing.T) {
	cloudlErr := clouderrors.InputError("test error")
	assert.True(t, IsError(cloudlErr), "IsError should return true for custom cloudlog errors")
	assert.False(t, IsError(nil), "IsError should return false for nil error")
}
