package stats_test

import (
	"sync"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/stats"
)

func TestCounter(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label        string
		records      []string
		expectedKey  string
		expectedHits int
		expectedOk   bool
	}{
		{
			label: "should report nothing without records",
		},
		{
			label:        "should report a single hit",
			records:      []string{"x"},
			expectedKey:  "x",
			expectedHits: 1,
			expectedOk:   true,
		},
		{
			label:        "should report the most frequent key",
			records:      []string{"x", "y", "y", "x", "x"},
			expectedKey:  "x",
			expectedHits: 3,
			expectedOk:   true,
		},
		{
			label:        "should keep the first key to reach a tied count",
			records:      []string{"x", "y", "y", "x", "x", "y"},
			expectedKey:  "x",
			expectedHits: 3,
			expectedOk:   true,
		},
		{
			label:        "should change the top when another key overtakes it",
			records:      []string{"x", "y", "y", "x", "x", "y", "y"},
			expectedKey:  "y",
			expectedHits: 4,
			expectedOk:   true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			c := stats.NewCounter[string]()

			for _, k := range tc.records {
				c.Record(k)
			}

			assertStatsTop(t, c, tc.expectedKey, tc.expectedHits, tc.expectedOk)
		})
	}
}

func assertStatsTop[K comparable](t *testing.T,
	c *stats.Counter[K],
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

	// Each iteration records a once, b twice and c three times, so the top
	// changes while goroutines race (a, then b, then c), and c must end on top
	// with exactly three times the iterations.
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perGoroutine {
				c.Record("a")
				c.Record("b")
				c.Record("b")
				c.Record("c")
				c.Record("c")
				c.Record("c")
			}
		})
	}
	wg.Wait()

	assertStatsTop(t, c, "c", 3*goroutines*perGoroutine, true)
}
