package limiter

import (
	"sync"
	"time"
)

type client struct {
	count     int
	expiresAt time.Time
}
type RateLimiter interface {
	Allow(key string) (bool, time.Duration)
}

type MemoryLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]*client
}

func NewMemoryLimiter(limit int, window time.Duration) *MemoryLimiter {
	return &MemoryLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*client),
	}
}

func (l *MemoryLimiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	c, exists := l.clients[key]

	if !exists || now.After(c.expiresAt) {
		l.clients[key] = &client{
			count:     1,
			expiresAt: now.Add(l.window),
		}
		return true, l.window
	}

	if c.count >= l.limit {
		return false, time.Until(c.expiresAt)
	}

	c.count++
	return true, time.Until(c.expiresAt)
}
