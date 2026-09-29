package fizzbuzz

import (
	"fmt"
	"strconv"
)

var (
	_ error = ExceededMaxValueError(0)
	_ error = ExceededMaxLengthError(0)
	_ error = (*ParamError)(nil)
)

// ExceededMaxValueError error.
type ExceededMaxValueError int

// Error implements error interface.
func (e ExceededMaxValueError) Error() string {
	return "must not exceed max value " + strconv.Itoa(int(e))
}

// ExceededMaxLengthError error.
type ExceededMaxLengthError int

// Error implements error interface.
func (e ExceededMaxLengthError) Error() string {
	return "must not exceed max string length " + strconv.Itoa(int(e))
}

// Field represents a param field name.
type Field string

// ParamError reports which parameter is invalid and why. It implements `Unwrap() error` interface.
type ParamError struct {
	Field Field
	Err   error
}

// Error implements error interface.
func (pe *ParamError) Error() string {
	return fmt.Sprintf("invalid field %q: %v", pe.Field, pe.Err)
}

// Unwrap returns the underlying reason, os errors.Is matches the sentinels.
func (pe *ParamError) Unwrap() error { return pe.Err }
