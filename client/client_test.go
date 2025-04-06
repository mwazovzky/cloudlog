package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	clouderrors "github.com/mwazovzky/cloudlog/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockHTTPClient implements the Doer interface for testing.
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	resp := args.Get(0)
	if resp == nil {
		return nil, args.Error(1)
	}
	return resp.(*http.Response), args.Error(1)
}

func TestLokiClient_Send_Success(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	mockResponse := &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	err := client.Send("test-job", []byte(`{"message":"test log"}`))
	assert.NoError(t, err, "Expected no error for successful log send")
	mockHTTPClient.AssertExpectations(t)
}

func TestLokiClient_Send_ResponseError(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	mockResponse := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewBufferString("internal server error")),
	}
	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	err := client.Send("test-job", []byte(`{"message":"test log"}`))

	// Test the error is present
	assert.Error(t, err, "Expected error for failed log send")

	// Test it's the right type of error
	assert.True(t, clouderrors.Is(err, clouderrors.ErrResponseError),
		"Expected a response error")

	// Test the error message contains relevant details
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "internal server error")

	mockHTTPClient.AssertExpectations(t)
}

func TestLokiClient_Send_ConnectionError(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	networkErr := errors.New("network error")
	mockHTTPClient.On("Do", mock.Anything).Return(nil, networkErr)

	client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	err := client.Send("test-job", []byte(`{"message":"test log"}`))

	// Test the error is present
	assert.Error(t, err, "Expected error for network failure")

	// Test it's the right type of error
	assert.True(t, clouderrors.Is(err, clouderrors.ErrConnectionFailed),
		"Expected a connection error")

	// Test the error message contains relevant details
	assert.Contains(t, err.Error(), "network error")

	mockHTTPClient.AssertExpectations(t)
}

func TestLokiClient_Send_ValidPayloadFormat(t *testing.T) {
	// Create a mock HTTP client that captures the request body
	mockHTTPClient := new(MockHTTPClient)

	// Create a response with success status
	mockResponse := &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}

	// Set up the mock to capture the request
	var capturedRequest *http.Request
	mockHTTPClient.On("Do", mock.Anything).Run(func(args mock.Arguments) { // Fixed missing comma here
		capturedRequest = args.Get(0).(*http.Request)
	}).Return(mockResponse, nil)

	// Create the client and send a log
	client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)
	testLogData := []byte(`{"message":"test log message","level":"info"}`)

	err := client.Send("test-job", testLogData)
	assert.NoError(t, err)

	// Verify the request was captured and the Do method was called
	require.NotNil(t, capturedRequest, "HTTP request was not captured")
	mockHTTPClient.AssertExpectations(t)

	// Read and parse the request body
	body, err := io.ReadAll(capturedRequest.Body)
	require.NoError(t, err, "Failed to read request body")

	// Parse the JSON payload
	var payload map[string]interface{}
	err = json.Unmarshal(body, &payload)
	require.NoError(t, err, "Failed to parse JSON payload")

	// Validate the payload structure
	streams, ok := payload["streams"].([]interface{})
	require.True(t, ok, "Payload should contain 'streams' array")
	require.Len(t, streams, 1, "Streams array should have 1 element")

	// Validate the stream object
	stream, ok := streams[0].(map[string]interface{})
	require.True(t, ok, "Stream should be an object")

	// Validate stream labels
	streamLabels, ok := stream["stream"].(map[string]interface{})
	require.True(t, ok, "Stream should have 'stream' labels object")
	assert.Equal(t, "test-job", streamLabels["job"], "Stream should have correct job label")

	// Validate stream values
	values, ok := stream["values"].([]interface{})
	require.True(t, ok, "Stream should have 'values' array")
	require.Len(t, values, 1, "Values array should have 1 element")

	// Validate the value entry (timestamp and message)
	valueEntry, ok := values[0].([]interface{})
	require.True(t, ok, "Value entry should be an array")
	require.Len(t, valueEntry, 2, "Value entry should have timestamp and message")

	// Timestamp should be a string of numbers (nanoseconds timestamp)
	timestamp, ok := valueEntry[0].(string)
	require.True(t, ok, "Timestamp should be a string")
	require.True(t, isNumericString(timestamp), "Timestamp should be numeric")

	// Message should match our original log data
	message, ok := valueEntry[1].(string)
	require.True(t, ok, "Message should be a string")
	assert.Equal(t, string(testLogData), message, "Message should match original log data")

	// Verify headers
	assert.Equal(t, "application/json", capturedRequest.Header.Get("Content-Type"))
}

// Helper function to check if string contains only digits
func isNumericString(s string) bool {
	return len(s) > 0 && strings.Trim(s, "0123456789") == ""
}
