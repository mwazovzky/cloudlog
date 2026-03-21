package formatter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLokiFormatter_CustomTimeFormat(t *testing.T) {
	timestamp := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)

	entry := LogEntry{
		Timestamp: timestamp,
		Job:       "test-job",
		Level:     "info",
		KeyVals: map[string]interface{}{
			"message": "Test message",
		},
	}

	customTimeFormat := time.RFC1123
	formatter := NewLokiFormatter(
		Loki.WithTimeFormat(customTimeFormat),
	)

	lokiEntry, err := formatter.Format(entry)
	require.NoError(t, err)

	formatted, err := json.Marshal(lokiEntry)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(formatted, &result)
	require.NoError(t, err)

	streams := result["streams"].([]interface{})
	stream := streams[0].(map[string]interface{})
	values := stream["values"].([]interface{})
	value := values[0].([]interface{})

	logContent := value[1].(string)
	var logData map[string]interface{}
	err = json.Unmarshal([]byte(logContent), &logData)
	require.NoError(t, err)

	assert.Equal(t, timestamp.Format(customTimeFormat), logData["timestamp"])
}

func TestLokiFormatter_Labels(t *testing.T) {
	timestamp := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)

	entry := LogEntry{
		Timestamp: timestamp,
		Job:       "test-job",
		Level:     "info",
		KeyVals: map[string]interface{}{
			"message":    "Test message",
			"user_id":    "user-123",
			"request_id": "req-456",
			"trace_id":   "trace-789",
			"ip":         "192.168.1.1",
		},
	}

	formatter := NewLokiFormatter(
		WithLabelKeys("user_id", "request_id", "trace_id"),
	)

	lokiEntry, err := formatter.Format(entry)
	require.NoError(t, err)

	formatted, err := json.Marshal(lokiEntry)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(formatted, &result)
	require.NoError(t, err)

	streams := result["streams"].([]interface{})
	stream := streams[0].(map[string]interface{})
	streamLabels := stream["stream"].(map[string]interface{})

	assert.Equal(t, "test-job", streamLabels["job"])
	assert.Equal(t, "user-123", streamLabels["user_id"])
	assert.Equal(t, "req-456", streamLabels["request_id"])
	assert.Equal(t, "trace-789", streamLabels["trace_id"])
	assert.Nil(t, streamLabels["ip"])
}

func TestLokiFormatter_EdgeCases(t *testing.T) {
	formatter := NewLokiFormatter()
	entry := LogEntry{
		Timestamp: time.Time{},
		Job:       "",
		Level:     "",
		KeyVals:   nil,
	}

	lokiEntry, err := formatter.Format(entry)
	require.NoError(t, err, "Format should not fail with empty entry")

	formattedBytes, err := json.Marshal(lokiEntry)
	require.NoError(t, err, "Should marshal LokiEntry to JSON")

	var result map[string]interface{}
	err = json.Unmarshal(formattedBytes, &result)
	require.NoError(t, err, "Result should be valid JSON")

	entry = LogEntry{
		Timestamp: time.Now(),
		Job:       "test-job",
		Level:     "info",
		KeyVals:   map[string]interface{}{},
	}

	formatter = NewLokiFormatter(WithLabelKeys())

	lokiEntry, err = formatter.Format(entry)
	require.NoError(t, err, "Format should not fail with empty labels")

	formattedBytes, err = json.Marshal(lokiEntry)
	require.NoError(t, err, "Should marshal LokiEntry to JSON")

	err = json.Unmarshal(formattedBytes, &result)
	require.NoError(t, err, "Result should be valid JSON")
}
