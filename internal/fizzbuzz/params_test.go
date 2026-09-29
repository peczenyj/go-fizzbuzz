package fizzbuzz_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

func TestConstants(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		constant string
		value    string
	}{
		{
			constant: fizzbuzz.FieldInt1,
			value:    "int1",
		},
		{
			constant: fizzbuzz.FieldInt2,
			value:    "int2",
		},
		{
			constant: fizzbuzz.FieldLimit,
			value:    "limit",
		},
		{
			constant: fizzbuzz.FieldStr1,
			value:    "str1",
		},
		{
			constant: fizzbuzz.FieldStr2,
			value:    "str2",
		},
	}

	for _, tc := range testcases {
		t.Run("check constant "+tc.constant, func(t *testing.T) {
			t.Parallel()

			if tc.constant != tc.value {
				t.Fatalf("unexpected constant value (got: %v, expected: %v)", tc.constant, tc.value)
			}
		})
	}
}

func TestParseParams(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label    string
		query    string
		expected fizzbuzz.Params
		fieldErr fizzbuzz.Field
		err      error
	}{
		{
			label:    "should return error if int1 is omitted",
			query:    "",
			fieldErr: fizzbuzz.FieldInt1,
			err:      fizzbuzz.ErrRequired,
		},
		{
			label:    "should return error if int2 is omitted",
			query:    "int1=3",
			fieldErr: fizzbuzz.FieldInt2,
			err:      fizzbuzz.ErrRequired,
		},
		{
			label:    "should return error if limit is omitted",
			query:    "int1=3&int2=5",
			fieldErr: fizzbuzz.FieldLimit,
			err:      fizzbuzz.ErrRequired,
		},
		{
			label:    "should return error if str1 is omitted",
			query:    "int1=3&int2=5&limit=100",
			fieldErr: fizzbuzz.FieldStr1,
			err:      fizzbuzz.ErrRequired,
		},
		{
			label:    "should return error if str2 is omitted",
			query:    "int1=3&int2=5&limit=100&str1=fizz",
			fieldErr: fizzbuzz.FieldStr2,
			err:      fizzbuzz.ErrRequired,
		},
		{
			label:    "should parse query with success",
			query:    "int1=3&int2=5&limit=100&str1=fizz&str2=buzz",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "should ignore non supported query strings",
			query:    "int1=3&int2=5&limit=100&str1=fizz&str2=buzz&lol=hehe",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "should return default params with limit of 15",
			query:    "int1=3&int2=5&limit=15&str1=fizz&str2=buzz",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "first value wins",
			query:    "int1=2&int1=3&int2=5&limit=15&str1=fizz&str2=buzz",
			expected: fizzbuzz.Params{Int1: 2, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "should parse explicit positive number (+ must be url encoded)",
			query:    "int1=%2B3&int2=5&limit=15&str1=fizz&str2=buzz",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "should parse negative number",
			query:    "int1=3&int2=-5&limit=15&str1=fizz&str2=buzz",
			expected: fizzbuzz.Params{Int1: 3, Int2: -5, Limit: 15, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "should parse zero as integer",
			query:    "int1=3&int2=5&limit=0&str1=fizz&str2=buzz",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 0, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "should return default params with limit of 15 and alternate strings",
			query:    "int1=3&int2=5&limit=15&str1=abc&str2=xyz",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "abc", Str2: "xyz"},
		},
		{
			label:    "should return default params with limit of 1",
			query:    "int1=3&int2=5&limit=1&str1=fizz&str2=buzz",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 1, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "should return error if int1 is not a number",
			query:    "int1=lol&int2=5&limit=15&str1=fizz&str2=buzz",
			fieldErr: `int1`,
			err:      fizzbuzz.ErrNotAnInteger,
		},
		{
			label:    "should return error if int1 is present but empty",
			query:    "int1=&int2=5&limit=15&str1=fizz&str2=buzz",
			fieldErr: `int1`,
			err:      fizzbuzz.ErrNotAnInteger,
		},
		{
			label:    "should return error if int1 is a number but not trimmed (with a space in front, urlencoded)",
			query:    "int1=%203&int2=5&limit=15&str1=fizz&str2=buzz",
			fieldErr: `int1`,
			err:      fizzbuzz.ErrNotAnInteger,
		},
		{
			label:    "should return error if int1 is a number but not trimmed (with a space in front, using + sign)",
			query:    "int1=+3&int2=5&limit=15&str1=fizz&str2=buzz",
			fieldErr: `int1`,
			err:      fizzbuzz.ErrNotAnInteger,
		},
		{
			label:    "should return error if int1 overflows",
			query:    "int1=99999999999999999999&int2=5&limit=15&str1=fizz&str2=buzz",
			fieldErr: `int1`,
			err:      fizzbuzz.ErrNotAnInteger,
		},
		{
			label:    "should return error if int2 is not a number",
			query:    "int1=3&int2=hehe&limit=15&str1=fizz&str2=buzz",
			fieldErr: `int2`,
			err:      fizzbuzz.ErrNotAnInteger,
		},
		{
			label:    "should return error if limit is not a number",
			query:    "int1=3&int2=5&limit=ops&str1=fizz&str2=buzz",
			fieldErr: `limit`,
			err:      fizzbuzz.ErrNotAnInteger,
		},
		{
			label:    "should return param with explicit empty str1",
			query:    "int1=3&int2=5&limit=15&str1=&str2=buzz",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "", Str2: "buzz"},
		},
		{
			label:    "should return param with explicit empty str2",
			query:    "int1=3&int2=5&limit=15&str1=fizz&str2=",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: ""},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			values, err := url.ParseQuery(tc.query)
			if err != nil {
				t.Fatalf("unexpected error while parse query string: %v", err)
			}

			var params fizzbuzz.Params

			err = params.Parse(values)
			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("unexpected error (got: %v, expected: %v)", err, tc.err)
				}

				pe, ok := errors.AsType[*fizzbuzz.ParamError](err)
				if !ok {
					t.Fatalf("unexpected error type (got: %v, expected: *fizzbuzz.ParamError)", err)
				}

				if pe.Field != tc.fieldErr {
					t.Fatalf("unexpected error field (got: %v, expected: %v)", pe.Field, tc.fieldErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error (got: %v, expected: nil)", err)
			}

			if params != tc.expected {
				t.Fatalf("unexpected params (got: %v, expected: %v)", params, tc.expected)
			}
		})
	}
}

var _ fizzbuzz.QueryValues = mapSource(nil)

type mapSource map[string]string

func (m mapSource) Get(key string) string {
	return m[key]
}

func (m mapSource) Has(key string) bool {
	_, found := m[key]

	return found
}

func TestParseParamsWithStub(t *testing.T) {
	t.Parallel()

	source := mapSource{"int1": "6", "int2": "5", "limit": "100", "str1": "fizz", "str2": "buzz"}

	var params fizzbuzz.Params

	err := params.Parse(source)
	if err != nil {
		t.Fatalf("unexpected error while parse map source stub: %v", err)
	}

	expected := fizzbuzz.Params{Int1: 6, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}

	if params != expected {
		t.Fatalf("unexpected params (got: %v, expected: %v)", params, expected)
	}
}

func FuzzParseParams(f *testing.F) {
	f.Add("int1=3&int2=5&limit=15&str1=fizz&str2=buzz")

	f.Fuzz(func(t *testing.T, raw string) {
		values, err := url.ParseQuery(raw)
		if err != nil {
			t.Skip() // not a valid query string; not ParseParams' concern
		}

		var params fizzbuzz.Params

		err = params.Parse(values)
		if err != nil && !errors.Is(err, fizzbuzz.ErrNotAnInteger) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
