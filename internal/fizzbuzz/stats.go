package fizzbuzz

import (
	"github.com/peczenyj/go-fizzbuzz/internal/stats"
)

// DefaultStatistics returns a counter that accepts fizzbuzz params.
func DefaultStatistics() *stats.Counter[Params] {
	return stats.NewCounter[Params]()
}
