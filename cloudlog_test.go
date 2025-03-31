package cloudlog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Client interface.
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Send(payload LogEntry) error {
	args := m.Called(payload)
	return args.Error(0)
}

func TestLogger_Info_Success(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(nil)

	logger := NewLogger(mockClient, "test-service") // Pass mockClient as the interface
	err := logger.Info("test-job", "key1", "value1", "key2", "value2")
	assert.NoError(t, err, "Expected no error when logging info")
	mockClient.AssertExpectations(t)
}

func TestLogger_Error_Success(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(nil)

	logger := NewLogger(mockClient, "test-service") // Pass mockClient as the interface
	err := logger.Error("test-job", "key1", "value1", "key2", "value2")
	assert.NoError(t, err, "Expected no error when logging error")
	mockClient.AssertExpectations(t)
}

func TestLogger_Debug_Success(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(nil)

	logger := NewLogger(mockClient, "test-service") // Pass mockClient as the interface
	err := logger.Debug("test-job", "key1", "value1", "key2", "value2")
	assert.NoError(t, err, "Expected no error when logging debug")
	mockClient.AssertExpectations(t)
}

func TestLogger_InvalidKeyValues(t *testing.T) {
	mockClient := new(MockClient)
	logger := NewLogger(mockClient, "test-service") // Pass mockClient as the interface

	// Test with an odd number of keyValues
	err := logger.Info("test-job", "key1", "value1", "key2") // Odd number of keyValues
	assert.Error(t, err, "Expected error when keyValues are not in pairs")
}

func TestLogger_ClientError(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(errors.New("mock error"))

	logger := NewLogger(mockClient, "test-service") // Pass mockClient as the interface
	err := logger.Info("test-job", "key1", "value1")
	assert.Error(t, err, "Expected error when client fails to send payload")
	mockClient.AssertExpectations(t)
}

func TestLogger_Log_StringArguments(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(nil)

	logger := NewLogger(mockClient, "test-service")

	// Call the Log method with string arguments
	err := logger.Log("info", "Test message", "key1", "value1")
	assert.NoError(t, err, "Expected no error when logging with string arguments")

	// Capture the payload passed to the mock client
	mockClient.AssertCalled(t, "Send", mock.MatchedBy(func(payload LogEntry) bool {
		// Verify the structure of the LogEntry
		assert.Len(t, payload.Streams, 1, "Expected one stream in the payload")
		stream := payload.Streams[0]

		// Verify the stream metadata
		assert.Equal(t, "test-service", stream.Stream["service_name"], "Expected service_name to match")
		assert.Equal(t, "info", stream.Stream["level"], "Expected level to match")
		assert.Equal(t, "value1", stream.Stream["key1"], "Expected key1 to match")

		// Verify the log message
		assert.Len(t, stream.Values, 1, "Expected one log entry in the values")
		assert.Equal(t, "Test message", stream.Values[0][1], "Expected log message to match")

		return true
	}))
}

func TestLogger_Log_IntArguments(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(nil)

	logger := NewLogger(mockClient, "test-service")

	// Call the Log method with integer arguments
	err := logger.Log("info", "Test message", "key1", 42)
	assert.NoError(t, err, "Expected no error when logging with integer arguments")

	// Capture the payload passed to the mock client
	mockClient.AssertCalled(t, "Send", mock.MatchedBy(func(payload LogEntry) bool {
		// Verify the structure of the LogEntry
		assert.Len(t, payload.Streams, 1, "Expected one stream in the payload")
		stream := payload.Streams[0]

		// Verify the stream metadata
		assert.Equal(t, "test-service", stream.Stream["service_name"], "Expected service_name to match")
		assert.Equal(t, "info", stream.Stream["level"], "Expected level to match")
		assert.Equal(t, "42", stream.Stream["key1"], "Expected key1 to be converted to string")

		// Verify the log message
		assert.Len(t, stream.Values, 1, "Expected one log entry in the values")
		assert.Equal(t, "Test message", stream.Values[0][1], "Expected log message to match")

		return true
	}))
}

