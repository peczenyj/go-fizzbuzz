package api

import "github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"

var _ Statistics = (*nop)(nil)

// NopStatistics returns a statistics object that does nothing.
func NopStatistics() Statistics { return &nop{} }

type nop struct{}

func (*nop) Record(_ fizzbuzz.Params) {}

func (*nop) Top() (params fizzbuzz.Params, hits int, ok bool) {
	return params, hits, false
}
