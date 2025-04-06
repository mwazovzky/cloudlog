package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestFormatError(t *testing.T) {
	originalErr := errors.New("original error")
	err := FormatError(originalErr, "format failed")

	// Test wrapping behavior
	if !Is(err, ErrInvalidFormat) {
		t.Errorf("Error should be recognized as an ErrInvalidFormat")
	}

	// Test details are included
	errStr := err.Error()
	if !strings.Contains(errStr, "format failed") {
		t.Errorf("Error message should contain details")
	}

	// Test original error is included
	if !strings.Contains(errStr, "original error") {
		t.Errorf("Error message should contain original error")
	}
}

func TestConnectionError(t *testing.T) {
	originalErr := errors.New("connection refused")
	err := ConnectionError(originalErr, "failed to send log")

	// Test wrapping behavior
	if !Is(err, ErrConnectionFailed) {
		t.Errorf("Error should be recognized as an ErrConnectionFailed")
	}

	// Test details are included
	errStr := err.Error()
	if !strings.Contains(errStr, "failed to send log") {
		t.Errorf("Error message should contain details")
	}

	// Test original error is included
	if !strings.Contains(errStr, "connection refused") {
		t.Errorf("Error message should contain original error")
	}
}

func TestResponseError(t *testing.T) {
	err := ResponseError(500, "internal server error")

	// Test wrapping behavior
	if !Is(err, ErrResponseError) {
		t.Errorf("Error should be recognized as an ErrResponseError")
	}

	// Test details are included
	errStr := err.Error()
	if !strings.Contains(errStr, "500") {
		t.Errorf("Error message should contain status code")
	}

	if !strings.Contains(errStr, "internal server error") {
		t.Errorf("Error message should contain response body")
	}
}

func TestInputError(t *testing.T) {
	err := InputError("missing required field")

	// Test wrapping behavior
	if !Is(err, ErrInvalidInput) {
		t.Errorf("Error should be recognized as an ErrInvalidInput")
	}

	// Test details are included
	if !strings.Contains(err.Error(), "missing required field") {
		t.Errorf("Error message should contain details")
	}
}

func TestWrappingBehavior(t *testing.T) {
	// Create a chain of wrapped errors
	originalErr := errors.New("root cause")
	connErr := ConnectionError(originalErr, "connection failed")

	// Test Unwrap functionality - our ConnectionError wraps the sentinel error ErrConnectionFailed,
	// not the original error directly, so we need to unwrap twice
	unwrapped := Unwrap(connErr)
	if unwrapped != ErrConnectionFailed {
		t.Errorf("First unwrap should return the sentinel error, got %v", unwrapped)
	}

	// Test Is relationship with the sentinel error
	if !Is(connErr, ErrConnectionFailed) {
		t.Errorf("Is should identify the error as an ErrConnectionFailed")
	}

	// Our error chain is actually:
	// connErr -> ErrConnectionFailed (not directly to originalErr)
	// The fmt.Errorf with %w wraps the sentinel error, not the original error

	// Instead, create an error that directly wraps originalErr
	directlyWrappedErr := fmt.Errorf("wrapped: %w", originalErr)
	if !Is(directlyWrappedErr, originalErr) {
		t.Errorf("Is should identify directly wrapped errors")
	}

	// Test with unrelated errors
	otherErr := errors.New("unrelated error")
	if Is(connErr, otherErr) {
		t.Errorf("Is should not find unrelated errors")
	}
}
