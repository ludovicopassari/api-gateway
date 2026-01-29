package redistorage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	instance *redis.Client
	once     sync.Once
	initErr  error
)

type RedisOption func(*redis.Options)

func Init(opts ...RedisOption) error {
	once.Do(func() {

		options := &redis.Options{}
		for _, opt := range opts {
			opt(options)
		}
		instance = redis.NewClient(options)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := instance.Ping(ctx).Err(); err != nil {
			if closeErr := instance.Close(); closeErr != nil {
				initErr = fmt.Errorf("redis ping failed: %w, close failed: %v", err, closeErr)
			} else {
				initErr = fmt.Errorf("redis ping failed: %w", err)
			}
			instance = nil
			return
		}
	})

	return initErr
}

func RedisClient() (*redis.Client, error) {
	if instance == nil {
		return nil, fmt.Errorf("redis client not initialized, call Init first")
	}
	return instance, nil
}

func Close() error {
	if instance != nil {
		return instance.Close()
	}
	return nil
}

func WithAddr(addr string) RedisOption {
	return func(opts *redis.Options) {
		opts.Addr = addr
	}
}

func WithPassword(password string) RedisOption {
	return func(opts *redis.Options) {
		opts.Password = password
	}
}
