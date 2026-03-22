package logger

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mwazovzky/cloudlog/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestLogger_Info(t *testing.T) {
	sender := &mockSender{}
	log := New(sender)

	err := log.Info(ctx, "test message", "key1", "value1")
	assert.NoError(t, err)
	require.Len(t, sender.contents, 1)

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.contents[0]), &logData)
	assert.NoError(t, err)

	assert.Equal(t, "test message", logData["message"])
	assert.Equal(t, "value1", logData["key1"])
	assert.Equal(t, "info", logData["level"])
	assert.Equal(t, "application", logData["job"])
}

func TestLogger_With(t *testing.T) {
	sender := &mockSender{}
	log := New(sender)

	logWithMeta := log.With("key1", "value1")
	err := logWithMeta.Info(ctx, "test message")
	assert.NoError(t, err)
	require.Len(t, sender.contents, 1)

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.contents[0]), &logData)
	assert.NoError(t, err)

	assert.Equal(t, "value1", logData["key1"])
}

func TestLogger_WithJob(t *testing.T) {
	sender := &mockSender{}
	log := New(sender)

	logWithJob := log.WithJob("new-job")
	err := logWithJob.Info(ctx, "test message")
	assert.NoError(t, err)
	require.Len(t, sender.labels, 1)

	assert.Equal(t, "new-job", sender.labels[0]["job"])

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.contents[0]), &logData)
	assert.NoError(t, err)
	assert.Equal(t, "new-job", logData["job"])
}

func TestLogger_AllLogLevels(t *testing.T) {
	sender := &mockSender{}
	log := New(sender)

	assert.NoError(t, log.Info(ctx, "info message"))
	assert.NoError(t, log.Error(ctx, "error message"))
	assert.NoError(t, log.Debug(ctx, "debug message"))
	assert.NoError(t, log.Warn(ctx, "warn message"))
	assert.Len(t, sender.contents, 4)

	levels := []string{"info", "error", "debug", "warn"}
	for i, level := range levels {
		var logData map[string]interface{}
		err := json.Unmarshal([]byte(sender.contents[i]), &logData)
		assert.NoError(t, err)
		assert.Equal(t, level, logData["level"])
	}
}

func TestLogger_Options(t *testing.T) {
	sender := &mockSender{}

	log := New(sender, WithJob("test-job")).(*logger)
	assert.Equal(t, "test-job", log.job)

	log = New(sender, WithMetadata("key1", "value1")).(*logger)
	assert.Equal(t, "value1", log.metadata["key1"])

	log = New(sender, WithFormatter(formatter.NewStringFormatter())).(*logger)
	assert.IsType(t, &formatter.StringFormatter{}, log.formatter)
}

func TestLogger_MinLevel(t *testing.T) {
	sender := &mockSender{}
	log := New(sender, WithMinLevel(LevelWarn))

	assert.NoError(t, log.Debug(ctx, "debug"))
	assert.NoError(t, log.Info(ctx, "info"))
	assert.NoError(t, log.Warn(ctx, "warn"))
	assert.NoError(t, log.Error(ctx, "error"))

	assert.Len(t, sender.contents, 2)
}

func TestLogger_LabelKeys(t *testing.T) {
	sender := &mockSender{}
	log := New(sender, WithLabelKeys("user_id"))

	err := log.Info(ctx, "test", "user_id", "123", "other", "val")
	assert.NoError(t, err)

	assert.Equal(t, "123", sender.labels[0]["user_id"])

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(sender.contents[0]), &logData)
	assert.NoError(t, err)
	assert.Nil(t, logData["user_id"])
	assert.Equal(t, "val", logData["other"])
}

func TestLogger_SendsLabels(t *testing.T) {
	sender := &mockSender{}
	log := New(sender, WithJob("my-service"))

	err := log.Info(ctx, "test")
	assert.NoError(t, err)

	assert.Equal(t, "my-service", sender.labels[0]["job"])
}
