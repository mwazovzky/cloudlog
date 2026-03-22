package logger

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLogSender is a test implementation of client.LogSender
type mockLogSender struct {
	mu      sync.Mutex
	entries []client.LokiEntry
	err     error
}

func (m *mockLogSender) Send(_ context.Context, entry client.LokiEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.entries = append(m.entries, entry)
	return nil
}

func TestSyncSender_Send(t *testing.T) {
	mock := &mockLogSender{}
	sender := NewSyncSender(mock)

	labels := map[string]string{"job": "test-job"}
	content := []byte(`{"message":"hello"}`)
	ts := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)

	err := sender.Send(context.Background(), content, labels, ts)
	assert.NoError(t, err)

	require.Len(t, mock.entries, 1)
	entry := mock.entries[0]

	require.Len(t, entry.Streams, 1)
	assert.Equal(t, "test-job", entry.Streams[0].Stream["job"])

	require.Len(t, entry.Streams[0].Values, 1)
	assert.Equal(t, fmt.Sprintf("%d", ts.UnixNano()), entry.Streams[0].Values[0][0])
	assert.Equal(t, `{"message":"hello"}`, entry.Streams[0].Values[0][1])
}

func TestSyncSender_PropagatesError(t *testing.T) {
	mock := &mockLogSender{err: fmt.Errorf("send failed")}
	sender := NewSyncSender(mock)

	err := sender.Send(context.Background(), []byte("content"), map[string]string{"job": "j"}, time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "send failed")
}
