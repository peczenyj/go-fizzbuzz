package fizzbuzz

import (
	"fmt"
	"strconv"
)

// Generator produces fizzbuzz sequences within a size policy fixed at
// construction: a maximum limit and a maximum label length. A Generator is
// immutable and safe for concurrent use.
type Generator struct {
	maxLimit        int
	maxStringLength int
}

// DefaultGenerator returns a Generator allowing up to 1024 elements and
// 64-byte labels: at most 1024 × 128 bytes of labels. JSON escaping can
// expand each byte up to 6× (e.g. "<" → "\u003c"), so the largest response
// is about 790 KB (measured: 789 506 bytes).
func DefaultGenerator() *Generator {
	return &Generator{
		maxLimit:        DefaultMaxLimit,
		maxStringLength: DefaultMaxStringLength,
	}
}

// Hard ceilings for any Generator. Worst case per call:
// maxLimitThreshold × 2 × maxStringLengthThreshold ≈ 100 000 × 510 B ≈ 51 MB
// of labels in memory; once JSON-encoded (up to 6× for escaped bytes), the
// response can reach ≈ 300 MB.
const (
	maxLimitThreshold        = 100_000
	maxStringLengthThreshold = 255
)

// NewGenerator returns a Generator allowing up to maxLimit elements and labels
// of up to maxStringLength bytes. Both must be between 1 and the hard ceilings
// (100 000 and 255); otherwise the error wraps ErrTooLow or ErrTooBig.
func NewGenerator(maxLimit, maxStringLength int) (*Generator, error) {
	if maxLimit > maxLimitThreshold {
		return nil, fmt.Errorf("invalid max limit %d: %w", maxLimit, ErrTooBig)
	}

	if maxLimit <= 0 {
		return nil, fmt.Errorf("invalid max limit %d: %w", maxLimit, ErrTooLow)
	}

	if maxStringLength > maxStringLengthThreshold {
		return nil, fmt.Errorf("invalid max string length %d: %w", maxStringLength, ErrTooBig)
	}

	if maxStringLength <= 0 {
		return nil, fmt.Errorf("invalid max string length %d: %w", maxStringLength, ErrTooLow)
	}

	return &Generator{maxLimit: maxLimit, maxStringLength: maxStringLength}, nil
}

// Generate returns the fizzbuzz sequence for p: numbers from 1 to p.Limit,
// with multiples of p.Int1 replaced by p.Str1, multiples of p.Int2 by p.Str2,
// and multiples of both by p.Str1+p.Str2.
// It first checks p against the domain rules, then against g's size policy;
// any violation is a *ParamError.
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
