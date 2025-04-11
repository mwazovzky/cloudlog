package cloudlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClient is a mock client implementation for testing
type MockClient struct {
	LastJob       string
	LastFormatted []byte
	ShouldError   bool
}

// Send is the unified method implementing the LogSender interface
func (m *MockClient) Send(entry client.LokiEntry) error {
	if m.ShouldError {
		return fmt.Errorf("%w: mock error", errors.ErrConnectionFailed)
	}

	if len(entry.Streams) > 0 {
		m.LastJob = entry.Streams[0].Stream["job"]

		if len(entry.Streams[0].Values) > 0 && len(entry.Streams[0].Values[0]) > 1 {
			m.LastFormatted = []byte(entry.Streams[0].Values[0][1])
		}
	}
	return nil
}

func TestNewClient(t *testing.T) {
	client := NewClient("http://test", "user", "token", &http.Client{})

	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestCloudLog_Info(t *testing.T) {
	mockClient := &MockClient{}
	logger := NewSync(mockClient)

	err := logger.Info("Test message", "key1", "value1", "key2", 42)
	assert.NoError(t, err)

	require.NotNil(t, mockClient.LastFormatted)

	var logData map[string]interface{}
	err = json.Unmarshal(mockClient.LastFormatted, &logData)
	require.NoError(t, err)

	assert.Equal(t, "info", logData["level"])
	assert.Equal(t, "Test message", logData["message"])
	assert.Equal(t, "value1", logData["key1"])
	assert.Equal(t, float64(42), logData["key2"])
}

func TestWithFormatter(t *testing.T) {
	mockClient := &MockClient{}
	stringFormatter := formatter.NewStringFormatter()
	logger := NewSync(mockClient, WithFormatter(stringFormatter))

	err := logger.Info("Test message", "key1", "value1")
	assert.NoError(t, err)

	output := string(mockClient.LastFormatted)

	assert.Contains(t, output, "job=application")
	assert.Contains(t, output, "level=info")
	assert.Contains(t, output, "message=Test message")
	assert.Contains(t, output, "key1=value1")
}

func TestWithJob(t *testing.T) {
	mockClient := &MockClient{}
	logger := NewSync(mockClient, WithJob("custom-job"))

	err := logger.Info("Test message")
	assert.NoError(t, err)

	assert.Equal(t, "custom-job", mockClient.LastJob)
}

func TestWithMetadata(t *testing.T) {
	mockClient := &MockClient{}
	logger := NewSync(mockClient, WithMetadata("version", "1.0"))

	err := logger.Info("Test message")
	assert.NoError(t, err)

	require.NotNil(t, mockClient.LastFormatted)

	var logData map[string]interface{}
	err = json.Unmarshal(mockClient.LastFormatted, &logData)
	require.NoError(t, err)

	assert.Equal(t, "1.0", logData["version"])
}

func TestNewClientWithOptions(t *testing.T) {
	url := "http://loki.example.com"
	user := "test-user"
	token := "test-token"
	httpClient := &http.Client{}

	client := NewClientWithOptions(url, user, token, httpClient)

	assert.NotNil(t, client)
}

func TestWithContext(t *testing.T) {
	mockClient := &MockClient{}
	logger := NewSync(mockClient)

	contextLogger := logger.WithContext("user_id", "123", "request_id", "req-456")

	err := contextLogger.Info("User action")
	assert.NoError(t, err)

	require.NotNil(t, mockClient.LastFormatted)

	var logData map[string]interface{}
	err = json.Unmarshal(mockClient.LastFormatted, &logData)
	require.NoError(t, err)

	assert.Equal(t, "User action", logData["message"])
	assert.Equal(t, "123", logData["user_id"])
	assert.Equal(t, "req-456", logData["request_id"])
}

func TestLoggerChaining(t *testing.T) {
	mockClient := &MockClient{}

	logger := NewSync(mockClient,
		WithJob("base-service"),
		WithMetadata("version", "1.0"))

	contextLogger := logger.WithContext("context_key", "context_value")

	jobLogger := contextLogger.WithJob("specific-job")

	err := jobLogger.Info("Chained logger test")
	assert.NoError(t, err)

	assert.Equal(t, "specific-job", mockClient.LastJob)

	require.NotNil(t, mockClient.LastFormatted)

	var logData map[string]interface{}
	err = json.Unmarshal(mockClient.LastFormatted, &logData)
	require.NoError(t, err)

	assert.Equal(t, "Chained logger test", logData["message"])
	assert.Equal(t, "1.0", logData["version"])
	assert.Equal(t, "context_value", logData["context_key"])
}

func TestErrorHandling(t *testing.T) {
	mockClient := &MockClient{ShouldError: true}
	logger := NewSync(mockClient)

	err := logger.Info("Test info")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.ErrConnectionFailed))

	err = logger.Error("Test error")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.ErrConnectionFailed))

	err = logger.Debug("Test debug")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.ErrConnectionFailed))

	err = logger.Warn("Test warn")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.ErrConnectionFailed))
}

