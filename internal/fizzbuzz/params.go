package fizzbuzz

import (
	"errors"
	"strconv"
)

const (
	FieldInt1  = `int1`
	FieldInt2  = `int2`
	FieldLimit = `limit`
	FieldStr1  = `str1`
	FieldStr2  = `str2`

	defaultFizzBuzzMaxLimit        = 1024
	defaultFizzBuzzMaxStringLength = 64
)

var ErrParamRequired = errors.New("parameter required")

// Params represents the fizzbuzz generate function parameters.
type Params struct {
	Int1  int
	Int2  int
	Limit int
	Str1  string
	Str2  string
}

// ErrNotAnInteger is returned when a numeric query parameter cannot be parsed.
var ErrNotAnInteger = errors.New(`must be an integer`)

// QueryValues interface.
type QueryValues interface {
	Has(key string) bool
	Get(key string) string
}

// Parse extract params from a query values interface, like url.Values from query string.
func (p *Params) Parse(query QueryValues) (err error) {
	if query.Has(FieldInt1) {
		p.Int1, err = strconv.Atoi(query.Get(FieldInt1))
		if err != nil {
			return &ParamError{Field: FieldInt1, Err: ErrNotAnInteger}
		}
	} else {
		return &ParamError{Field: FieldInt1, Err: ErrParamRequired}
	}

	if query.Has(FieldInt2) {
		p.Int2, err = strconv.Atoi(query.Get(FieldInt2))
		if err != nil {
			return &ParamError{Field: FieldInt2, Err: ErrNotAnInteger}
		}
	} else {
		return &ParamError{Field: FieldInt2, Err: ErrParamRequired}
	}

	if query.Has(FieldLimit) {
		p.Limit, err = strconv.Atoi(query.Get(FieldLimit))
		if err != nil {
			return &ParamError{Field: FieldLimit, Err: ErrNotAnInteger}
		}
	} else {
		return &ParamError{Field: FieldLimit, Err: ErrParamRequired}
	}

	if query.Has(FieldStr1) {
		p.Str1 = query.Get(FieldStr1)
	} else {
		return &ParamError{Field: FieldStr1, Err: ErrParamRequired}
	}

	if query.Has(FieldStr2) {
		p.Str2 = query.Get(FieldStr2)
	} else {
		return &ParamError{Field: FieldStr2, Err: ErrParamRequired}
	}

	return nil
}
