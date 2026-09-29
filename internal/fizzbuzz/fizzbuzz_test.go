package fizzbuzz_test

import (
	"errors"
	"net/url"
	"slices"
	"strconv"
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

func TestDefaultParams(t *testing.T) {
	t.Parallel()

	params := fizzbuzz.DefaultParams()

	expected := fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}

	if params != expected {
		t.Fatalf("unexpected default params (got: %v, expected: %v)", params, expected)
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
			label:    "should return default params if no query string is present",
			query:    "",
			expected: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"},
		},
		{
			label:    "should ignore non supported query strings",
			query:    "lol=hehe",
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

			params, err := fizzbuzz.ParseParams(values)
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

	source := mapSource{"int1": "6"}

	params, err := fizzbuzz.ParseParams(source)
	if err != nil {
		t.Fatalf("unexpected error while parse map source stub: %v", err)
	}

	expected := fizzbuzz.Params{Int1: 6, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}

	if params != expected {
		t.Fatalf("unexpected params (got: %v, expected: %v)", params, expected)
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label     string
		params    fizzbuzz.Params
		targetErr error
		fieldErr  fizzbuzz.Field
		expected  []string

		skipExpectedCheck bool
	}{
		{
			label:  "should return first 100 elements of fizzbuzz with default parameter",
			params: fizzbuzz.DefaultParams(),
			expected: []string{
				"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz",
				"16", "17", "fizz", "19", "buzz", "fizz", "22", "23", "fizz", "buzz", "26", "fizz", "28", "29", "fizzbuzz",
				"31", "32", "fizz", "34", "buzz", "fizz", "37", "38", "fizz", "buzz", "41", "fizz", "43", "44", "fizzbuzz",
				"46", "47", "fizz", "49", "buzz", "fizz", "52", "53", "fizz", "buzz", "56", "fizz", "58", "59", "fizzbuzz",
				"61", "62", "fizz", "64", "buzz", "fizz", "67", "68", "fizz", "buzz", "71", "fizz", "73", "74", "fizzbuzz",
				"76", "77", "fizz", "79", "buzz", "fizz", "82", "83", "fizz", "buzz", "86", "fizz", "88", "89", "fizzbuzz",
				"91", "92", "fizz", "94", "buzz", "fizz", "97", "98", "fizz", "buzz",
			},
		},
		{
			label:    "should return one element if limit is 1",
			params:   fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 1, Str1: "fizz", Str2: "buzz"},
			expected: []string{"1"},
		},
		{
			label:    "should return first three elements if limit is 3",
			params:   fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 3, Str1: "fizz", Str2: "buzz"},
			expected: []string{"1", "2", "fizz"},
		},
		{
			label:  "should return first 15 elements of default fizzbuzz sequence",
			params: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"},
			expected: []string{
				"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz",
			},
		},
		{
			label:  "should return first 15 elements of alternate fizzbuzz sequence",
			params: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "buzz", Str2: "fizz"},
			expected: []string{
				"1", "2", "buzz", "4", "fizz", "buzz", "7", "8", "buzz", "fizz", "11", "buzz", "13", "14", "buzzfizz",
			},
		},
		{
			label:  "should return first 15 elements of a different fizzbuzz sequence",
			params: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "abc", Str2: "xyz"},
			expected: []string{
				"1", "2", "abc", "4", "xyz", "abc", "7", "8", "abc", "xyz", "11", "abc", "13", "14", "abcxyz",
			},
		},
		{
			label:  "should return first 15 elements of fizzfizz sequence",
			params: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "fizz"},
			expected: []string{
				"1", "2", "fizz", "4", "fizz", "fizz", "7", "8", "fizz", "fizz", "11", "fizz", "13", "14", "fizzfizz",
			},
		},
		{
			label:  "should return first 15 elements of unicode sequence",
			params: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 15, Str1: "é", Str2: "🍕"},
			expected: []string{
				"1", "2", "é", "4", "🍕", "é", "7", "8", "é", "🍕", "11", "é", "13", "14", "é🍕",
			},
		},
		{
			label:  "should return first 15 elements of sequence without fizz or buzz",
			params: fizzbuzz.Params{Int1: 30, Int2: 50, Limit: 15, Str1: "fizz", Str2: "fizz"},
			expected: []string{
				"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15",
			},
		},
		{
			label:  "should return first 10 elements of sequence with number and fizz+buzz",
			params: fizzbuzz.Params{Int1: 2, Int2: 2, Limit: 10, Str1: "fizz", Str2: "buzz"},
			expected: []string{
				"1", "fizzbuzz", "3", "fizzbuzz", "5", "fizzbuzz", "7", "fizzbuzz", "9", "fizzbuzz",
			},
		},
		{
			label:  "should return first 10 elements of sequence with all fizz+buzz",
			params: fizzbuzz.Params{Int1: 1, Int2: 1, Limit: 10, Str1: "fizz", Str2: "buzz"},
			expected: []string{
				"fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz", "fizzbuzz",
			},
		},
		{
			label:     "should return error if param 'Int1' is invalid (0)",
			params:    fizzbuzz.Params{Int1: 0, Int2: 0, Limit: 0, Str1: "", Str2: ""},
			targetErr: fizzbuzz.ErrMustBeBiggerThanZero,
			fieldErr:  `int1`,
		},
		{
			label:     "should return error if param 'Int2' is invalid (0)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 0, Limit: 0, Str1: "", Str2: ""},
			targetErr: fizzbuzz.ErrMustBeBiggerThanZero,
			fieldErr:  `int2`,
		},
		{
			label:     "should return error if param 'Limit' is invalid (0)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 0, Str1: "", Str2: ""},
			targetErr: fizzbuzz.ErrMustBeBiggerThanZero,
			fieldErr:  `limit`,
		},
		{
			label:     "should return error if param 'Str1' is invalid (empty)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 3, Str1: "", Str2: ""},
			targetErr: fizzbuzz.ErrStringMustNotBeEmpty,
			fieldErr:  `str1`,
		},
		{
			label:     "should return error if param 'Str2' is invalid (empty)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 3, Str1: "fizz", Str2: ""},
			targetErr: fizzbuzz.ErrStringMustNotBeEmpty,
			fieldErr:  `str2`,
		},
		{
			label:             "using max limit should not trigger an error",
			params:            fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 1024, Str1: "x", Str2: "y"},
			skipExpectedCheck: true,
		},
		{
			label:             "using max size str1 should not trigger an error",
			params:            fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 15, Str1: "1234567890123456789012345678901234567890123456789012345678901234", Str2: "y"},
			skipExpectedCheck: true,
		},
		{
			label:             "using max size str2 should not trigger an error",
			params:            fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 15, Str1: "x", Str2: "1234567890123456789012345678901234567890123456789012345678901234"},
			skipExpectedCheck: true,
		},
		{
			label:     "should return error if param 'Limit' is invalid (bigger than 1024)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 1025, Str1: "x", Str2: "y"},
			targetErr: fizzbuzz.ErrMustNotExceedMaxValue,
			fieldErr:  `limit`,
		},
		{
			label:     "should return error if param 'Str1' is invalid (bigger than 64 chars)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 3, Str1: "12345678901234567890123456789012345678901234567890123456789012345", Str2: ""},
			targetErr: fizzbuzz.ErrStringMustNotExceedMaxLength,
			fieldErr:  `str1`,
		},
		{
			label:     "should return error if param 'Str2' is invalid (bigger than 64 chars)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 3, Str1: "x", Str2: "12345678901234567890123456789012345678901234567890123456789012345"},
			targetErr: fizzbuzz.ErrStringMustNotExceedMaxLength,
			fieldErr:  `str2`,
		},
		{
			label:     "should return error if param 'Str2' is invalid (bigger than 64 chars using 17 emojis = 68 bytes)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 3, Str1: "x", Str2: "🍕🍕🍕🍕🍕🍕🍕🍕🍕🍕🍕🍕🍕🍕🍕🍕🍕"},
			targetErr: fizzbuzz.ErrStringMustNotExceedMaxLength,
			fieldErr:  `str2`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			result, err := fizzbuzz.Generate(tc.params)
			if tc.targetErr != nil {
				if err == nil {
					t.Fatalf("unexpected nil error (expects: %v)", tc.targetErr)
				}

				if !errors.Is(err, tc.targetErr) {
					t.Fatalf("unexpected error (got: %v, expects: %v", err, tc.targetErr)
				}

				pe, ok := errors.AsType[*fizzbuzz.ParamError](err)
				if !ok {
					t.Fatalf("error is not a ParamError (got: %v)", err)
				}

				if pe.Field != tc.fieldErr {
					t.Fatalf("unexpected field (got: %v, expected; %v)", pe.Field, tc.fieldErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error from generate: %v", err)
			}

			if tc.skipExpectedCheck {
				return
			}

			if !slices.Equal(result, tc.expected) {
				t.Fatalf("unexpected result (got: %v, expect: %v)", result, tc.expected)
			}
		})
	}
}

