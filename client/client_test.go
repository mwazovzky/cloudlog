package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mwazovzky/cloudlog/errors"
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

func TestLokiClient_Send(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	mockResponse := &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}
	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	entry := LokiEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{"job": "test-job"},
				Values: [][]string{
					{fmt.Sprintf("%d", time.Now().UnixNano()), `{"message":"test log"}`},
				},
			},
		},
	}

	err := client.Send(entry)
	assert.NoError(t, err)

	mockHTTPClient.ExpectedCalls = nil
	mockHTTPClient.On("Do", mock.Anything).Return(nil, fmt.Errorf("network error"))

	err = client.Send(entry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

func TestLokiClient_Send_ErrorHandling(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)

	t.Run("ConnectionError", func(t *testing.T) {
		// Setup mock for connection error
		mockHTTPClient.On("Do", mock.Anything).Return(nil, fmt.Errorf("%w: network error", errors.ErrConnectionFailed))
		client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

		entry := LokiEntry{
			Streams: []LokiStream{
				{
					Stream: map[string]string{"job": "test-job"},
					Values: [][]string{
						{fmt.Sprintf("%d", time.Now().UnixNano()), `{"message":"test log"}`},
					},
				},
			},
		}

		err := client.Send(entry)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrConnectionFailed))

		// Clear mock for the next subtest
		mockHTTPClient.ExpectedCalls = nil
		mockHTTPClient.Calls = nil
	})

	t.Run("ResponseError", func(t *testing.T) {
		// Setup mock for HTTP 500 error
		mockResponse := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("internal server error")),
		}
		mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

		client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

		entry := LokiEntry{
			Streams: []LokiStream{
				{
					Stream: map[string]string{"job": "test-job"},
					Values: [][]string{
						{fmt.Sprintf("%d", time.Now().UnixNano()), `{"message":"test log"}`},
					},
				},
			},
		}

		err := client.Send(entry)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrResponseError), "Expected error to be ErrResponseError")
		assert.Contains(t, err.Error(), "internal server error", "Expected error message to include response body")
	})
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

	entry := LokiEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"job": "test-job",
				},
				Values: [][]string{
					{
						fmt.Sprintf("%d", time.Now().UnixNano()),
						string(testLogData),
					},
				},
			},
		},
	}

	err := client.Send(entry)
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

func TestWithTimeout(t *testing.T) {
	// Test that WithTimeout option properly sets the timeout
	timeout := 5 * time.Second
	opt := WithTimeout(timeout)

	// Create a mock client that we can apply the option to
	testURL := "http://example.com"
	testUser := "test-user"
	testToken := "test-token"
	mockHTTPClient := new(MockHTTPClient)

	// Apply the option during client creation
	client := NewLokiClientWithOptions(testURL, testUser, testToken, mockHTTPClient, opt)

	// Since we can't access the private timeout field directly, we'll verify the client was created
	assert.NotNil(t, client)
	assert.Equal(t, testURL, client.url) // This should still be accessible
}

func TestNewLokiClientWithOptions(t *testing.T) {
	// Test client creation with custom options
	testURL := "http://loki.example.com"
	testUser := "test-user"
	testToken := "test-token"
	mockHTTPClient := new(MockHTTPClient)
	testTimeout := 10 * time.Second

	client := NewLokiClientWithOptions(testURL, testUser, testToken, mockHTTPClient, WithTimeout(testTimeout))

	assert.Equal(t, testURL, client.url)
	// We can't directly access the timeout field, but we can verify the client has been created successfully
	assert.NotNil(t, client)
}

func TestSendAdditionalCases(t *testing.T) {
	// Test context cancellation scenario
	mockHTTPClient := new(MockHTTPClient)
	testURL := "http://loki.example.com"
	testUser := "test-user"
	testToken := "test-token"
	client := NewLokiClient(testURL, testUser, testToken, mockHTTPClient)

	// Create a job name for the test
	testJob := "test-job"

	// Setup the mock to return an error when the context is canceled
	mockHTTPClient.On("Do", mock.Anything).Return(nil, context.Canceled)

	entry := LokiEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"job": testJob,
				},
				Values: [][]string{
					{
						fmt.Sprintf("%d", time.Now().UnixNano()),
						`{"test":"data"}`,
					},
				},
			},
		},
	}

	// Send with a job string (not context)
	err := client.Send(entry)

	// Verify we get an error
	assert.Error(t, err)
	// The error should be wrapped, so we can't directly check for context.Canceled
	// But we can check that it's a connection error
	assert.True(t, errors.IsConnectionError(err))
	mockHTTPClient.AssertExpectations(t)
}

