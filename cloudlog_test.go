package cloudlog

import (
	"context"
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

var ctx = context.Background()

// mockLogSender implements client.LogSender for testing
type mockLogSender struct {
	LastJob       string
	LastFormatted []byte
	ShouldError   bool
}

func (m *mockLogSender) Send(_ context.Context, entry client.LokiEntry) error {
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

func newTestLogger(mock *mockLogSender, opts ...Option) Logger {
	return New(NewSyncSender(mock), opts...)
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://test", "user", "token", &http.Client{})
	assert.NotNil(t, c)
}

func TestCloudLog_Info(t *testing.T) {
	mock := &mockLogSender{}
	log := newTestLogger(mock)

	err := log.Info(ctx, "Test message", "key1", "value1", "key2", 42)
	assert.NoError(t, err)

	require.NotNil(t, mock.LastFormatted)

	var logData map[string]interface{}
	err = json.Unmarshal(mock.LastFormatted, &logData)
	require.NoError(t, err)

	assert.Equal(t, "info", logData["level"])
	assert.Equal(t, "Test message", logData["message"])
	assert.Equal(t, "value1", logData["key1"])
	assert.Equal(t, float64(42), logData["key2"])
}

func TestWithFormatter(t *testing.T) {
	mock := &mockLogSender{}
	stringFormatter := formatter.NewStringFormatter()
	log := newTestLogger(mock, WithFormatter(stringFormatter))

	err := log.Info(ctx, "Test message", "key1", "value1")
	assert.NoError(t, err)

	output := string(mock.LastFormatted)

	assert.Contains(t, output, "job=application")
	assert.Contains(t, output, "level=info")
	assert.Contains(t, output, "message=Test message")
	assert.Contains(t, output, "key1=value1")
}

func TestWithJob(t *testing.T) {
	mock := &mockLogSender{}
	log := newTestLogger(mock, WithJob("custom-job"))

	err := log.Info(ctx, "Test message")
	assert.NoError(t, err)

	assert.Equal(t, "custom-job", mock.LastJob)
}

func TestWithMetadata(t *testing.T) {
	mock := &mockLogSender{}
	log := newTestLogger(mock, WithMetadata("version", "1.0"))

	err := log.Info(ctx, "Test message")
	assert.NoError(t, err)

	require.NotNil(t, mock.LastFormatted)

	var logData map[string]interface{}
	err = json.Unmarshal(mock.LastFormatted, &logData)
	require.NoError(t, err)

	assert.Equal(t, "1.0", logData["version"])
}

func TestWith(t *testing.T) {
	mock := &mockLogSender{}
	log := newTestLogger(mock)

	userLogger := log.With("user_id", "123", "request_id", "req-456")

	err := userLogger.Info(ctx, "User action")
	assert.NoError(t, err)

	require.NotNil(t, mock.LastFormatted)

	var logData map[string]interface{}
	err = json.Unmarshal(mock.LastFormatted, &logData)
	require.NoError(t, err)

	assert.Equal(t, "User action", logData["message"])
	assert.Equal(t, "123", logData["user_id"])
	assert.Equal(t, "req-456", logData["request_id"])
}

func TestLoggerChaining(t *testing.T) {
	mock := &mockLogSender{}
	log := newTestLogger(mock,
		WithJob("base-service"),
		WithMetadata("version", "1.0"))

	userLogger := log.With("context_key", "context_value")
	jobLogger := userLogger.WithJob("specific-job")

	err := jobLogger.Info(ctx, "Chained logger test")
	assert.NoError(t, err)

	assert.Equal(t, "specific-job", mock.LastJob)

	require.NotNil(t, mock.LastFormatted)

	var logData map[string]interface{}
	err = json.Unmarshal(mock.LastFormatted, &logData)
	require.NoError(t, err)

	assert.Equal(t, "Chained logger test", logData["message"])
	assert.Equal(t, "1.0", logData["version"])
	assert.Equal(t, "context_value", logData["context_key"])
}

func TestErrorHandling(t *testing.T) {
	mock := &mockLogSender{ShouldError: true}
	log := newTestLogger(mock)

	err := log.Info(ctx, "Test info")
	assert.Error(t, err)
	assert.True(t, IsConnectionError(err))
}

func TestFormatterOptions(t *testing.T) {
	assert.NotNil(t, NewLokiFormatter())
	assert.NotNil(t, WithLabelKeys("request_id", "user_id"))
	assert.NotNil(t, WithTimeFormat(time.RFC3339))
}

func TestHttpClientOptions(t *testing.T) {
	c := NewClient("http://example.com", "user", "token", &http.Client{})
	assert.NotNil(t, c)
}
