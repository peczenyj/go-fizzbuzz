package fizzbuzz_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label     string
		params    fizzbuzz.Params
		targetErr error
		expected  []string
	}{
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
			label:     "should return error if param is invalid",
			params:    fizzbuzz.Params{},
			targetErr: fizzbuzz.ErrMustBeBiggerThanZero,
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

				return
			}

			if err != nil {
				t.Fatalf("unexpected error from generate: %v", err)
			}

			if !slices.Equal(result, tc.expected) {
				t.Fatalf("unexpected result (got: %v, expect: %v)", result, tc.expected)
			}
		})
	}
}
