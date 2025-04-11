package formatter

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringFormatter(t *testing.T) {
	// Create a formatter with default settings
	formatter := NewStringFormatter()

	// Create a log entry
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

	// Format the entry
	lokiEntry, err := formatter.Format(entry)
	assert.NoError(t, err)

	// Extract string content from the LokiEntry
	require.NotEmpty(t, lokiEntry.Streams)
	require.NotEmpty(t, lokiEntry.Streams[0].Values)
	require.GreaterOrEqual(t, len(lokiEntry.Streams[0].Values[0]), 2)

	formatted := lokiEntry.Streams[0].Values[0][1]

	// Check the formatted string
	assert.Contains(t, formatted, "time="+timestamp.Format(time.RFC3339))
	assert.Contains(t, formatted, "job=test-job")
	assert.Contains(t, formatted, "level=info")
	assert.Contains(t, formatted, "message=Test message")
	assert.Contains(t, formatted, "user_id=user-123")
}

func TestStringFormatter_CustomSettings(t *testing.T) {
	// Create a formatter with custom settings
	formatter := NewStringFormatter(
		String.WithTimeFormat("2006-01-02"),
		WithKeyValueSeparator(": "),
		WithPairSeparator(" | "),
	)

	// Create a log entry
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

	// Format the entry
	lokiEntry, err := formatter.Format(entry)
	assert.NoError(t, err)

	// Extract string content from the LokiEntry
	require.NotEmpty(t, lokiEntry.Streams)
	require.NotEmpty(t, lokiEntry.Streams[0].Values)
	require.GreaterOrEqual(t, len(lokiEntry.Streams[0].Values[0]), 2)

	formatted := lokiEntry.Streams[0].Values[0][1]

	// Check the formatted string with custom settings
	assert.Contains(t, formatted, "time: "+timestamp.Format("2006-01-02"))
	assert.Contains(t, formatted, "job: test-job")
	assert.Contains(t, formatted, "level: info")
	assert.Contains(t, formatted, "message: Test message")
	assert.Contains(t, formatted, "user_id: user-123")

	// Check the separator between pairs
	pairs := strings.Split(formatted, " | ")
	assert.GreaterOrEqual(t, len(pairs), 5)

	// Each pair should use the custom key-value separator
	for _, pair := range pairs {
		if strings.Contains(pair, ":") {
			parts := strings.SplitN(pair, ": ", 2)
			assert.Equal(t, 2, len(parts))
		}
	}
}

func TestStringFormatter_EmptyEntry(t *testing.T) {
	// Create a formatter
	formatter := NewStringFormatter()

	// Create an empty log entry
	entry := LogEntry{
		Timestamp: time.Time{}, // Zero time
		Job:       "",
		Level:     "",
		KeyVals:   map[string]interface{}{},
	}

	// Format should not panic with empty entry
	lokiEntry, err := formatter.Format(entry)
	assert.NoError(t, err)

	// Extract string content from the LokiEntry
	require.NotEmpty(t, lokiEntry.Streams)
	require.NotEmpty(t, lokiEntry.Streams[0].Values)
	require.GreaterOrEqual(t, len(lokiEntry.Streams[0].Values[0]), 2)

	formatted := lokiEntry.Streams[0].Values[0][1]

	// Should still have the basic structure
	assert.Contains(t, formatted, "time=")
	assert.Contains(t, formatted, "job=")
	assert.Contains(t, formatted, "level=")
}

func TestStringFormatter_FieldAccess(t *testing.T) {
	f := NewStringFormatter()

	// Field names should be keyValueSep and pairSep (not pairSeparator or kvSeparator)
	assert.Equal(t, "=", f.keyValueSep)
	assert.Equal(t, " ", f.pairSep)
}

func TestStringFormatter_FormatBatch(t *testing.T) {
	// Create a formatter with default settings
	formatter := NewStringFormatter()

	// Create timestamps for testing
	timestamp1 := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)
	timestamp2 := time.Date(2023, 6, 15, 12, 31, 0, 0, time.UTC)

	// Create test entries
	entries := []LogEntry{
		{
			Timestamp: timestamp1,
			Job:       "test-job",
			Level:     "info",
			KeyVals: map[string]interface{}{
				"message": "Test message 1",
				"user_id": "user-123",
			},
		},
		{
			Timestamp: timestamp2,
			Job:       "test-job",
			Level:     "error",
			KeyVals: map[string]interface{}{
				"message": "Test message 2",
				"error":   "Something went wrong",
			},
		},
	}

	// Format the batch
	batchedEntry, err := formatter.FormatBatch("test-job", entries)
	assert.NoError(t, err)

	// Extract string content from the LokiEntry
	require.NotEmpty(t, batchedEntry.Streams)
	require.Equal(t, "test-job", batchedEntry.Streams[0].Stream["job"])
	require.Len(t, batchedEntry.Streams[0].Values, 2)

	// Check the first formatted string
	formatted1 := batchedEntry.Streams[0].Values[0][1]
	assert.Contains(t, formatted1, "time="+timestamp1.Format(time.RFC3339))
	assert.Contains(t, formatted1, "job=test-job")
	assert.Contains(t, formatted1, "level=info")
	assert.Contains(t, formatted1, "message=Test message 1")
	assert.Contains(t, formatted1, "user_id=user-123")

	// Check the second formatted string
	formatted2 := batchedEntry.Streams[0].Values[1][1]
	assert.Contains(t, formatted2, "time="+timestamp2.Format(time.RFC3339))
	assert.Contains(t, formatted2, "job=test-job")
	assert.Contains(t, formatted2, "level=error")
	assert.Contains(t, formatted2, "message=Test message 2")
	assert.Contains(t, formatted2, "error=Something went wrong")
}
