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

// Param represents the fizzzbuzz generate function parameters.
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

// SetDefaults method will set the initial default values.
func (p *Params) SetDefaults() {
	p.Int1 = defaultFizzDivisor
	p.Int2 = defaultBuzzDivisor
	p.Limit = defaultFizzBuzzLimit
	p.Str1 = defaultFizzLabel
	p.Str2 = defaultBuzzLabel
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
		return fmt.Errorf("invalid field 'Int1': %w", ErrMustBeBiggerThanZero)
	case p.Int2 <= 0:
		return fmt.Errorf("invalid field 'Int2': %w", ErrMustBeBiggerThanZero)
	case p.Limit <= 0:
		return fmt.Errorf("invalid field 'Limit': %w", ErrMustBeBiggerThanZero)
	case p.Limit > maxFizzBuzzLimit:
		return fmt.Errorf("invalid field 'Limit': %w", ErrMustNotExceedMaxValue)
	case p.Str1 == "":
		return fmt.Errorf("invalid field 'Str1': %w", ErrStringMustNotBeEmpty)
	case len(p.Str1) > maxFizzBuzzStringLength:
		return fmt.Errorf("invalid field 'Str1': %w", ErrStringMustNotExceedMaxLength)
	case p.Str2 == "":
		return fmt.Errorf("invalid field 'Str2': %w", ErrStringMustNotBeEmpty)
	case len(p.Str2) > maxFizzBuzzStringLength:
		return fmt.Errorf("invalid field 'Str2': %w", ErrStringMustNotExceedMaxLength)
	}

	return nil
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
