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

	// ErrBufferFull indicates that the async buffer is full
	ErrBufferFull = errors.New("log buffer is full")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("operation timed out")

	// ErrShutdown indicates a problem during shutdown
	ErrShutdown = errors.New("error during shutdown")

	// ErrLoggerClosed indicates an attempt to log after logger is closed
	ErrLoggerClosed = errors.New("logger is closed")

	// ErrProcessingFailed indicates an error while processing logs asynchronously
	ErrProcessingFailed = errors.New("asynchronous log processing failed")
)

// ErrorHandler defines a function type for handling errors that occur during
// asynchronous operations where errors can't be returned to the caller
type ErrorHandler func(error)

// NoopErrorHandler is a default error handler that does nothing
func NoopErrorHandler(_ error) {}

// Is checks if an error is of a specific type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Error type check functions - simplified API that uses standard errors package
func IsFormatError(err error) bool {
	return errors.Is(err, ErrInvalidFormat)
}

func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}

func IsResponseError(err error) bool {
	return errors.Is(err, ErrResponseError)
}

func IsInputError(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

func IsBufferFullError(err error) bool {
	return errors.Is(err, ErrBufferFull)
}

func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}

func IsShutdownError(err error) bool {
	return errors.Is(err, ErrShutdown)
}

func IsLoggerClosedError(err error) bool {
	return errors.Is(err, ErrLoggerClosed)
}

func IsProcessingError(err error) bool {
	return errors.Is(err, ErrProcessingFailed)
}
