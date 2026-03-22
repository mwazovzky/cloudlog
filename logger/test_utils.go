package logger

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// mockSender is a test implementation of Sender
type mockSender struct {
	mu         sync.Mutex
	contents   []string
	labels     []map[string]string
	shouldFail bool
}

func (m *mockSender) Send(_ context.Context, content []byte, labels map[string]string, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("simulated send failure")
	}

	m.contents = append(m.contents, string(content))
	m.labels = append(m.labels, labels)
	return nil
}

func (m *mockSender) getContents() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.contents
}

func (m *mockSender) getLabels() []map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.labels
}
