package formatter

import (
	"testing"
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
