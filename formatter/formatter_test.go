package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogEntry(t *testing.T) {
	entry := NewLogEntry("test-job", "info", "key1", "value1", "key2", 42)

	assert.Equal(t, "test-job", entry.Job)
	assert.Equal(t, "info", entry.Level)
	assert.Equal(t, "value1", entry.KeyVals["key1"])
	assert.Equal(t, 42, entry.KeyVals["key2"])
}

func TestLogEntryWithInvalidKeyVals(t *testing.T) {
	entry := NewLogEntry("test-job", "info", 123, "value")
	assert.Empty(t, entry.KeyVals)

	entry = NewLogEntry("test-job", "info", "key")
	assert.Empty(t, entry.KeyVals)
}

func TestLogEntryWithOptions(t *testing.T) {
	entry := NewLogEntryWithOptions(
		WithJob("test-job"),
		WithLevel("INFO"),
	)

	entry.KeyVals = map[string]interface{}{"message": "test message"}

	assert.Equal(t, "test-job", entry.Job)
	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "test message", entry.KeyVals["message"])
}
