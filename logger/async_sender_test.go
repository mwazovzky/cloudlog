package logger

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mwazovzky/cloudlog/client"
	"github.com/mwazovzky/cloudlog/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	stderrors "errors"
)

// asyncMockLogSender captures LokiEntry sends for testing
type asyncMockLogSender struct {
	mu      sync.Mutex
	entries []client.LokiEntry
	err     error
}

func (m *asyncMockLogSender) Send(_ context.Context, entry client.LokiEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *asyncMockLogSender) getEntries() []client.LokiEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]client.LokiEntry{}, m.entries...)
}

func (m *asyncMockLogSender) totalValues() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, e := range m.entries {
		for _, s := range e.Streams {
			count += len(s.Values)
		}
	}
	return count
}

func TestAsyncSender_BasicSendAndFlush(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock, WithBatchSize(10))
	defer sender.Close()

	labels := map[string]string{"job": "test"}
	err := sender.Send(ctx, []byte(`{"msg":"hello"}`), labels, time.Now())
	assert.NoError(t, err)

	sender.Flush()

	assert.Equal(t, 1, mock.totalValues())
}

func TestAsyncSender_Batching(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock, WithBatchSize(5), WithFlushInterval(time.Hour))
	defer sender.Close()

	labels := map[string]string{"job": "test"}
	for i := 0; i < 5; i++ {
		err := sender.Send(ctx, []byte(fmt.Sprintf(`{"i":%d}`, i)), labels, time.Now())
		assert.NoError(t, err)
	}

	sender.Flush()

	entries := mock.getEntries()
	require.Len(t, entries, 1)
	assert.Len(t, entries[0].Streams[0].Values, 5)
}

func TestAsyncSender_FlushInterval(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock,
		WithBatchSize(1000),
		WithFlushInterval(50*time.Millisecond),
	)
	defer sender.Close()

	labels := map[string]string{"job": "test"}
	err := sender.Send(ctx, []byte(`{"msg":"tick"}`), labels, time.Now())
	assert.NoError(t, err)

	// Wait for flush interval to fire
	time.Sleep(150 * time.Millisecond)

	assert.Equal(t, 1, mock.totalValues())
}

func TestAsyncSender_FlushDrainsAll(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock, WithBatchSize(1000))
	defer sender.Close()

	labels := map[string]string{"job": "test"}
	for i := 0; i < 50; i++ {
		err := sender.Send(ctx, []byte(fmt.Sprintf(`{"i":%d}`, i)), labels, time.Now())
		assert.NoError(t, err)
	}

	sender.Flush()

	assert.Equal(t, 50, mock.totalValues())
}

func TestAsyncSender_CloseFlushesAndStops(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock, WithBatchSize(1000))

	labels := map[string]string{"job": "test"}
	for i := 0; i < 10; i++ {
		err := sender.Send(ctx, []byte(fmt.Sprintf(`{"i":%d}`, i)), labels, time.Now())
		assert.NoError(t, err)
	}

	sender.Close()

	assert.Equal(t, 10, mock.totalValues())
}

func TestAsyncSender_BufferFullNonBlocking(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock,
		WithBufferSize(2),
		WithBatchSize(1000),
		WithFlushInterval(time.Hour),
	)
	defer sender.Close()

	labels := map[string]string{"job": "test"}

	// Fill the buffer (2 entries) plus one that may or may not fit depending on worker timing
	var bufferFullErr error
	for i := 0; i < 100; i++ {
		err := sender.Send(ctx, []byte(`{"msg":"fill"}`), labels, time.Now())
		if err != nil {
			bufferFullErr = err
			break
		}
	}

	require.Error(t, bufferFullErr)
	assert.True(t, stderrors.Is(bufferFullErr, errors.ErrBufferFull))
}

func TestAsyncSender_BufferFullBlocking(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock,
		WithBufferSize(1),
		WithBatchSize(1),
		WithBlockOnFull(true),
	)
	defer sender.Close()

	labels := map[string]string{"job": "test"}

	// Should not block forever — worker drains the buffer
	for i := 0; i < 10; i++ {
		err := sender.Send(ctx, []byte(fmt.Sprintf(`{"i":%d}`, i)), labels, time.Now())
		assert.NoError(t, err)
	}

	sender.Flush()
	assert.Equal(t, 10, mock.totalValues())
}

func TestAsyncSender_ErrorHandler(t *testing.T) {
	mock := &asyncMockLogSender{err: fmt.Errorf("loki down")}
	var errorCount atomic.Int32
	sender := NewAsyncSender(mock,
		WithBatchSize(1),
		WithErrorHandler(func(_ error) {
			errorCount.Add(1)
		}),
	)
	defer sender.Close()

	labels := map[string]string{"job": "test"}
	err := sender.Send(ctx, []byte(`{"msg":"fail"}`), labels, time.Now())
	assert.NoError(t, err) // Send itself succeeds (buffered)

	sender.Flush()

	assert.GreaterOrEqual(t, errorCount.Load(), int32(1))
}

func TestAsyncSender_ConcurrentSend(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock,
		WithBufferSize(1000),
		WithBatchSize(50),
	)
	defer sender.Close()

	var wg sync.WaitGroup
	labels := map[string]string{"job": "test"}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = sender.Send(ctx, []byte(fmt.Sprintf(`{"g":%d,"i":%d}`, n, j)), labels, time.Now())
			}
		}(i)
	}

	wg.Wait()
	sender.Flush()

	assert.Equal(t, 1000, mock.totalValues())
}

func TestAsyncSender_GroupsByJob(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock, WithBatchSize(10))
	defer sender.Close()

	err := sender.Send(ctx, []byte(`{"msg":"a"}`), map[string]string{"job": "svc-a"}, time.Now())
	assert.NoError(t, err)
	err = sender.Send(ctx, []byte(`{"msg":"b"}`), map[string]string{"job": "svc-b"}, time.Now())
	assert.NoError(t, err)

	sender.Flush()

	entries := mock.getEntries()
	// May be 1 or 2 LokiEntry calls depending on batching, but total values = 2
	assert.Equal(t, 2, mock.totalValues())

	// Verify each stream has the correct job
	for _, e := range entries {
		for _, s := range e.Streams {
			job := s.Stream["job"]
			assert.Contains(t, []string{"svc-a", "svc-b"}, job)
		}
	}
}

func TestAsyncSender_DoubleCloseIsSafe(t *testing.T) {
	mock := &asyncMockLogSender{}
	sender := NewAsyncSender(mock)

	sender.Close()
	sender.Close() // should not panic
}
