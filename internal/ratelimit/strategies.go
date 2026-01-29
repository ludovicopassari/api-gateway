// internal/ratelimit/strategies.go
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/ludovicopassari/api-gateway/internal/storage"
)

type RateLimiter interface {
	Allow(key string) bool
}

type TokenBucket struct {
	storage    storage.Storage
	capacity   int
	refillRate time.Duration
}

func NewTokenBucket(storage storage.Storage, capacity int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		storage:    storage,
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (tb *TokenBucket) Allow(key string) bool {
	ctx := context.Background()

	tokensKey := fmt.Sprintf("tb:tokens:%s", key)
	lastRefillKey := fmt.Sprintf("tb:refill:%s", key)

	// Get current tokens
	tokensStr, err := tb.storage.Get(ctx, tokensKey)
	tokens := tb.capacity
	if err == nil {
		fmt.Sscanf(tokensStr, "%d", &tokens)
	}

	// Calculate refill
	lastRefillStr, err := tb.storage.Get(ctx, lastRefillKey)
	if err == nil {
		var lastRefillUnix int64
		fmt.Sscanf(lastRefillStr, "%d", &lastRefillUnix)
		lastRefill := time.Unix(lastRefillUnix, 0)

		elapsed := time.Since(lastRefill)
		refillTokens := int(elapsed / tb.refillRate)

		if refillTokens > 0 {
			tokens = min(tokens+refillTokens, tb.capacity)
			tb.storage.Set(ctx, lastRefillKey, fmt.Sprintf("%d", time.Now().Unix()), 0)
		}
	} else {
		// First request
		tb.storage.Set(ctx, lastRefillKey, fmt.Sprintf("%d", time.Now().Unix()), 0)
	}

	// Check and consume token
	if tokens > 0 {
		tokens--
		tb.storage.Set(ctx, tokensKey, fmt.Sprintf("%d", tokens), 0)
		return true
	}

	return false
}

type SlidingWindow struct {
	storage storage.Storage
	limit   int
	window  time.Duration
}

func NewSlidingWindow(storage storage.Storage, limit int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		storage: storage,
		limit:   limit,
		window:  window,
	}
}

func (sw *SlidingWindow) Allow(key string) bool {
	ctx := context.Background()
	zsetKey := fmt.Sprintf("sw:%s", key)

	now := float64(time.Now().UnixNano())
	windowStart := now - float64(sw.window.Nanoseconds())

	// Remove old entries
	sw.storage.ZRemRangeByScore(ctx, zsetKey, 0, windowStart)

	// Count current requests
	count, err := sw.storage.ZCount(ctx, zsetKey, windowStart, now)
	if err != nil {
		return false
	}

	if count < int64(sw.limit) {
		// Add current request
		memberID := fmt.Sprintf("%d", time.Now().UnixNano())
		sw.storage.ZAdd(ctx, zsetKey, now, memberID)
		sw.storage.Expire(ctx, zsetKey, sw.window*2) // Cleanup
		return true
	}

	return false
}

type FixedWindow struct {
	storage storage.Storage
	limit   int
	window  time.Duration
}

func NewFixedWindow(storage storage.Storage, limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		storage: storage,
		limit:   limit,
		window:  window,
	}
}

func (fw *FixedWindow) Allow(key string) bool {
	ctx := context.Background()

	now := time.Now()
	windowStart := now.Truncate(fw.window)
	counterKey := fmt.Sprintf("fw:%s:%d", key, windowStart.Unix())

	count, err := fw.storage.Increment(ctx, counterKey)
	if err != nil {
		return false
	}

	if count == 1 {
		fw.storage.Expire(ctx, counterKey, fw.window*2)
	}

	return count <= int64(fw.limit)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
