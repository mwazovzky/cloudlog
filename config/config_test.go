package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Success(t *testing.T) {
	// Set environment variables
	os.Setenv("LOKI_URL", "http://mock-loki-url")
	os.Setenv("LOKI_USERNAME", "mock-username")
	os.Setenv("LOKI_AUTH_TOKEN", "mock-token")
	defer os.Unsetenv("LOKI_URL")
	defer os.Unsetenv("LOKI_USERNAME")
	defer os.Unsetenv("LOKI_AUTH_TOKEN")

	cfg, err := Load()
	assert.NoError(t, err, "Expected no error when loading valid configuration")
	assert.Equal(t, "http://mock-loki-url", cfg.LokiURL, "Expected LokiURL to match")
	assert.Equal(t, "mock-username", cfg.LokiUsername, "Expected LokiUsername to match")
	assert.Equal(t, "mock-token", cfg.LokiAuthToken, "Expected LokiAuthToken to match")
}

func TestLoad_MissingRequiredEnv(t *testing.T) {
	// Unset environment variables
	os.Unsetenv("LOKI_URL")
	os.Unsetenv("LOKI_USERNAME")
	os.Unsetenv("LOKI_AUTH_TOKEN")

	_, err := Load()
	assert.Error(t, err, "Expected error when required environment variables are missing")
	assert.Contains(t, err.Error(), "missing required environment variable: LOKI_URL", "Expected error message for missing LOKI_URL")
}

func TestLoad_PartialEnv(t *testing.T) {
	// Set only some environment variables
	os.Setenv("LOKI_URL", "http://mock-loki-url")
	os.Setenv("LOKI_USERNAME", "mock-username")
	defer os.Unsetenv("LOKI_URL")
	defer os.Unsetenv("LOKI_USERNAME")
	os.Unsetenv("LOKI_AUTH_TOKEN")

	_, err := Load()
	assert.Error(t, err, "Expected error when some required environment variables are missing")
	assert.Contains(t, err.Error(), "missing required environment variable: LOKI_AUTH_TOKEN", "Expected error message for missing LOKI_AUTH_TOKEN")
}
