package cloudlog

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestClient_Send_Success(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	mockResponse := &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	client := NewClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	payload := LogEntry{
		Streams: []Stream{
			{
				Stream: map[string]string{"job": "test-job"},
				Values: [][]string{{"1234567890", "test log"}},
			},
		},
	}

	err := client.Send(payload)
	assert.NoError(t, err, "Expected no error for successful log send")
	mockHTTPClient.AssertExpectations(t)
}

func TestClient_Send_Failure(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	mockResponse := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	client := NewClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	payload := LogEntry{
		Streams: []Stream{
			{
				Stream: map[string]string{"job": "test-job"},
				Values: [][]string{{"1234567890", "test log"}},
			},
		},
	}

	err := client.Send(payload)
	assert.Error(t, err, "Expected error for failed log send")
	mockHTTPClient.AssertExpectations(t)
}

func TestClient_Send_NetworkError(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	mockHTTPClient.On("Do", mock.Anything).Return(nil, errors.New("network error"))

	client := NewClient("http://mock-loki-url", "test-user", "test-token", mockHTTPClient)

	payload := LogEntry{
		Streams: []Stream{
			{
				Stream: map[string]string{"job": "test-job"},
				Values: [][]string{{"1234567890", "test log"}},
			},
		},
	}

	err := client.Send(payload)
	assert.Error(t, err, "Expected error for network failure")
	mockHTTPClient.AssertExpectations(t)
}
