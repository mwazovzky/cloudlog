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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.shouldPass, tc.checkFunc(tc.err))
		})
	}
}

func TestWrappingUnwrapping(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	wrappedErr := fmt.Errorf("%w: wrapping: %v", ErrInvalidFormat, baseErr)

	assert.True(t, IsFormatError(wrappedErr))

	errMsg := wrappedErr.Error()
	assert.Contains(t, errMsg, "invalid log format")
	assert.Contains(t, errMsg, "wrapping")
	assert.Contains(t, errMsg, "base error")

	assert.True(t, stderrors.Is(wrappedErr, ErrInvalidFormat))
}

func TestNestedWrapping(t *testing.T) {
	origErr := fmt.Errorf("original")
	_ = fmt.Errorf("wrap1: %w", origErr)
	wrap2 := fmt.Errorf("%w: wrap2", ErrInvalidFormat)

	assert.True(t, IsFormatError(wrap2))
}
