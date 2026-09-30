package fizzbuzz_test

import (
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
