// Package errors provides custom error types and error handling utilities
// for the cloudlog library.
package errors

import (
	"errors"
	"fmt"
)

// Common errors that can be used for comparison
var (
	// ErrInvalidFormat indicates an issue with log formatting
	ErrInvalidFormat = errors.New("invalid log format")
	// ErrConnectionFailed indicates a failure to connect to the log service
	ErrConnectionFailed = errors.New("connection to log service failed")
	// ErrResponseError indicates an error response from the log service
	ErrResponseError = errors.New("received error response from log service")
	// ErrInvalidInput indicates invalid input parameters
	ErrInvalidInput = errors.New("invalid input parameters")
	// ErrBufferFull indicates that the async buffer is full
	ErrBufferFull = errors.New("log buffer is full")
	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("operation timed out")
	// ErrShutdown indicates a problem during shutdown
	ErrShutdown = errors.New("error during shutdown")
)

// FormatError wraps an error related to log formatting
// The error chain will be: returned error -> ErrInvalidFormat -> originalError
func FormatError(err error, details string) error {
	// Using fmt.Errorf with %w to properly wrap the sentinel error
	return fmt.Errorf("%w: %s: %v", ErrInvalidFormat, details, err)
}

// ConnectionError wraps an error related to connection issues
// The error chain will be: returned error -> ErrConnectionFailed -> originalError
func ConnectionError(err error, details string) error {
	return fmt.Errorf("%w: %s: %v", ErrConnectionFailed, details, err)
}

// ResponseError wraps an error related to service responses
func ResponseError(statusCode int, body string) error {
	return fmt.Errorf("%w: status code %d: %s", ErrResponseError, statusCode, body)
}

// InputError wraps an error related to invalid inputs
func InputError(details string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, details)
}

// BufferFullError wraps an error to indicate buffer full condition
func BufferFullError(err error) error {
	return wrapError(err, ErrBufferFull, "log buffer is full")
}

// TimeoutError wraps an error to indicate timeout
func TimeoutError(err error) error {
	return wrapError(err, ErrTimeout, "operation timed out")
}

// ShutdownError wraps an error to indicate shutdown problems
func ShutdownError(err error) error {
	return wrapError(err, ErrShutdown, "error during shutdown")
}

// IsBufferFullError returns true if error is related to buffer full condition
func IsBufferFullError(err error) bool {
	return errors.Is(err, ErrBufferFull)
}

// IsTimeoutError returns true if error is related to timeout
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsShutdownError returns true if error is related to shutdown problems
func IsShutdownError(err error) bool {
	return errors.Is(err, ErrShutdown)
}

// Create simple errors (when not wrapping another error)
func NewError(msg string) error {
	return errors.New(msg)
}

// Is provides a convenience wrapper around errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As provides a convenience wrapper around errors.As
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Unwrap provides a convenience wrapper around errors.Unwrap
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// WrapError wraps an error with additional message
func WrapError(err error, msg string) error {
	if err == nil {
		// Return a new error instead of nil when given nil error
		return fmt.Errorf("%s", msg)
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// wrapError is an internal helper for wrapping errors with sentinel error types
func wrapError(err error, sentinel error, msg string) error {
	if err == nil {
		return fmt.Errorf("%w: %s", sentinel, msg)
	}
	return fmt.Errorf("%w: %s: %v", sentinel, msg, err)
}