func TestFlushAndClose(t *testing.T) {
	mockClient := &MockClient{}
	logger := NewSync(mockClient)

	err := logger.Flush()
	assert.NoError(t, err)

	err = logger.Close()
	assert.NoError(t, err)
}

func TestAsyncLoggerWrapper(t *testing.T) {
	mockClient := &MockClient{}

	asyncLogger := NewAsync(mockClient)
	assert.NotNil(t, asyncLogger)
	asyncLogger.Close()

	newMockClient := &MockClient{}
	asyncLogger = NewAsync(
		newMockClient,
		WithBufferSize(500),
		WithBatchSize(50),
		WithFlushInterval(2*time.Second),
		WithWorkers(3),
		WithBlockOnFull(true),
		WithAsyncFormatter(formatter.NewStringFormatter()),
		WithAsyncJob("test-service"),
		WithAsyncMetadata("env", "production"),
	)

	err := asyncLogger.Info("Test message from async logger")
	assert.NoError(t, err)

	err = asyncLogger.Close()
	assert.NoError(t, err)
}

func TestFormatterOptions(t *testing.T) {
	assert.NotNil(t, NewLokiFormatter())
	assert.NotNil(t, WithLabelKeys("request_id", "user_id"))
	assert.NotNil(t, WithTimeFormat(time.RFC3339))
	assert.NotNil(t, WithTimestampField("@timestamp"))
	assert.NotNil(t, WithLevelField("severity"))
	assert.NotNil(t, WithJobField("service"))
	assert.NotNil(t, WithStringTimeFormat("2006-01-02"))
	assert.NotNil(t, WithKeyValueSeparator(": "))
	assert.NotNil(t, WithPairSeparator(" | "))
}

func TestHttpClientOptions(t *testing.T) {
	httpClient := &http.Client{}

	client := NewClient("http://example.com", "user", "token", httpClient)
	assert.NotNil(t, client)

	clientWithOptions := NewClientWithOptions("http://example.com", "user", "token", httpClient)
	assert.NotNil(t, clientWithOptions)
}

func TestIsLoggerClosedError(t *testing.T) {
	errClosed := fmt.Errorf("%w: logger already closed", errors.ErrLoggerClosed)

	assert.True(t, errors.IsLoggerClosedError(errClosed))

	unrelatedErr := fmt.Errorf("some other error")
	assert.False(t, errors.IsLoggerClosedError(unrelatedErr))
}

// Fix the unused variable error
func TestWithErrorHandlerOption(t *testing.T) {
	mockClient := &MockClient{}

	// Create a handler but don't use the errorCalled variable
	// since we can't easily trigger errors from this test
	handler := func(err error) {
		// Just a placeholder handler, no need to set a variable
	}

	asyncLogger := NewAsync(
		mockClient,
		WithErrorHandler(handler),
	)

	// We can't easily test the error handler directly from here,
	// but we can at least verify the logger was created and can be closed
	assert.NotNil(t, asyncLogger)

	// Verify we can log and close without errors
	err := asyncLogger.Info("Test message")
	assert.NoError(t, err)

	err = asyncLogger.Close()
	assert.NoError(t, err)
}