func FuzzGenerate(f *testing.F) {
	f.Add(3, 5, 15, "fizz", "buzz")
	f.Add(1, 1, 1, "a", "b")
	f.Add(0, 5, 10, "x", "y")

	f.Fuzz(func(t *testing.T, int1, int2, limit int, str1, str2 string) {
		p := fizzbuzz.Params{Int1: int1, Int2: int2, Limit: limit, Str1: str1, Str2: str2}

		out, err := fizzbuzz.Generate(p)
		if err != nil {
			if _, ok := errors.AsType[*fizzbuzz.ParamError](err); !ok {
				t.Fatalf("error is not a *ParamError: %v", err)
			}
			return
		}

		if len(out) != limit {
			t.Fatalf("len = %d, want %d", len(out), limit)
		}
		for i, got := range out {
			n := i + 1
			var want string
			switch {
			case n%int1 == 0 && n%int2 == 0:
				want = str1 + str2
			case n%int1 == 0:
				want = str1
			case n%int2 == 0:
				want = str2
			default:
				want = strconv.Itoa(n)
			}
			if got != want {
				t.Fatalf("out[%d] = %q, want %q", i, got, want)
			}
		}
	})
}

func FuzzParseParams(f *testing.F) {
	f.Add("int1=3&int2=5&limit=15&str1=fizz&str2=buzz")
	f.Add("int1=abc")
	f.Add("")

	f.Fuzz(func(t *testing.T, raw string) {
		values, err := url.ParseQuery(raw)
		if err != nil {
			t.Skip() // not a valid query string; not ParseParams' concern
		}
		_, err = fizzbuzz.ParseParams(values)
		if err != nil && !errors.Is(err, fizzbuzz.ErrNotAnInteger) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
