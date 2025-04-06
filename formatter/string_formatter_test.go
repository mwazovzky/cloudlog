package formatter

import (
	"strings"
	"testing"
	"time"
)

func TestStringFormatter_Format(t *testing.T) {
	// Create a fixed timestamp for testing
	timestamp := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)

	entry := LogEntry{
		Timestamp: timestamp,
		Job:       "test-job",
		Level:     "info",
		KeyVals: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	formatter := NewStringFormatter()

	// Format the entry
	bytes, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := string(bytes)

	// We need to check parts because map iteration order is non-deterministic
	if !strings.Contains(result, "job=test-job") {
		t.Errorf("Expected string to contain 'job=test-job', got: %s", result)
	}

	if !strings.Contains(result, "level=info") {
		t.Errorf("Expected string to contain 'level=info', got: %s", result)
	}

	if !strings.Contains(result, "key1=value1") {
		t.Errorf("Expected string to contain 'key1=value1', got: %s", result)
	}

	if !strings.Contains(result, "key2=42") {
		t.Errorf("Expected string to contain 'key2=42', got: %s", result)
	}
}

func TestStringFormatter_CustomOptions(t *testing.T) {
	// Create a fixed timestamp for testing
	timestamp := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)

	entry := LogEntry{
		Timestamp: timestamp,
		Job:       "test-job",
		Level:     "info",
		KeyVals: map[string]interface{}{
			"key1": "value1",
		},
	}

	// Create formatter with custom options
	formatter := NewStringFormatter(
		WithStringTimeFormat("2006/01/02"),
		WithKeyValueSeparator(": "),
		WithPairSeparator(" | "),
	)

	// Format the entry
	bytes, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := string(bytes)

	// Check date format
	if !strings.HasPrefix(result, "2023/06/15") {
		t.Errorf("Expected string to start with '2023/06/15', got: %s", result)
	}

	// Check separators
	if !strings.Contains(result, "job: test-job") {
		t.Errorf("Expected string to contain 'job: test-job', got: %s", result)
	}

	if !strings.Contains(result, "level: info") {
		t.Errorf("Expected string to contain 'level: info', got: %s", result)
	}

	if !strings.Contains(result, "key1: value1") {
		t.Errorf("Expected string to contain 'key1: value1', got: %s", result)
	}

	// Check pair separator
	if !strings.Contains(result, " | ") {
		t.Errorf("Expected string to contain ' | ' as pair separator, got: %s", result)
	}
}

func TestStringFormatter_EmptyEntry(t *testing.T) {
	// Create a fixed timestamp for testing
	timestamp := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)

	entry := LogEntry{
		Timestamp: timestamp,
		Job:       "test-job",
		KeyVals:   map[string]interface{}{},
	}

	formatter := NewStringFormatter()

	// Format the entry
	bytes, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := string(bytes)
	expected := timestamp.Format(time.RFC3339) + " job=test-job"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// Add method for validating string format for test usage
func (f *StringFormatter) ValidString(s string) bool {
	// This is a simple validation that can be used in tests
	// A "valid" string has the right separators and contains required fields

	// Check for timestamp format
	if !strings.Contains(s, " job") {
		return false
	}

	// Check separator usage
	for _, pair := range strings.Split(s, f.pairSeparator) {
		if len(pair) > 0 && !strings.Contains(pair, f.kvSeparator) {
			return false
		}
	}

	return true
}
