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
