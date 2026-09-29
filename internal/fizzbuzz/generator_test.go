package fizzbuzz_test

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	biggestAsciiValidString := strings.Repeat("x", 64)
	invalidAsciiString := biggestAsciiValidString + "x"

	biggestUnicodeValidString := strings.Repeat("🍕", 16)
	invalidUnicodeString := biggestUnicodeValidString + "🍕"

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
			params: fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"},
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
			params:            fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 15, Str1: biggestAsciiValidString, Str2: "y"},
			skipExpectedCheck: true,
		},
		{
			label:             "using max size str2 should not trigger an error",
			params:            fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 15, Str1: "x", Str2: biggestAsciiValidString},
			skipExpectedCheck: true,
		},
		{
			label:     "should return error if param 'Limit' is invalid (bigger than 1024)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 1025 + 1, Str1: "x", Str2: "y"},
			targetErr: fizzbuzz.ExceededMaxValueError(1024),
			fieldErr:  `limit`,
		},
		{
			label:     "should return error if param 'Str1' is invalid (bigger than 64 chars)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 3, Str1: invalidAsciiString, Str2: "x"},
			targetErr: fizzbuzz.ExceededMaxLengthError(64),
			fieldErr:  `str1`,
		},
		{
			label:     "should return error if param 'Str2' is invalid (bigger than 64 chars)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 3, Str1: "x", Str2: invalidAsciiString},
			targetErr: fizzbuzz.ExceededMaxLengthError(64),
			fieldErr:  `str2`,
		},
		{
			label:  "using max size str2 unicode should not trigger an error",
			params: fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 15, Str1: "x", Str2: biggestUnicodeValidString},

			skipExpectedCheck: true,
		},
		{
			label:     "should return error if param 'Str2' is invalid (bigger than 64 chars using 17 emojis = 68 bytes)",
			params:    fizzbuzz.Params{Int1: 1, Int2: 2, Limit: 3, Str1: "x", Str2: invalidUnicodeString},
			targetErr: fizzbuzz.ExceededMaxLengthError(64),
			fieldErr:  `str2`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			generator := fizzbuzz.DefaultGenerator()

			result, err := generator.Generate(tc.params)
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
		if limit > 10_000 {
			t.Skip("checking for bigger limits may use a lot of resources")
		}

		params := fizzbuzz.Params{Int1: int1, Int2: int2, Limit: limit, Str1: str1, Str2: str2}

		generator := fizzbuzz.DefaultGenerator()

		out, err := generator.Generate(params)
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