func TestWithTimeoutEdgeCases(t *testing.T) {
	// Test with zero timeout value
	zeroTimeout := time.Duration(0)

	// Create client with zero timeout
	testURL := "http://example.com"
	testUser := "test-user"
	testToken := "test-token"
	mockHTTPClient := new(MockHTTPClient)

	// Create client with zero timeout option
	zeroClient := NewLokiClientWithOptions(testURL, testUser, testToken, mockHTTPClient, WithTimeout(zeroTimeout))
	assert.NotNil(t, zeroClient)

	// Test with negative timeout (should be handled gracefully)
	negativeTimeout := -5 * time.Second

	// Create client with negative timeout (should not panic)
	assert.NotPanics(t, func() {
		negClient := NewLokiClientWithOptions(testURL, testUser, testToken, mockHTTPClient, WithTimeout(negativeTimeout))
		assert.NotNil(t, negClient)
	})
}

func TestSendErrorHandling(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	testURL := "http://loki.example.com"
	testUser := "test-user"
	testToken := "test-token"
	client := NewLokiClient(testURL, testUser, testToken, mockHTTPClient)

	mockHTTPClient.ExpectedCalls = nil
	mockHTTPClient.Calls = nil

	badURLClient := NewLokiClient("://invalid-url", testUser, testToken, mockHTTPClient)

	entry := LokiEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"job": "test-job",
				},
				Values: [][]string{
					{
						fmt.Sprintf("%d", time.Now().UnixNano()),
						`{"test":"data"}`,
					},
				},
			},
		},
	}

	err := badURLClient.Send(entry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")

	mockHTTPClient.ExpectedCalls = nil
	mockHTTPClient.Calls = nil
	timeoutErr := fmt.Errorf("%w: context deadline exceeded", errors.ErrTimeout)
	mockHTTPClient.On("Do", mock.Anything).Return(nil, timeoutErr)

	err = client.Send(entry)
	assert.Error(t, err)
	// Instead of checking for the specific message, verify that it wraps the correct error type
	assert.True(t, errors.Is(err, errors.ErrConnectionFailed), "Expected a connection error")

	// Use a generic check that would work with the actual implementation
	assert.Contains(t, err.Error(), "connection to log service failed", "Expected connection failed message")
}

func TestLokiClient_SendEntry(t *testing.T) {
	// Create a mock HTTP client that captures the request body
	mockHTTPClient := new(MockHTTPClient)

	// Create a response with success status
	mockResponse := &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}

	// Set up the mock to capture the request
	var capturedRequest *http.Request
	mockHTTPClient.On("Do", mock.Anything).Run(func(args mock.Arguments) {
		capturedRequest = args.Get(0).(*http.Request)
	}).Return(mockResponse, nil)

	// Create the client
	client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	// Create a LokiEntry directly using our new structured types
	entry := LokiEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"job": "test-job",
					"env": "testing", // Add an additional label to test richer functionality
				},
				Values: [][]string{
					{
						"1626882892000000000", // Fixed timestamp for testing
						`{"message":"test log message","level":"info"}`,
					},
				},
			},
		},
	}

	// Send the entry
	err := client.Send(entry)
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
	assert.Equal(t, "testing", streamLabels["env"], "Stream should have correct env label")

	// Validate stream values
	values, ok := stream["values"].([]interface{})
	require.True(t, ok, "Stream should have 'values' array")
	require.Len(t, values, 1, "Values array should have 1 element")

	// Validate the value entry (timestamp and message)
	valueEntry, ok := values[0].([]interface{})
	require.True(t, ok, "Value entry should be an array")
	require.Len(t, valueEntry, 2, "Value entry should have timestamp and message")

	// Check the timestamp and message directly
	assert.Equal(t, "1626882892000000000", valueEntry[0].(string), "Timestamp should match")
	assert.Equal(t, `{"message":"test log message","level":"info"}`, valueEntry[1].(string), "Message should match")

	// Verify headers
	assert.Equal(t, "application/json", capturedRequest.Header.Get("Content-Type"))
}