func TestLogger_Log_FloatArguments(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(nil)

	logger := NewLogger(mockClient, "test-service")

	// Call the Log method with float arguments
	err := logger.Log("info", "Test message", "key1", 3.14)
	assert.NoError(t, err, "Expected no error when logging with float arguments")

	// Capture the payload passed to the mock client
	mockClient.AssertCalled(t, "Send", mock.MatchedBy(func(payload LogEntry) bool {
		// Verify the structure of the LogEntry
		assert.Len(t, payload.Streams, 1, "Expected one stream in the payload")
		stream := payload.Streams[0]

		// Verify the stream metadata
		assert.Equal(t, "test-service", stream.Stream["service_name"], "Expected service_name to match")
		assert.Equal(t, "info", stream.Stream["level"], "Expected level to match")
		assert.Equal(t, "3.14", stream.Stream["key1"], "Expected key1 to be converted to string")

		// Verify the log message
		assert.Len(t, stream.Values, 1, "Expected one log entry in the values")
		assert.Equal(t, "Test message", stream.Values[0][1], "Expected log message to match")

		return true
	}))
}

type TestStruct struct {
	Field1 string
	Field2 int
}

func TestLogger_Log_StructArgument(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(nil)

	logger := NewLogger(mockClient, "test-service")

	// Create a struct to pass as a parameter
	testStruct := TestStruct{
		Field1: "value1",
		Field2: 42,
	}

	// Call the Log method with a struct argument
	err := logger.Log("info", "Test message", "key1", testStruct)
	assert.NoError(t, err, "Expected no error when logging with a struct argument")

	// Capture the payload passed to the mock client
	mockClient.AssertCalled(t, "Send", mock.MatchedBy(func(payload LogEntry) bool {
		// Verify the structure of the LogEntry
		assert.Len(t, payload.Streams, 1, "Expected one stream in the payload")
		stream := payload.Streams[0]

		// Verify the stream metadata
		assert.Equal(t, "test-service", stream.Stream["service_name"], "Expected service_name to match")
		assert.Equal(t, "info", stream.Stream["level"], "Expected level to match")
		assert.Equal(t, "{Field1:value1 Field2:42}", stream.Stream["key1"], "Expected key1 to match the exact string representation of the struct")

		// Verify the log message
		assert.Len(t, stream.Values, 1, "Expected one log entry in the values")
		assert.Equal(t, "Test message", stream.Values[0][1], "Expected log message to match")

		return true
	}))
}

func TestLogger_Log_SliceArgument(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.On("Send", mock.Anything).Return(nil)

	logger := NewLogger(mockClient, "test-service")

	// Create a slice of strings to pass as a parameter
	testSlice := []string{"value1", "value2", "value3"}

	// Call the Log method with a slice argument
	err := logger.Log("info", "Test message", "key1", testSlice)
	assert.NoError(t, err, "Expected no error when logging with a slice argument")

	// Capture the payload passed to the mock client
	mockClient.AssertCalled(t, "Send", mock.MatchedBy(func(payload LogEntry) bool {
		// Verify the structure of the LogEntry
		assert.Len(t, payload.Streams, 1, "Expected one stream in the payload")
		stream := payload.Streams[0]

		// Verify the stream metadata
		assert.Equal(t, "test-service", stream.Stream["service_name"], "Expected service_name to match")
		assert.Equal(t, "info", stream.Stream["level"], "Expected level to match")
		assert.Equal(t, "[value1 value2 value3]", stream.Stream["key1"], "Expected key1 to match the string representation of the slice")

		// Verify the log message
		assert.Len(t, stream.Values, 1, "Expected one log entry in the values")
		assert.Equal(t, "Test message", stream.Values[0][1], "Expected log message to match")

		return true
	}))
}
