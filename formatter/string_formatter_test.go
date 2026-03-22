package formatter

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringFormatter(t *testing.T) {
	formatter := NewStringFormatter()

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

	content, err := formatter.Format(entry)
	require.NoError(t, err)

	formatted := string(content)

	assert.Contains(t, formatted, "time="+timestamp.Format(time.RFC3339))
	assert.Contains(t, formatted, "job=test-job")
	assert.Contains(t, formatted, "level=info")
	assert.Contains(t, formatted, "message=Test message")
	assert.Contains(t, formatted, "user_id=user-123")
}

func TestStringFormatter_CustomSettings(t *testing.T) {
	formatter := NewStringFormatter(
		String.WithTimeFormat("2006-01-02"),
		WithKeyValueSeparator(": "),
		WithPairSeparator(" | "),
	)

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

	content, err := formatter.Format(entry)
	require.NoError(t, err)

	formatted := string(content)

	assert.Contains(t, formatted, "time: "+timestamp.Format("2006-01-02"))
	assert.Contains(t, formatted, "job: test-job")
	assert.Contains(t, formatted, "level: info")

	pairs := strings.Split(formatted, " | ")
	assert.GreaterOrEqual(t, len(pairs), 5)
}

func TestStringFormatter_EmptyEntry(t *testing.T) {
	formatter := NewStringFormatter()

	entry := LogEntry{
		Timestamp: time.Time{},
		Job:       "",
		Level:     "",
		KeyVals:   map[string]interface{}{},
	}

	content, err := formatter.Format(entry)
	require.NoError(t, err)

	formatted := string(content)
	assert.Contains(t, formatted, "time=")
	assert.Contains(t, formatted, "job=")
	assert.Contains(t, formatted, "level=")
}

func TestStringFormatter_FieldAccess(t *testing.T) {
	f := NewStringFormatter()

	assert.Equal(t, "=", f.keyValueSep)
	assert.Equal(t, " ", f.pairSep)
}
