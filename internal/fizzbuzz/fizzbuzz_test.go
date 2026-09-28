package fizzbuzz_test

import (
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label    string
		params   fizzbuzz.Params
		errMsg   string
		expected []string
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
			label:  "should return error if param is invalid",
			params: fizzbuzz.Params{},
			errMsg: "unable to generate fizzbuzz sequence: invalid field 'Int1': must be bigger than zero",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			result, err := fizzbuzz.Generate(tc.params)
			if tc.errMsg != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.errMsg)
				}

				if tc.errMsg != err.Error() {
					t.Fatalf("expected error %q, got %q", tc.errMsg, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error from generate: %v", err)
			}

			if len(result) != len(tc.expected) {
				t.Fatalf("insufficient result (got %d elements, expects %d)", len(result), len(tc.expected))
			}

			for index, element := range result {
				if element != tc.expected[index] {
					t.Fatalf("element mismatch at position %d (got %v, expects %v)", index, element, tc.expected[index])
				}
			}
		})
	}
}
