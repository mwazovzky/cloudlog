package logger

import (
	"encoding/json"
	"testing"

	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/stretchr/testify/assert"
)

func TestSyncLogger_Info(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender).(*SyncLogger)

	err := logger.Info("test message", "key1", "value1")
	assert.NoError(t, err)
	assert.Len(t, sender.messages, 1)

	// Parse the JSON log message
	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.messages[0]), &logData)
	assert.NoError(t, err, "Should be able to parse log content as JSON")

	// Verify contents
	assert.Equal(t, "test message", logData["message"])
	assert.Equal(t, "value1", logData["key1"])
	assert.Equal(t, "info", logData["level"])
	assert.Equal(t, "application", logData["job"])
}

func TestSyncLogger_WithContext(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender).(*SyncLogger)

	loggerWithCtx := logger.WithContext("key1", "value1").(*SyncLogger)
	err := loggerWithCtx.Info("test message")
	assert.NoError(t, err)
	assert.Len(t, sender.messages, 1)

	// Parse the JSON log message
	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.messages[0]), &logData)
	assert.NoError(t, err, "Should be able to parse log content as JSON")

	// Verify context was included
	assert.Equal(t, "value1", logData["key1"])
}

func TestSyncLogger_WithJob(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender).(*SyncLogger)

	loggerWithJob := logger.WithJob("new-job").(*SyncLogger)
	err := loggerWithJob.Info("test message")
	assert.NoError(t, err)
	assert.Len(t, sender.messages, 1)

	// Verify the job name in the sender's recorded jobs
	assert.Equal(t, "new-job", sender.jobs[0])

	// Parse the JSON log message
	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.messages[0]), &logData)
	assert.NoError(t, err, "Should be able to parse log content as JSON")

	// Verify job is in the log content
	assert.Equal(t, "new-job", logData["job"])
}

func TestSyncLogger_AllLogLevels(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender).(*SyncLogger)

	// Test all log methods
	err := logger.Info("info message")
	assert.NoError(t, err)

	err = logger.Error("error message")
	assert.NoError(t, err)

	err = logger.Debug("debug message")
	assert.NoError(t, err)

	err = logger.Warn("warn message")
	assert.NoError(t, err)

	assert.Len(t, sender.messages, 4)

	// Verify each log level
	levels := []string{"info", "error", "debug", "warn"}
	for i, level := range levels {
		var logData map[string]interface{}
		err = json.Unmarshal([]byte(sender.messages[i]), &logData)
		assert.NoError(t, err, "Should be able to parse log content as JSON")
		assert.Equal(t, level, logData["level"])
	}
}

func TestSyncLogger_CloseAndFlush(t *testing.T) {
	sender := &mockSender{} // Use the shared mockSender from test_utils.go
	logger := NewSync(sender).(*SyncLogger)

	// Test Close - it's a no-op but should return nil
	err := logger.Close()
	assert.NoError(t, err)

	// Test Flush - also a no-op but should return nil
	err = logger.Flush()
	assert.NoError(t, err)
}

func TestSyncLoggerOptions(t *testing.T) {
	sender := &mockSender{}

	// Test WithFormatter
	customFormatter := formatter.NewStringFormatter()
	logger := NewSync(sender, WithFormatter(customFormatter)).(*SyncLogger)
	assert.Equal(t, customFormatter, logger.formatter)

	// Test WithJob functional option
	jobName := "test-job"
	logger = NewSync(sender, WithJob(jobName)).(*SyncLogger)
	assert.Equal(t, jobName, logger.job)

	// Test WithMetadata
	logger = NewSync(sender, WithMetadata("key1", "value1")).(*SyncLogger)
	assert.Equal(t, "value1", logger.metadata["key1"])

	// Test WithMetadataMap
	metadataMap := map[string]interface{}{
		"env":     "test",
		"version": "1.0.0",
	}
	logger = NewSync(sender, WithMetadataMap(metadataMap)).(*SyncLogger)
	assert.Equal(t, "test", logger.metadata["env"])
	assert.Equal(t, "1.0.0", logger.metadata["version"])
}
