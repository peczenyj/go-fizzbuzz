package fizzbuzz

import (
	"strconv"
)

// Parameter names, as used in the query string and in ParamError.Field.
const (
	FieldInt1  = `int1`
	FieldInt2  = `int2`
	FieldLimit = `limit`
	FieldStr1  = `str1`
	FieldStr2  = `str2`

	defaultFizzBuzzMaxLimit        = 1024
	defaultFizzBuzzMaxStringLength = 64
)

// Params holds the five inputs of a fizzbuzz request. It is a comparable value
// type and can be used as a map key (e.g. for request statistics).
type Params struct {
	Int1  int
	Int2  int
	Limit int
	Str1  string
	Str2  string
}

// QueryValues is the minimal read-only view Parse needs; url.Values satisfies it.
type QueryValues interface {
	Has(key string) bool
	Get(key string) string
}

// Parse fills p from query. All five parameters are required; int1, int2 and
// limit must be integers. It returns a *ParamError wrapping ErrRequired or
// ErrNotAnInteger; on error, p's contents are unspecified.
// Parse does not check domain rules or size limits: Generator.Generate does.
func (p *Params) Parse(query QueryValues) (err error) {
	if !query.Has(FieldInt1) {
		return &ParamError{Field: FieldInt1, Err: ErrRequired}
	}

	p.Int1, err = strconv.Atoi(query.Get(FieldInt1))
	if err != nil {
		return &ParamError{Field: FieldInt1, Err: ErrNotAnInteger}
	}

	if !query.Has(FieldInt2) {
		return &ParamError{Field: FieldInt2, Err: ErrRequired}
	}

	p.Int2, err = strconv.Atoi(query.Get(FieldInt2))
	if err != nil {
		return &ParamError{Field: FieldInt2, Err: ErrNotAnInteger}
	}

	if !query.Has(FieldLimit) {
		return &ParamError{Field: FieldLimit, Err: ErrRequired}
	}

	p.Limit, err = strconv.Atoi(query.Get(FieldLimit))
	if err != nil {
		return &ParamError{Field: FieldLimit, Err: ErrNotAnInteger}
	}

	if !query.Has(FieldStr1) {
		return &ParamError{Field: FieldStr1, Err: ErrRequired}
	}

	p.Str1 = query.Get(FieldStr1)

	if !query.Has(FieldStr2) {
		return &ParamError{Field: FieldStr2, Err: ErrRequired}
	}

	p.Str2 = query.Get(FieldStr2)

	return nil
}
