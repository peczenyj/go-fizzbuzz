package fizzbuzz

import (
	"fmt"
)

var _ error = (*ParamError)(nil)

type ParamError struct {
	Field Field
	Err   error
}

func (pe *ParamError) Error() string {
	return fmt.Sprintf("invalid field %q: %v", pe.Field, pe.Err)
}

func (pe *ParamError) Unwrap() error { return pe.Err }
