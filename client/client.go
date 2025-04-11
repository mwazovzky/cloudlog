package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mwazovzky/cloudlog/errors"
)

// LogSender defines the interface for sending log entries to backends
type LogSender interface {
	Send(entry LokiEntry) error
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

// The custom MarshalJSON method is redundant since the struct already has proper JSON tags
// and doesn't require any special handling during marshaling.
// Removing this method will use Go's default JSON marshaling which works perfectly fine here.

// Doer is an interface that matches http.Client's Do method
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// LokiClient sends log entries to Grafana Loki
type LokiClient struct {
	url    string
	user   string
	token  string
	client Doer
}

// NewLokiClient creates a new LokiClient
func NewLokiClient(url, user, token string, httpClient Doer) *LokiClient {
	return &LokiClient{
		url:    url,
		user:   user,
		token:  token,
		client: httpClient,
	}
}

// LokiClientOption defines a function to configure a LokiClient
type LokiClientOption func(*LokiClient)

// WithTimeout sets a custom timeout for the client
func WithTimeout(timeout time.Duration) LokiClientOption {
	return func(c *LokiClient) {
		if httpClient, ok := c.client.(*http.Client); ok {
			httpClient.Timeout = timeout
		}
	}
}

// NewLokiClientWithOptions creates a client with the provided options
func NewLokiClientWithOptions(url, user, token string, httpClient Doer, options ...LokiClientOption) *LokiClient {
	client := &LokiClient{
		url:    url,
		user:   user,
		token:  token,
		client: httpClient,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// Send sends a pre-constructed Loki entry to the Loki server
func (c *LokiClient) Send(entry LokiEntry) error {
	// Convert the Loki payload to JSON
	lokiData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("%w: failed to format Loki payload: %v", errors.ErrInvalidFormat, err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(lokiData))
	if err != nil {
		return fmt.Errorf("%w: failed to create request: %v", errors.ErrInvalidInput, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		// Make sure we use ErrResponseError to wrap the error for non-2xx responses
		return fmt.Errorf("%w: status code %d: %s", errors.ErrResponseError, resp.StatusCode, string(body))
	}

	return nil
}
