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

// Send sends a log entry to Loki
// Returns an error if the request fails or receives a non-204 response
func (c *LokiClient) Send(job string, data []byte) error {
	// Create Loki-specific payload structure
	// Loki expects a specific format with 'streams' containing 'labels' and 'entries'
	lokiPayload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{
					"job": job,
				},
				"values": [][]string{
					{
						fmt.Sprintf("%d", time.Now().UnixNano()),
						string(data),
					},
				},
			},
		},
	}

	// Convert the Loki payload to JSON
	lokiData, err := json.Marshal(lokiPayload)
	if err != nil {
		return errors.FormatError(err, "failed to format Loki payload")
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(lokiData))
	if err != nil {
		return errors.InputError(fmt.Sprintf("failed to create request: %v", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.ConnectionError(err, "failed to send log entry to Loki")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.ResponseError(resp.StatusCode, string(body))
	}

	return nil
}
