package cloudlog

// Package logger provides a simple interface for logging messages to Grafana Loki.
// It allows logging messages with different severity levels (info, error, debug)
// and supports adding key-value pairs to the log entries.

import (
	"bytes"
	"fmt"
	"net/http"
)

// Doer is an interface that matches the behavior of http.Client's Do method.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	url    string
	user   string
	token  string
	client Doer // Use the Doer interface instead of *http.Client
}

func NewClient(url, user, token string, httpClient Doer) *Client {
	return &Client{
		url:    url,
		user:   user,
		token:  token,
		client: httpClient,
	}
}

func (c *Client) Send(logEntry LogEntry) error {
	data, err := logEntry.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize log entry: %w", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.client.Do(req) // Use the Doer interface
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}
