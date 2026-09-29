package fizzbuzz

import (
	"fmt"
)

var _ error = (*ParamError)(nil)

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
