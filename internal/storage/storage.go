// pkg/storage/storage.go
package storage

import (
	"context"
	"time"
)

// Storage definisce l'interfaccia unificata per tutti i backend
type Storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)

	SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error

	ListPush(ctx context.Context, key string, values ...string) error
	ListRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ListTrim(ctx context.Context, key string, start, stop int64) error
	ListLen(ctx context.Context, key string) (int64, error)

	ZAdd(ctx context.Context, key string, score float64, member string) error
	ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error)
	ZCount(ctx context.Context, key string, min, max float64) (int64, error)

	Ping(ctx context.Context) error
	Close() error
	FlushAll(ctx context.Context) error
}

type StorageError struct {
	Code    string
	Message string
	Err     error
}

var (
	ErrKeyNotFound      = &StorageError{Code: "KEY_NOT_FOUND", Message: "key not found"}
	ErrConnectionFailed = &StorageError{Code: "CONNECTION_FAILED", Message: "storage connection failed"}
	ErrInvalidValue     = &StorageError{Code: "INVALID_VALUE", Message: "invalid value"}
)

func (e *StorageError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *StorageError) Unwrap() error {
	return e.Err
}
