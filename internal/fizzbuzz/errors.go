package fizzbuzz

import (
	"errors"
	"fmt"
	"strconv"
)

// Errors returned while parsing a request (Params.Parse).
var (
	// ErrRequired is returned when a parameter is missing from the request.
	ErrRequired = errors.New("required")

	// ErrNotAnInteger is returned when a numeric query parameter cannot be parsed.
	ErrNotAnInteger = errors.New(`must be an integer`)
)

// Errors returned while validating a request (Generator.Generate).
var (
	// ErrMustBeBiggerThanZero is returned when int1, int2 or limit is zero or negative.
	ErrMustBeBiggerThanZero = errors.New("must be bigger than zero")

	// ErrStringMustNotBeEmpty is returned when str1 or str2 is empty.
	ErrStringMustNotBeEmpty = errors.New("must not be empty string")
)

// Errors returned when building a Generator (NewGenerator).
var (
	// ErrTooBig is returned when a configured maximum exceeds its hard ceiling.
	ErrTooBig = errors.New("too big")

	// ErrTooLow is returned when a configured maximum is zero or negative.
	ErrTooLow = errors.New("too low")
)

var (
	_ error = ExceededMaxValueError(0)
	_ error = ExceededMaxLengthError(0)
	_ error = (*ParamError)(nil)
)

// ExceededMaxValueError is returned when limit exceeds the Generator's maximum;
// its value is that maximum.
type ExceededMaxValueError int

// Error implements error interface.
func (e ExceededMaxValueError) Error() string {
	return "must not exceed max value " + strconv.Itoa(int(e))
}

// ExceededMaxLengthError is returned when str1 or str2 exceeds the Generator's
// maximum length in bytes; its value is that maximum.
type ExceededMaxLengthError int

// Error implements error interface.
func (e ExceededMaxLengthError) Error() string {
	return "must not exceed max string length " + strconv.Itoa(int(e))
}

// Field names a request parameter, as used in the API (see the Field* constants).
type Field string

// ParamError reports which parameter is invalid and why.
// Use errors.As to read Field, and errors.Is to match the reason (Err).
type ParamError struct {
	Field Field // the offending parameter
	Err   error // the reason: a sentinel or Exceeded*Error
}

// Error implements the error interface.
func (pe *ParamError) Error() string {
	return fmt.Sprintf("invalid field %q: %v", pe.Field, pe.Err)
}

// Unwrap returns the underlying reason, so errors.Is matches the sentinels.
func (pe *ParamError) Unwrap() error { return pe.Err }
