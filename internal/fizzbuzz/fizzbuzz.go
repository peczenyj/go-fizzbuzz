package fizzbuzz

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	defaultFizzLabel        = "fizz"
	defaultBuzzLabel        = "buzz"
	defaultFizzDivisor      = 3
	defaultBuzzDivisor      = 5
	defaultFizzBuzzLimit    = 100
	maxFizzBuzzLimit        = 1024
	maxFizzBuzzStringLength = 64
)

// Params represents the fizzbuzz generate function parameters.
type Params struct {
	Int1  int
	Int2  int
	Limit int
	Str1  string
	Str2  string
}

// DefaultParams returns a Params struct filled with all default values.
func DefaultParams() Params {
	return Params{
		Int1:  defaultFizzDivisor,
		Int2:  defaultBuzzDivisor,
		Limit: defaultFizzBuzzLimit,
		Str1:  defaultFizzLabel,
		Str2:  defaultBuzzLabel,
	}
}

var (
	// ErrMustBeBiggerThanZero public error.
	ErrMustBeBiggerThanZero = errors.New("must be bigger than zero")
	// ErrMustNotExceedMaxValue public error.
	ErrMustNotExceedMaxValue = fmt.Errorf("must not exceed max value %d", maxFizzBuzzLimit)
	// ErrStringMustNotBeEmpty public error.
	ErrStringMustNotBeEmpty = errors.New("must not be empty string")
	// ErrStringMustNotExceedMaxLength public error.
	ErrStringMustNotExceedMaxLength = fmt.Errorf("must not exceed max string length %d", maxFizzBuzzStringLength)
)

// Validate inspect params values.
func (p *Params) Validate() error {
	switch {
	case p.Int1 <= 0:
		return &ParamError{Field: FieldInt1, Err: ErrMustBeBiggerThanZero}
	case p.Int2 <= 0:
		return &ParamError{Field: FieldInt2, Err: ErrMustBeBiggerThanZero}
	case p.Limit <= 0:
		return &ParamError{Field: FieldLimit, Err: ErrMustBeBiggerThanZero}
	case p.Limit > maxFizzBuzzLimit:
		return &ParamError{Field: FieldLimit, Err: ErrMustNotExceedMaxValue}
	case p.Str1 == "":
		return &ParamError{Field: FieldStr1, Err: ErrStringMustNotBeEmpty}
	case len(p.Str1) > maxFizzBuzzStringLength:
		return &ParamError{Field: FieldStr1, Err: ErrStringMustNotExceedMaxLength}
	case p.Str2 == "":
		return &ParamError{Field: FieldStr2, Err: ErrStringMustNotBeEmpty}
	case len(p.Str2) > maxFizzBuzzStringLength:
		return &ParamError{Field: FieldStr2, Err: ErrStringMustNotExceedMaxLength}
	}

	return nil
}

// ErrNotAnInteger is returned when a numeric query parameter cannot be parsed.
var ErrNotAnInteger = errors.New(`must be an integer`)

// QueryValues interface.
type QueryValues interface {
	Has(key string) bool
	Get(key string) string
}

// ParseParams extract params from query string.
func ParseParams(query QueryValues) (params Params, err error) {
	params = DefaultParams()

	if query.Has("int1") {
		params.Int1, err = strconv.Atoi(query.Get("int1"))
		if err != nil {
			return params, &ParamError{Field: FieldInt1, Err: ErrNotAnInteger}
		}
	}

	if query.Has("int2") {
		params.Int2, err = strconv.Atoi(query.Get("int2"))
		if err != nil {
			return params, &ParamError{Field: FieldInt2, Err: ErrNotAnInteger}
		}
	}

	if query.Has("limit") {
		params.Limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			return params, &ParamError{Field: FieldLimit, Err: ErrNotAnInteger}
		}
	}

	if query.Has("str1") {
		params.Str1 = query.Get("str1")
	}

	if query.Has("str2") {
		params.Str2 = query.Get("str2")
	}

	return params, nil
}

// Generate will build the fizzbuzz sequence based on the input parameters.
// It will return an error if the parameters are invalid.
func Generate(p Params) ([]string, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("unable to generate fizzbuzz sequence: %w", err)
	}

	result := make([]string, 0, p.Limit)

	fizzBuzz := p.Str1 + p.Str2

	for i := 1; i <= p.Limit; i++ {
		switch {
		case i%p.Int1 == 0 && i%p.Int2 == 0:
			result = append(result, fizzBuzz)
		case i%p.Int1 == 0:
			result = append(result, p.Str1)
		case i%p.Int2 == 0:
			result = append(result, p.Str2)
		default:
			result = append(result, strconv.Itoa(i))
		}
	}

	return result, nil
}
