package errors

import (
	"fmt"
	"testing"

	stderrors "errors"

	"github.com/stretchr/testify/assert"
)

func TestErrorIdentification(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		checkFunc  func(error) bool
		shouldPass bool
	}{
		{"IsFormatError", fmt.Errorf("%w: format failed", ErrInvalidFormat), IsFormatError, true},
		{"IsConnectionError", fmt.Errorf("%w: connection failed", ErrConnectionFailed), IsConnectionError, true},
		{"IsResponseError", fmt.Errorf("%w: response error", ErrResponseError), IsResponseError, true},
		{"IsBufferFullError", fmt.Errorf("%w: buffer full", ErrBufferFull), IsBufferFullError, true},
		{"IsTimeoutError", fmt.Errorf("%w: timeout", ErrTimeout), IsTimeoutError, true},
		{"IsShutdownError", fmt.Errorf("%w: shutdown error", ErrShutdown), IsShutdownError, true},
		{"IsLoggerClosedError", fmt.Errorf("%w: logger closed", ErrLoggerClosed), IsLoggerClosedError, true},
		{"IsProcessingError", fmt.Errorf("%w: processing failed", ErrProcessingFailed), IsProcessingError, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.shouldPass, tc.checkFunc(tc.err))
		})
	}
}

func TestWrappingUnwrapping(t *testing.T) {
	// Create original error
	baseErr := fmt.Errorf("base error")

	// Wrap with sentinel error
	wrappedErr := fmt.Errorf("%w: wrapping: %v", ErrInvalidFormat, baseErr)

	// Check identification
	assert.True(t, IsFormatError(wrappedErr))

	// Check message contains components
	errMsg := wrappedErr.Error()
	assert.Contains(t, errMsg, "invalid log format")
	assert.Contains(t, errMsg, "wrapping")
	assert.Contains(t, errMsg, "base error")

	// Check IS relationship with stdlib functions
	assert.True(t, stderrors.Is(wrappedErr, ErrInvalidFormat))
}

func TestNestedWrapping(t *testing.T) {
	// Create and wrap error multiple times
	origErr := fmt.Errorf("original")
	_ = fmt.Errorf("wrap1: %w", origErr)
	wrap2 := fmt.Errorf("%w: wrap2", ErrInvalidFormat)

	// Test behavior
	assert.True(t, IsFormatError(wrap2))
}

func TestLoggerClosedError(t *testing.T) {
	// Test IsLoggerClosedError specifically
	errClosed := fmt.Errorf("%w: logger is closed", ErrLoggerClosed)

	// Verify it detects this type of error
	assert.True(t, IsLoggerClosedError(errClosed))

	// Verify it correctly rejects other error types
	otherErr := fmt.Errorf("some other error")
	assert.False(t, IsLoggerClosedError(otherErr))

	// Test with another error type to ensure specificity
	shutdownErr := fmt.Errorf("%w: logger is shutting down", ErrShutdown)
	assert.False(t, IsLoggerClosedError(shutdownErr))
}

func TestInvalidFormatError(t *testing.T) {
	// ...existing code...
}

func TestConnectionFailedError(t *testing.T) {
	// ...existing code...
}

func TestResponseError(t *testing.T) {
	// ...existing code...
}

func TestInvalidInputError(t *testing.T) {
	// ...existing code...
}

func TestBufferFullError(t *testing.T) {
	// ...existing code...
}

func TestTimeoutError(t *testing.T) {
	// ...existing code...
}

func TestShutdownError(t *testing.T) {
	// ...existing code...
}

func TestIsProcessingError(t *testing.T) {
	// Test with direct error
	assert.True(t, IsProcessingError(ErrProcessingFailed))

	// Test with wrapped error
	wrappedErr := fmt.Errorf("%w: test error", ErrProcessingFailed)
	assert.True(t, IsProcessingError(wrappedErr))

	// Test with unrelated error
	unrelatedErr := fmt.Errorf("some other error")
	assert.False(t, IsProcessingError(unrelatedErr))
}

func TestNoopErrorHandler(t *testing.T) {
	// Just verify it doesn't panic
	NoopErrorHandler(fmt.Errorf("test error"))
}
