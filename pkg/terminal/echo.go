package terminal

import (
	"hash/fnv"
	"sync"
)

// EchoCache is used to prevent the Spoke from relaying packets it just broadcasted
// back to the Hub (network loop).
type EchoCache struct {
	mu     sync.Mutex
	hashes []uint64
	size   int
	head   int
}

// NewEchoCache creates a new EchoCache of the specified size.
// A size of 256 or 512 is usually sufficient for short-lived echo loops.
func NewEchoCache(size int) *EchoCache {
	return &EchoCache{
		hashes: make([]uint64, size),
		size:   size,
		head:   0,
	}
}

// Add computes the hash of a packet and stores it in the ring buffer.
func (c *EchoCache) Add(data []byte) {
	h := fnv.New64a()
	h.Write(data)
	sum := h.Sum64()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.hashes[c.head] = sum
	c.head = (c.head + 1) % c.size
}

// Contains checks if the packet's hash is currently in the cache.
func (c *EchoCache) Contains(data []byte) bool {
	h := fnv.New64a()
	h.Write(data)
	sum := h.Sum64()

	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < c.size; i++ {
		if c.hashes[i] == sum {
			return true
		}
	}
	return false
}
