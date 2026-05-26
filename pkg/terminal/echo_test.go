package terminal

import (
	"testing"
)

func TestEchoCache(t *testing.T) {
	cache := NewEchoCache(5)

	payload1 := []byte("hello world")
	payload2 := []byte("another payload")
	payload3 := []byte("hello world") // Same as 1

	if cache.Contains(payload1) {
		t.Error("Cache should be empty")
	}

	cache.Add(payload1)

	if !cache.Contains(payload1) {
		t.Error("Cache should contain payload1")
	}

	if cache.Contains(payload2) {
		t.Error("Cache should not contain payload2")
	}

	if !cache.Contains(payload3) {
		t.Error("Cache should contain payload3 (identical to payload1)")
	}

	// Fill the cache to trigger wrap-around
	for i := 0; i < 5; i++ {
		cache.Add([]byte{byte(i)})
	}

	// payload1 should have been evicted
	if cache.Contains(payload1) {
		t.Error("Cache should have evicted payload1")
	}

	// newest should be there
	if !cache.Contains([]byte{4}) {
		t.Error("Cache should contain recently added item")
	}
}
