package config

import (
	"errors"
	"os"
)

type Config struct {
	LokiURL       string
	LokiUsername  string
	LokiAuthToken string
}

func Load() (*Config, error) {
	url := os.Getenv("LOKI_URL")
	username := os.Getenv("LOKI_USERNAME")
	token := os.Getenv("LOKI_AUTH_TOKEN")

	if url == "" {
		return nil, errors.New("missing required environment variable: LOKI_URL")
	}
	if username == "" {
		return nil, errors.New("missing required environment variable: LOKI_USERNAME")
	}
	if token == "" {
		return nil, errors.New("missing required environment variable: LOKI_AUTH_TOKEN")
	}

	return &Config{
		LokiURL:       url,
		LokiUsername:  username,
		LokiAuthToken: token,
	}, nil
}
