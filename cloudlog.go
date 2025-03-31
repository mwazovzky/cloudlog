package cloudlog

// Package logger provides a simple interface for logging messages to Grafana Loki.
// It allows logging messages with different severity levels (info, error, debug)
// and supports adding key-value pairs to the log entries.

import (
	"encoding/json"
	"fmt"
	"time"
)

// LogStream represents a single log stream.
type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type LogEntry struct {
	Streams []Stream `json:"streams"`
}

func (l *LogEntry) ToJSON() ([]byte, error) {
	return json.Marshal(l)
}

type ClientInterface interface {
	Send(payload LogEntry) error
}

type Logger struct {
	client  ClientInterface
	service string
}

func NewLogger(client ClientInterface, service string) *Logger {
	return &Logger{client: client, service: service}
}

func (l *Logger) Log(level string, message string, keyValues ...interface{}) error {
	if len(keyValues)%2 != 0 {
		return fmt.Errorf("keyValues must be in key-value pairs")
	}

	stream := make(map[string]string)
	// Add service name to the stream
	stream["service_name"] = l.service
	// Add log level to the stream
	stream["level"] = level
	// Convert keyValues to a map
	for i := 0; i < len(keyValues); i += 2 {
		key, ok := keyValues[i].(string)
		if !ok {
			return fmt.Errorf("key must be a string, got %T", keyValues[i])
		}
		// Use %+v to include field names in struct string representation
		stream[key] = fmt.Sprintf("%+v", keyValues[i+1])
	}

	timestamp := time.Now().UnixNano()

	// Prepare the log payload
	payload := LogEntry{
		Streams: []Stream{
			{
				Stream: stream,
				Values: [][]string{{fmt.Sprintf("%d", timestamp), message}},
			},
		},
	}

	return l.client.Send(payload)
}

func (l *Logger) Info(message string, keyValues ...interface{}) error {
	return l.Log("info", message, keyValues...)
}
func (l *Logger) Error(message string, keyValues ...interface{}) error {
	return l.Log("error", message, keyValues...)
}
func (l *Logger) Debug(message string, keyValues ...interface{}) error {
	return l.Log("debug", message, keyValues...)
}
