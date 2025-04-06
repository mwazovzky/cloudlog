package formatter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mwazovzky/cloudlog/errors"
)

func TestJSONFormatter_Format(t *testing.T) {
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

	formatter := NewJSONFormatter()

	// Format the entry
	bytes, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Parse the JSON
	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify fields
	if result["timestamp"] != timestamp.Format(time.RFC3339) {
		t.Errorf("Expected timestamp %s, got %s", timestamp.Format(time.RFC3339), result["timestamp"])
	}

	if result["job"] != "test-job" {
		t.Errorf("Expected job test-job, got %s", result["job"])
	}

	if result["level"] != "info" {
		t.Errorf("Expected level info, got %s", result["level"])
	}

	if result["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got key1=%v", result["key1"])
	}

	if result["key2"].(float64) != 42 {
		t.Errorf("Expected key2=42, got key2=%v", result["key2"])
	}
}

type UnmarshalableType struct {
	Ch chan int
}

func TestJSONFormatter_FormatError(t *testing.T) {
	formatter := NewJSONFormatter()

	// Create an entry with a value that can't be marshaled to JSON
	entry := LogEntry{
		Timestamp: time.Now(),
		Job:       "test",
		Level:     "info",
		KeyVals: map[string]interface{}{
			"unmarshalable": UnmarshalableType{Ch: make(chan int)},
		},
	}

	// Attempt to format should fail
	_, err := formatter.Format(entry)

	// Verify error is of the right type
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}

	if !errors.Is(err, errors.ErrInvalidFormat) {
		t.Fatalf("Expected format error, got: %v", err)
	}

	// Error should contain useful message
	if !contains(err.Error(), "marshal JSON") {
		t.Errorf("Error message doesn't mention JSON marshaling: %v", err)
	}
}

func contains(s, substr string) bool {
	return s != "" && s != substr && s != " " && strings.Contains(s, substr)
}

func TestJSONFormatter_CustomOptions(t *testing.T) {
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
	formatter := NewJSONFormatter(
		WithTimeFormat(time.RFC1123),
		WithTimestampField("time"),
		WithLevelField("severity"),
		WithJobField("service"),
	)

	// Format the entry
	bytes, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Parse the JSON
	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify custom fields
	if result["time"] != timestamp.Format(time.RFC1123) {
		t.Errorf("Expected time %s, got %s", timestamp.Format(time.RFC1123), result["time"])
	}

	if result["service"] != "test-job" {
		t.Errorf("Expected service test-job, got %s", result["service"])
	}

	if result["severity"] != "info" {
		t.Errorf("Expected severity info, got %s", result["severity"])
	}

	// Verify original fields don't exist
	if _, exists := result["timestamp"]; exists {
		t.Errorf("Field timestamp should not exist")
	}

	if _, exists := result["job"]; exists {
		t.Errorf("Field job should not exist")
	}

	if _, exists := result["level"]; exists {
		t.Errorf("Field level should not exist")
	}
}
