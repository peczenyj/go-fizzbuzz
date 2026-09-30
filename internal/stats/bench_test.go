package stats_test

import (
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/stats"
)

// BenchmarkCounter_Record measures Record under contention: every goroutine
// takes the same mutex. Run with -cpu=1,2,4,8 to see how it scales.
func BenchmarkCounter_Record(b *testing.B) {
	c := stats.NewCounter[string]()

	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Record("x")
		}
	})
}
