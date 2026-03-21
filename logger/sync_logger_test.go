package logger

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestSyncLogger_Info(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender).(*SyncLogger)

	err := logger.Info(ctx, "test message", "key1", "value1")
	assert.NoError(t, err)
	assert.Len(t, sender.messages, 1)

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.messages[0]), &logData)
	assert.NoError(t, err)

	assert.Equal(t, "test message", logData["message"])
	assert.Equal(t, "value1", logData["key1"])
	assert.Equal(t, "info", logData["level"])
	assert.Equal(t, "application", logData["job"])
}

func TestSyncLogger_With(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender)

	loggerWithMeta := logger.With("key1", "value1")
	err := loggerWithMeta.Info(ctx, "test message")
	assert.NoError(t, err)
	assert.Len(t, sender.messages, 1)

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.messages[0]), &logData)
	assert.NoError(t, err)

	assert.Equal(t, "value1", logData["key1"])
}

func TestSyncLogger_WithJob(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender)

	loggerWithJob := logger.WithJob("new-job")
	err := loggerWithJob.Info(ctx, "test message")
	assert.NoError(t, err)
	assert.Len(t, sender.messages, 1)

	assert.Equal(t, "new-job", sender.jobs[0])

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.messages[0]), &logData)
	assert.NoError(t, err)

	assert.Equal(t, "new-job", logData["job"])
}

func TestSyncLogger_AllLogLevels(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender)

	assert.NoError(t, logger.Info(ctx, "info message"))
	assert.NoError(t, logger.Error(ctx, "error message"))
	assert.NoError(t, logger.Debug(ctx, "debug message"))
	assert.NoError(t, logger.Warn(ctx, "warn message"))
	assert.Len(t, sender.messages, 4)

	levels := []string{"info", "error", "debug", "warn"}
	for i, level := range levels {
		var logData map[string]interface{}
		err := json.Unmarshal([]byte(sender.messages[i]), &logData)
		assert.NoError(t, err)
		assert.Equal(t, level, logData["level"])
	}
}

func TestSyncLoggerOptions(t *testing.T) {
	sender := &mockSender{}

	// Test WithJob functional option
	logger := NewSync(sender, WithJob("test-job")).(*SyncLogger)
	assert.Equal(t, "test-job", logger.job)

	// Test WithMetadata
	logger = NewSync(sender, WithMetadata("key1", "value1")).(*SyncLogger)
	assert.Equal(t, "value1", logger.metadata["key1"])
}

func TestSyncLogger_MinLevel(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender, WithMinLevel(LevelWarn))

	assert.NoError(t, logger.Debug(ctx, "debug"))
	assert.NoError(t, logger.Info(ctx, "info"))
	assert.NoError(t, logger.Warn(ctx, "warn"))
	assert.NoError(t, logger.Error(ctx, "error"))

	assert.Len(t, sender.messages, 2) // only warn and error
}

func TestSyncLogger_LabelKeys(t *testing.T) {
	sender := &mockSender{}
	logger := NewSync(sender, WithLabelKeys("user_id"))

	err := logger.Info(ctx, "test", "user_id", "123", "other", "val")
	assert.NoError(t, err)

	// Verify user_id is NOT in log content (promoted to label)
	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.messages[0]), &logData)
	assert.NoError(t, err)
	assert.Nil(t, logData["user_id"])
	assert.Equal(t, "val", logData["other"])
}
