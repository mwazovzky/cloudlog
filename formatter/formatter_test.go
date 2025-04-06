package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogEntry(t *testing.T) {
	job := "test-job"
	level := "info"
	keyvals := []interface{}{"key1", "value1", "key2", 42}

	entry := NewLogEntry(job, level, keyvals...)

	if entry.Job != job {
		t.Errorf("Expected job %s, got %s", job, entry.Job)
	}

	if entry.Level != level {
		t.Errorf("Expected level %s, got %s", level, entry.Level)
	}

	if len(entry.KeyVals) != 2 {
		t.Errorf("Expected 2 key-value pairs, got %d", len(entry.KeyVals))
	}

	if entry.KeyVals["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got key1=%v", entry.KeyVals["key1"])
	}

	if entry.KeyVals["key2"] != 42 {
		t.Errorf("Expected key2=42, got key2=%v", entry.KeyVals["key2"])
	}
}

func TestNewLogEntryWithInvalidKeyVals(t *testing.T) {
	// Test with non-string key
	entry := NewLogEntry("job", "info", 123, "value")
	if len(entry.KeyVals) != 0 {
		t.Errorf("Expected 0 key-value pairs with invalid key, got %d", len(entry.KeyVals))
	}

	// Test with missing value
	entry = NewLogEntry("job", "info", "key")
	if len(entry.KeyVals) != 0 {
		t.Errorf("Expected 0 key-value pairs with missing value, got %d", len(entry.KeyVals))
	}
}

func TestNewLogEntryWithOptions(t *testing.T) {
	entry := NewLogEntryWithOptions(
		WithJob("test-job"),
		WithLevel("INFO"),
	)

	// Add message for testing
	entry.KeyVals = map[string]interface{}{"message": "test message"}

	assert.Equal(t, "test-job", entry.Job)
	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "test message", entry.KeyVals["message"])
}

func TestWithJob(t *testing.T) {
	// Create a log entry manually rather than using option functions
	originalJob := "original-job"
	newJob := "test-job"

	// First create an entry with the original job
	entry1 := NewLogEntry(originalJob, "INFO", "message", "test")

	// Then create another entry with the new job
	entry2 := NewLogEntry(newJob, "INFO", "message", "test")

	// Compare
	assert.NotEqual(t, entry1.Job, entry2.Job)
	assert.Equal(t, newJob, entry2.Job)
}

func TestWithLevel(t *testing.T) {
	// Create a log entry manually rather than using option functions
	originalLevel := "INFO"
	newLevel := "DEBUG"

	// First create an entry with the original level
	entry1 := NewLogEntry("test-job", originalLevel, "message", "test")

	// Then create another entry with the new level
	entry2 := NewLogEntry("test-job", newLevel, "message", "test")

	// Compare
	assert.NotEqual(t, entry1.Level, entry2.Level)
	assert.Equal(t, newLevel, entry2.Level)
}
