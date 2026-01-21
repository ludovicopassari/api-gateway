package limiters

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type FixedWindowLimiter struct {
	*redis.Client
	limit  int64
	window time.Duration
}

func NewFixedWindow(redis_client *redis.Client, limit int64, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		Client: redis_client,
		limit:  limit,
		window: window,
	}
}
func (wl FixedWindowLimiter) Allow(clientID string) (bool, error) {

	var luaScript = redis.NewScript(`
		local current = redis.call("GET", KEYS[1])
		if current == false then
			redis.call("SET", KEYS[1], 1, "EX", ARGV[1])
			return 1
		else
			return redis.call("INCR", KEYS[1])
		end
	`)

	key := fmt.Sprintf("ip:%s", clientID)
	ctx := context.Background()

	val, err := luaScript.Run(ctx, wl, []string{key}, 10).Int64()
	if err != nil {
		return false, err
	}

	return val <= wl.limit, nil
}
