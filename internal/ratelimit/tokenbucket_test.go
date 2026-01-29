package ratelimit

import (
	"testing"
	"time"

	"github.com/ludovicopassari/api-gateway/internal/storage"
)

func TestTokenBucket_Allow(t *testing.T) {
	tests := []struct {
		name           string
		capacity       int
		refillRate     time.Duration
		requests       int
		expectedAllows int
		waitBetween    time.Duration
	}{
		{
			name:           "all requests within capacity",
			capacity:       5,
			refillRate:     time.Second,
			requests:       3,
			expectedAllows: 3,
			waitBetween:    0,
		},
		{
			name:           "exceed capacity",
			capacity:       3,
			refillRate:     time.Second,
			requests:       5,
			expectedAllows: 3,
			waitBetween:    0,
		},
		{
			name:           "refill allows more requests",
			capacity:       2,
			refillRate:     100 * time.Millisecond,
			requests:       3,
			expectedAllows: 3,
			waitBetween:    150 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := storage.NewMemoryStorage()
			tb := NewTokenBucket(storage, tt.capacity, tt.refillRate)
			key := "test-key"

			allowed := 0
			for i := 0; i < tt.requests; i++ {
				if i > 0 && tt.waitBetween > 0 {
					time.Sleep(tt.waitBetween)
				}
				if tb.Allow(key) {
					allowed++
				}
			}

			if allowed != tt.expectedAllows {
				t.Errorf("expected %d allowed requests, got %d", tt.expectedAllows, allowed)
			}
		})
	}
}

/* func TestTokenBucket_Concurrent(t *testing.T) {
	storage := storage.NewMemoryStorage()
	tb := NewTokenBucket(storage, 10, time.Second)
	key := "concurrent-test"

	done := make(chan bool)
	allowed := make(chan bool, 20)

	// Simulate 20 concurrent requests
	for i := 0; i < 20; i++ {
		go func() {
			allowed <- tb.Allow(key)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
	close(allowed)

	count := 0
	for a := range allowed {
		if a {
			count++
		}
	}

	// Should allow exactly 10 (capacity)
	if count != 10 {
		t.Errorf("expected 10 allowed requests, got %d", count)
	}
} */

/* func TestTokenBucket_Refill(t *testing.T) {
	storage := storage.NewMemoryStorage()
	refillRate := 50 * time.Millisecond
	tb := NewTokenBucket(storage, 2, refillRate)
	key := "refill-test"

	// Consume all tokens
	tb.Allow(key)
	tb.Allow(key)

	// Should be denied
	if tb.Allow(key) {
		t.Error("expected request to be denied after consuming all tokens")
	}

	// Wait for refill
	time.Sleep(refillRate + 10*time.Millisecond)

	// Should be allowed after refill
	if !tb.Allow(key) {
		t.Error("expected request to be allowed after refill")
	}
}
*/
