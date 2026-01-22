package limiters

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	*redis.Client
	limit  int64
	window time.Duration
}

func NewSlidingWindowLimiter(redis_client *redis.Client, limit int64, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		Client: redis_client,
		limit:  limit,
		window: window,
	}
}

// this implementation of sliding window limiter use Redis sorted-sets
func (sw SlidingWindowLimiter) Allow(clientID string) (bool, error) {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	key := fmt.Sprintf("rate:%s", clientID)
	member := fmt.Sprintf("%d-%s", now, clientID)

	luaScript := `
        redis.call("ZADD", KEYS[1], ARGV[1], ARGV[2])
        redis.call("ZREMRANGEBYSCORE", KEYS[1], 0, ARGV[3])
        return redis.call("ZCARD", KEYS[1])
    `

	windowStart := now - sw.window.Milliseconds()
	count, err := sw.Eval(ctx, luaScript, []string{key}, now, member, windowStart).Int64()

	if err != nil {
		return false, err
	}

	return count <= sw.limit, nil
}
