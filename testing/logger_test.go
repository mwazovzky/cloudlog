package testing

import (
	"encoding/json"
	"testing"
)

func TestTestLogger_Basics(t *testing.T) {
	testLogger := NewTestLogger()

	// Send a log entry
	testLogger.Send("test-job", []byte(`{"level":"info","message":"Test message"}`))

	// Check it was captured
	logs := testLogger.Logs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(logs))
	}

	if logs[0].Level != "info" {
		t.Errorf("Expected level info, got %s", logs[0].Level)
	}

	if logs[0].Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %s", logs[0].Message)
	}
}

func TestTestLogger_LogsOfLevel(t *testing.T) {
	testLogger := NewTestLogger()

	// Send log entries of different levels
	testLogger.Send("test-job", []byte(`{"level":"info","message":"Info message"}`))
	testLogger.Send("test-job", []byte(`{"level":"error","message":"Error message"}`))
	testLogger.Send("test-job", []byte(`{"level":"debug","message":"Debug message"}`))

	// Check filtering by level
	infoLogs := testLogger.LogsOfLevel("info")
	if len(infoLogs) != 1 {
		t.Errorf("Expected 1 info log, got %d", len(infoLogs))
	}

	errorLogs := testLogger.LogsOfLevel("error")
	if len(errorLogs) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(errorLogs))
	}

	if errorLogs[0].Message != "Error message" {
		t.Errorf("Expected message 'Error message', got %s", errorLogs[0].Message)
	}
}

func TestTestLogger_ContainsEntry(t *testing.T) {
	testLogger := NewTestLogger()

	// Send a log entry with key-value pairs
	testLogger.Send("test-job", []byte(`{"level":"error","message":"Operation failed","status":500,"request_id":"123"}`))

	// Test ContainsEntry with various criteria
	if !testLogger.ContainsEntry("error", "Operation failed", "status", 500.0) {
		t.Error("Expected entry not found with exact match")
	}

	if testLogger.ContainsEntry("info", "Operation failed", "status", 500.0) {
		t.Error("Found entry with wrong level")
	}

	if testLogger.ContainsEntry("error", "Different message", "status", 500.0) {
		t.Error("Found entry with wrong message")
	}

	if testLogger.ContainsEntry("error", "Operation failed", "status", 404.0) {
		t.Error("Found entry with wrong value")
	}
}

func TestTestLogger_Clear(t *testing.T) {
	testLogger := NewTestLogger()

	// Send a log entry
	testLogger.Send("test-job", []byte(`{"level":"info","message":"Test message"}`))

	// Verify it was captured
	if len(testLogger.Logs()) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(testLogger.Logs()))
	}

	// Clear logs
	testLogger.Clear()

	// Verify logs were cleared
	if len(testLogger.Logs()) != 0 {
		t.Errorf("Expected 0 log entries after Clear(), got %d", len(testLogger.Logs()))
	}
}

func TestTestLogger_ComplexJSON(t *testing.T) {
	testLogger := NewTestLogger()

	// Create a complex JSON log
	complexLog := map[string]interface{}{
		"level":     "warn",
		"message":   "Complex test",
		"timestamp": "2023-06-15T12:30:45Z",
		"context": map[string]interface{}{
			"request_id": "abc-123",
			"user_id":    456,
		},
		"metrics": []interface{}{
			map[string]interface{}{
				"name":  "response_time",
				"value": 123.45,
			},
		},
	}

	logBytes, _ := json.Marshal(complexLog)
	err := testLogger.Send("test-job", logBytes)
	if err != nil {
		t.Fatalf("Failed to send complex log: %v", err)
	}

	// Verify it was captured correctly
	logs := testLogger.Logs()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	if logs[0].Level != "warn" {
		t.Errorf("Expected level warn, got %s", logs[0].Level)
	}

	if logs[0].Message != "Complex test" {
		t.Errorf("Expected message 'Complex test', got %s", logs[0].Message)
	}

	// Check that context data was captured in Data map
	contextData, ok := logs[0].Data["context"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected context to be a map, got %T", logs[0].Data["context"])
	} else {
		if contextData["request_id"] != "abc-123" {
			t.Errorf("Expected request_id abc-123, got %v", contextData["request_id"])
		}
	}
}
