package fizzbuzz

import (
	"errors"
	"fmt"
	"strconv"
)

// Generator type.
type Generator struct {
	maxLimit        int
	maxStringLength int
}

// DefaultGenerator: ~1024 elements of ~128 bytes each per response.
func DefaultGenerator() *Generator {
	return &Generator{
		maxLimit:        defaultFizzBuzzMaxLimit,
		maxStringLength: defaultFizzBuzzMaxStringLength,
	}
}

// NewGenerator ctor.
func NewGenerator(maxLimit, maxStringLength int) (*Generator, error) {
	if maxLimit > 100_000 || maxStringLength > 256 {
		return nil, errors.New("limits too big")
	}

	return &Generator{maxLimit: maxLimit, maxStringLength: maxStringLength}, nil
}

var (
	// ErrMustBeBiggerThanZero public error.
	ErrMustBeBiggerThanZero = errors.New("must be bigger than zero")
	// ErrStringMustNotBeEmpty public error.
	ErrStringMustNotBeEmpty = errors.New("must not be empty string")
)

// Generate will build the fizzbuzz sequence based on the input parameters.
// It will return an error if the parameters are invalid.
func (g *Generator) Generate(p Params) ([]string, error) {
	if err := g.validateParameters(&p); err != nil {
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

func (g *Generator) validateParameters(p *Params) error {
	switch {
	case p.Int1 <= 0:
		return &ParamError{Field: FieldInt1, Err: ErrMustBeBiggerThanZero}
	case p.Int2 <= 0:
		return &ParamError{Field: FieldInt2, Err: ErrMustBeBiggerThanZero}
	case p.Limit <= 0:
		return &ParamError{Field: FieldLimit, Err: ErrMustBeBiggerThanZero}
	case p.Str1 == "":
		return &ParamError{Field: FieldStr1, Err: ErrStringMustNotBeEmpty}
	case p.Str2 == "":
		return &ParamError{Field: FieldStr2, Err: ErrStringMustNotBeEmpty}
	case p.Limit > g.maxLimit:
		return &ParamError{
			Field: FieldLimit,
			Err:   ExceededMaxValueError(g.maxLimit),
		}
	case len(p.Str1) > g.maxStringLength:
		return &ParamError{
			Field: FieldStr1,
			Err:   ExceededMaxLengthError(g.maxStringLength),
		}
	case len(p.Str2) > g.maxStringLength:
		return &ParamError{
			Field: FieldStr2,
			Err:   ExceededMaxLengthError(g.maxStringLength),
		}
	}

	return nil
}
