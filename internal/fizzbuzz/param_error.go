package fizzbuzz

import (
	"fmt"
)

var _ error = (*ParamError)(nil)

// ParamError reports witch parameter is invalid and why. It implements `Unwrap() error` interface.
type ParamError struct {
	Field Field
	Err   error
}

// Error method
func (pe *ParamError) Error() string {
	return fmt.Sprintf("invalid field %q: %v", pe.Field, pe.Err)
}

// Unwrap method
func (pe *ParamError) Unwrap() error { return pe.Err }
