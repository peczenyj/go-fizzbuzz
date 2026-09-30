package stats_test

import (
	"sync"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/stats"
)

func TestCounter_empty(t *testing.T) {
	t.Parallel()

	c := stats.NewCounter[string]()

	assertStatsTop(t, c, "", 0, false)
}

func TestCounter_single_hit(t *testing.T) {
	t.Parallel()

	c := stats.NewCounter[string]()

	c.Record("x")

	assertStatsTop(t, c, "x", 1, true)
}

func TestCounter_multiple_hits(t *testing.T) {
	t.Parallel()

	c := stats.NewCounter[string]()

	c.Record("x")
	c.Record("y")
	c.Record("y")
	c.Record("x")
	c.Record("x")

	assertStatsTop(t, c, "x", 3, true)

	c.Record("y")
	c.Record("y")

	assertStatsTop(t, c, "y", 4, true)
}

func assertStatsTop[K comparable](t *testing.T,
	c stats.Interface[K],
	expectedKey K,
	expectedHits int,
	expectedOk bool,
) {
	t.Helper()

	key, hits, ok := c.Top()

	if ok != expectedOk {
		t.Fatalf("unexpected ok (got: %v, expected: %v)", ok, expectedOk)
	}

	if hits != expectedHits {
		t.Fatalf("unexpected hits (got: %v, expected: %v)", hits, expectedHits)
	}

	if key != expectedKey {
		t.Fatalf("unexpected key (got: %v, expected: %v)", key, expectedKey)
	}
}

func TestCounter_concurrent(t *testing.T) {
	t.Parallel()

	const goroutines, perGoroutine = 50, 1000

	c := stats.NewCounter[string]()

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perGoroutine {
				c.Record("x")
			}
		})
	}
	wg.Wait()

	key, hits, ok := c.Top()
	if !ok || key != "x" || hits != goroutines*perGoroutine {
		t.Fatalf("got (%q, %d, %v), want (%q, %d, true)", key, hits, ok, "x", goroutines*perGoroutine)
	}
}