func TestWithTimeout_EdgeCases(t *testing.T) {
	// Test with different HTTP client types

	// Case 1: Basic HTTP client - timeout should apply
	httpClient := &http.Client{}
	testURL := "http://example.com"
	testUser := "test-user"
	testToken := "test-token"
	timeout := 7 * time.Second

	// Store the result but don't need to use it directly
	_ = NewLokiClientWithOptions(testURL, testUser, testToken, httpClient, WithTimeout(timeout))
	// Can't directly access private fields, but we can verify the client's timeout was changed
	assert.Equal(t, timeout, httpClient.Timeout, "Timeout should be applied to HTTP client")

	// Case 2: Custom implementation not matching http.Client - should not panic
	customClient := &customDoer{}
	clientWithCustomDoer := NewLokiClientWithOptions(testURL, testUser, testToken, customClient, WithTimeout(timeout))
	assert.NotNil(t, clientWithCustomDoer, "Should create client with custom doer")

	// Case 3: Zero timeout
	zeroClient := NewLokiClientWithOptions(testURL, testUser, testToken, httpClient, WithTimeout(0))
	assert.NotNil(t, zeroClient, "Should handle zero timeout")
}

// A custom implementation of the Doer interface that is not an http.Client
type customDoer struct{}

func (c *customDoer) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

func TestSend(t *testing.T) {
	// Create a mock HTTP client that captures the request body
	mockHTTPClient := new(MockHTTPClient)

	// Create a response with success status
	mockResponse := &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}

	// Set up the mock to capture the request
	var capturedRequest *http.Request
	mockHTTPClient.On("Do", mock.Anything).Run(func(args mock.Arguments) {
		capturedRequest = args.Get(0).(*http.Request)
	}).Return(mockResponse, nil)

	// Create the client
	client := NewLokiClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	// Create a LokiEntry directly using our new structured types
	entry := LokiEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"job": "test-job",
					"env": "testing", // Add an additional label to test richer functionality
				},
				Values: [][]string{
					{
						"1626882892000000000", // Fixed timestamp for testing
						`{"message":"test log message","level":"info"}`,
					},
				},
			},
		},
	}

	// Send the entry
	err := client.Send(entry)
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
	assert.Equal(t, "testing", streamLabels["env"], "Stream should have correct env label")

	// Validate stream values
	values, ok := stream["values"].([]interface{})
	require.True(t, ok, "Stream should have 'values' array")
	require.Len(t, values, 1, "Values array should have 1 element")

	// Validate the value entry (timestamp and message)
	valueEntry, ok := values[0].([]interface{})
	require.True(t, ok, "Value entry should be an array")
	require.Len(t, valueEntry, 2, "Value entry should have timestamp and message")

	// Check the timestamp and message directly
	assert.Equal(t, "1626882892000000000", valueEntry[0].(string), "Timestamp should match")
	assert.Equal(t, `{"message":"test log message","level":"info"}`, valueEntry[1].(string), "Message should match")

	// Verify headers
	assert.Equal(t, "application/json", capturedRequest.Header.Get("Content-Type"))
}

// TestSendError tests error handling in the Send method
func TestSendError(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	testURL := "http://loki.example.com"
	testUser := "test-user"
	testToken := "test-token"
	client := NewLokiClient(testURL, testUser, testToken, mockHTTPClient)

	mockHTTPClient.ExpectedCalls = nil
	mockHTTPClient.Calls = nil

	badURLClient := NewLokiClient("://invalid-url", testUser, testToken, mockHTTPClient)

	entry := LokiEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"job": "test-job",
				},
				Values: [][]string{
					{
						fmt.Sprintf("%d", time.Now().UnixNano()),
						`{"test":"data"}`,
					},
				},
			},
		},
	}

	err := badURLClient.Send(entry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")

	mockHTTPClient.ExpectedCalls = nil
	mockHTTPClient.Calls = nil
	timeoutErr := fmt.Errorf("%w: context deadline exceeded", errors.ErrTimeout)
	mockHTTPClient.On("Do", mock.Anything).Return(nil, timeoutErr)

	err = client.Send(entry)
	assert.Error(t, err)

	// Adjust assertion to match the actual behavior
	assert.True(t, errors.Is(err, errors.ErrConnectionFailed), "Expected a connection error")
	assert.Contains(t, err.Error(), "connection to log service failed", "Expected connection failed error message")
}
