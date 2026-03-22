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

	content, err := formatter.Format(entry)
	require.NoError(t, err)

	var logData map[string]interface{}
	err = json.Unmarshal(content, &logData)
	require.NoError(t, err)

	assert.Equal(t, timestamp.Format(customTimeFormat), logData["timestamp"])
}

func TestLokiFormatter_Format(t *testing.T) {
	timestamp := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)

	entry := LogEntry{
		Timestamp: timestamp,
		Job:       "test-job",
		Level:     "info",
		KeyVals: map[string]interface{}{
			"message": "Test message",
			"user_id": "user-123",
		},
	}

	formatter := NewLokiFormatter()

	content, err := formatter.Format(entry)
	require.NoError(t, err)

	var logData map[string]interface{}
	err = json.Unmarshal(content, &logData)
	require.NoError(t, err)

	assert.Equal(t, "test-job", logData["job"])
	assert.Equal(t, "info", logData["level"])
	assert.Equal(t, "Test message", logData["message"])
	assert.Equal(t, "user-123", logData["user_id"])
	assert.Contains(t, logData, "timestamp")
}

func TestLokiFormatter_EdgeCases(t *testing.T) {
	formatter := NewLokiFormatter()
	entry := LogEntry{
		Timestamp: time.Time{},
		Job:       "",
		Level:     "",
		KeyVals:   nil,
	}

	content, err := formatter.Format(entry)
	require.NoError(t, err, "Format should not fail with empty entry")

	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	require.NoError(t, err, "Result should be valid JSON")

	entry = LogEntry{
		Timestamp: time.Now(),
		Job:       "test-job",
		Level:     "info",
		KeyVals:   map[string]interface{}{},
	}

	content, err = formatter.Format(entry)
	require.NoError(t, err)

	err = json.Unmarshal(content, &result)
	require.NoError(t, err, "Result should be valid JSON")
}
