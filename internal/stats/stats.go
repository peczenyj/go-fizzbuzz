package stats

import "sync"

var _ Interface[string] = (*Counter[string])(nil)

// Interface abstraction.
type Interface[K comparable] interface {
	Record(k K)
	Top() (key K, hits int, ok bool)
}

// Counter counts occurrences of keys and tracks the most frequent one in O(1).
// It is safe for concurrent use.
type Counter[K comparable] struct {
	mu      sync.Mutex
	counts  map[K]int
	top     K
	topHits int
}

// NewCounter ctor.
func NewCounter[K comparable]() *Counter[K] {
	return &Counter[K]{counts: map[K]int{}}
}

// Record counts one occurrence of k. When two keys have the same count,
// the first one to reach it stays on top.
func (c *Counter[K]) Record(k K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts[k]++
	if n := c.counts[k]; n > c.topHits {
		c.top, c.topHits = k, n
	}
}

// Top returns the most frequent key and its count; ok is false if nothing was recorded.
func (c *Counter[K]) Top() (key K, hits int, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.top, c.topHits, c.topHits > 0
}
