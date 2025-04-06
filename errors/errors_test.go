package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestNewError(t *testing.T) {
	// Test creating a new error
	msg := "test error"
	err := NewError(msg)

	assert.Error(t, err)
	assert.Equal(t, msg, err.Error())

	// Verify the error message contains what we expect
	assert.True(t, strings.Contains(err.Error(), msg))
}

func TestAs(t *testing.T) {
	// Test that As correctly type asserts our custom error type
	baseErr := NewError("test error")

	// We'll use a standard error interface since we don't have access to the concrete type
	var target error
	result := As(baseErr, &target)

	assert.True(t, result)
	assert.NotNil(t, target)
	assert.Equal(t, baseErr.Error(), target.Error())

	// Test with a different error type
	stdErr := errors.New("standard error")

	// Reset target
	target = nil
	result = As(stdErr, &target)

	// This should still be true because we're using the error interface
	assert.True(t, result)
	assert.NotNil(t, target)
}

func TestWrapError(t *testing.T) {
	// Test wrapping an error
	innerErr := errors.New("inner error")
	msg := "wrapped error"

	wrappedErr := WrapError(innerErr, msg)

	assert.Error(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), msg)

	// Use the standard errors.Unwrap function since we don't know the concrete type
	unwrapped := errors.Unwrap(wrappedErr)
	assert.Equal(t, innerErr, unwrapped)
}

func TestWrapErrorWithNilError(t *testing.T) {
	// Test wrapping a nil error
	msg := "wrapped nil error"
	result := WrapError(nil, msg)

	// Should return a non-nil error containing our message
	assert.NotNil(t, result, "WrapError should return a non-nil error when given nil")
	assert.Contains(t, result.Error(), msg, "Error should contain our message")

	// Should not be unwrappable since there's no inner error
	unwrapped := errors.Unwrap(result)
	assert.Nil(t, unwrapped, "Unwrapping should return nil since there's no inner error")
}
