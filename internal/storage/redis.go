package storage

import (
	"context"
	"sync"
	"time"

	"github.com/ludovicopassari/api-gateway/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var (
	instance *redis.Client
	once     sync.Once
)

type RedisOption func(*redis.Options)

func GetClient(opts ...RedisOption) (*redis.Client, error) {
	once.Do(func() {
		// Configurazione di default
		options := &redis.Options{}

		// Applica le opzioni fornite
		for _, opt := range opts {
			opt(options)
		}

		instance = redis.NewClient(options)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := instance.Ping(ctx).Err(); err != nil {
			logger.Error("redis connection failed")
		}
	})

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
