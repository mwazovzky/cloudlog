package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mwazovzky/cloudlog/errors"
)

// LogSender defines the interface for sending log entries to backends
type LogSender interface {
	Send(ctx context.Context, entry LokiEntry) error
}

// LokiStream represents a single stream in the Loki protocol
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// LokiEntry represents the full payload for Loki's push API
type LokiEntry struct {
	Streams []LokiStream `json:"streams"`
}

// HTTPClient is an interface that matches http.Client's Do method
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// LokiClient sends log entries to Grafana Loki
type LokiClient struct {
	url    string
	user   string
	token  string
	client HTTPClient
}

// NewLokiClient creates a new LokiClient
func NewLokiClient(url, user, token string, httpClient HTTPClient) *LokiClient {
	return &LokiClient{
		url:    url,
		user:   user,
		token:  token,
		client: httpClient,
	}
}

// Send sends a pre-constructed Loki entry to the Loki server
func (c *LokiClient) Send(ctx context.Context, entry LokiEntry) error {
	lokiData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("%w: failed to format Loki payload: %v", errors.ErrInvalidFormat, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewBuffer(lokiData))
	if err != nil {
		return fmt.Errorf("%w: failed to create request: %v", errors.ErrInvalidInput, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrConnectionFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status code %d: %s", errors.ErrResponseError, resp.StatusCode, string(body))
	}

	return nil
}
