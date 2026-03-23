// Package errors provides error definitions and helper functions for the cloudlog library.
package errors

import (
	"errors"
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

	// ErrBufferFull indicates the async sender buffer is full
	ErrBufferFull = errors.New("log buffer is full")

	// ErrSenderClosed indicates the sender has been closed
	ErrSenderClosed = errors.New("sender is closed")
)

// Error type check functions
func IsFormatError(err error) bool {
	return errors.Is(err, ErrInvalidFormat)
}

func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}

func IsResponseError(err error) bool {
	return errors.Is(err, ErrResponseError)
}
