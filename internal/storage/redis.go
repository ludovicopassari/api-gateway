// pkg/storage/redis.go
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, &StorageError{
			Code:    "CONNECTION_FAILED",
			Message: "failed to connect to Redis",
			Err:     err,
		}
	}

	return &RedisStorage{client: client}, nil
}

func (r *RedisStorage) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", &StorageError{Code: "GET_FAILED", Message: "failed to get key", Err: err}
	}
	return val, nil
}

func (r *RedisStorage) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return &StorageError{Code: "SET_FAILED", Message: "failed to set key", Err: err}
	}
	return nil
}

func (r *RedisStorage) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return &StorageError{Code: "DELETE_FAILED", Message: "failed to delete key", Err: err}
	}
	return nil
}

func (r *RedisStorage) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, &StorageError{Code: "EXISTS_FAILED", Message: "failed to check key existence", Err: err}
	}
	return count > 0, nil
}

func (r *RedisStorage) Increment(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, &StorageError{Code: "INCREMENT_FAILED", Message: "failed to increment key", Err: err}
	}
	return val, nil
}

func (r *RedisStorage) Decrement(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, &StorageError{Code: "DECREMENT_FAILED", Message: "failed to decrement key", Err: err}
	}
	return val, nil
}

func (r *RedisStorage) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	val, err := r.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, &StorageError{Code: "INCREMENT_BY_FAILED", Message: "failed to increment by value", Err: err}
	}
	return val, nil
}

func (r *RedisStorage) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.Set(ctx, key, value, ttl)
}

func (r *RedisStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, &StorageError{Code: "GET_TTL_FAILED", Message: "failed to get TTL", Err: err}
	}
	return ttl, nil
}

func (r *RedisStorage) Expire(ctx context.Context, key string, ttl time.Duration) error {
	err := r.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return &StorageError{Code: "EXPIRE_FAILED", Message: "failed to set expiration", Err: err}
	}
	return nil
}

// List operations
func (r *RedisStorage) ListPush(ctx context.Context, key string, values ...string) error {
	err := r.client.RPush(ctx, key, values).Err()
	if err != nil {
		return &StorageError{Code: "LIST_PUSH_FAILED", Message: "failed to push to list", Err: err}
	}
	return nil
}

func (r *RedisStorage) ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	vals, err := r.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, &StorageError{Code: "LIST_RANGE_FAILED", Message: "failed to get list range", Err: err}
	}
	return vals, nil
}

func (r *RedisStorage) ListTrim(ctx context.Context, key string, start, stop int64) error {
	err := r.client.LTrim(ctx, key, start, stop).Err()
	if err != nil {
		return &StorageError{Code: "LIST_TRIM_FAILED", Message: "failed to trim list", Err: err}
	}
	return nil
}

func (r *RedisStorage) ListLen(ctx context.Context, key string) (int64, error) {
	length, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, &StorageError{Code: "LIST_LEN_FAILED", Message: "failed to get list length", Err: err}
	}
	return length, nil
}

// Sorted Set operations
func (r *RedisStorage) ZAdd(ctx context.Context, key string, score float64, member string) error {
	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
	if err != nil {
		return &StorageError{Code: "ZADD_FAILED", Message: "failed to add to sorted set", Err: err}
	}
	return nil
}

func (r *RedisStorage) ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error) {
	count, err := r.client.ZRemRangeByScore(ctx, key,
		fmt.Sprintf("%f", min),
		fmt.Sprintf("%f", max)).Result()
	if err != nil {
		return 0, &StorageError{Code: "ZREM_RANGE_FAILED", Message: "failed to remove from sorted set", Err: err}
	}
	return count, nil
}

func (r *RedisStorage) ZCount(ctx context.Context, key string, min, max float64) (int64, error) {
	count, err := r.client.ZCount(ctx, key,
		fmt.Sprintf("%f", min),
		fmt.Sprintf("%f", max)).Result()
	if err != nil {
		return 0, &StorageError{Code: "ZCOUNT_FAILED", Message: "failed to count sorted set", Err: err}
	}
	return count, nil
}

func (r *RedisStorage) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return &StorageError{Code: "PING_FAILED", Message: "ping failed", Err: err}
	}
	return nil
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}

func (r *RedisStorage) FlushAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}
