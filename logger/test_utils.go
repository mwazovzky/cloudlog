package logger

import (
	"fmt"
	"sync"

	"github.com/mwazovzky/cloudlog/client"
)

// mockSender is a test implementation of client.LogSender
type mockSender struct {
	mu         sync.Mutex
	messages   []string
	jobs       []string
	shouldFail bool
}

func (m *mockSender) Send(entry client.LokiEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("simulated send failure")
	}

	m.messages = append(m.messages, entry.Streams[0].Values[0][1])
	m.jobs = append(m.jobs, entry.Streams[0].Stream["job"])
	return nil
}

func (m *mockSender) GetMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

func (m *mockSender) GetJobs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.jobs
}
