package fizzbuzz_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

func BenchmarkGenerate(b *testing.B) {
	gen := fizzbuzz.DefaultGenerator()

	long := strings.Repeat("x", 64) // max label length by default

	for _, bc := range []struct {
		name   string
		params fizzbuzz.Params
	}{
		{"classic_100", fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}},
		{"max_limit_1024", fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 1024, Str1: "fizz", Str2: "buzz"}},
		{"max_labels_1024", fizzbuzz.Params{Int1: 1, Int2: 1, Limit: 1024, Str1: long, Str2: long}},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				if _, err := gen.Generate(bc.params); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkNewGenerator measures the one-off cost of precomputing the numbers.
func BenchmarkNewGenerator(b *testing.B) {
	for _, maxLimit := range []int{fizzbuzz.DefaultMaxLimit, 100_000} {
		b.Run(strconv.Itoa(maxLimit), func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				if _, err := fizzbuzz.NewGenerator(maxLimit, fizzbuzz.DefaultMaxStringLength); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
