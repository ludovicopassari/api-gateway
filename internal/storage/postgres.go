// pkg/storage/postgres.go
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, &StorageError{
			Code:    "CONNECTION_FAILED",
			Message: "failed to connect to PostgreSQL",
			Err:     err,
		}
	}

	if err := db.Ping(); err != nil {
		return nil, &StorageError{
			Code:    "CONNECTION_FAILED",
			Message: "failed to ping PostgreSQL",
			Err:     err,
		}
	}

	storage := &PostgresStorage{db: db}
	if err := storage.initSchema(); err != nil {
		return nil, err
	}

	// Background cleanup
	go storage.cleanupExpired()

	return storage, nil
}

func (p *PostgresStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS kv_store (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		expires_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS list_store (
		key TEXT PRIMARY KEY,
		values JSONB NOT NULL,
		expires_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS zset_store (
		key TEXT NOT NULL,
		member TEXT NOT NULL,
		score DOUBLE PRECISION NOT NULL,
		PRIMARY KEY (key, member)
	);

	CREATE INDEX IF NOT EXISTS idx_kv_expires ON kv_store(expires_at);
	CREATE INDEX IF NOT EXISTS idx_list_expires ON list_store(expires_at);
	CREATE INDEX IF NOT EXISTS idx_zset_score ON zset_store(key, score);
	`

	_, err := p.db.Exec(schema)
	if err != nil {
		return &StorageError{
			Code:    "SCHEMA_INIT_FAILED",
			Message: "failed to initialize schema",
			Err:     err,
		}
	}
	return nil
}

func (p *PostgresStorage) cleanupExpired() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := p.db.ExecContext(ctx,
			"DELETE FROM kv_store WHERE expires_at IS NOT NULL AND expires_at < NOW()")
		if err != nil {
			// Log error but continue
		}

		_, err = p.db.ExecContext(ctx,
			"DELETE FROM list_store WHERE expires_at IS NOT NULL AND expires_at < NOW()")
		if err != nil {
			// Log error but continue
		}

		cancel()
	}
}

func (p *PostgresStorage) Get(ctx context.Context, key string) (string, error) {
	var value string
	var expiresAt sql.NullTime

	err := p.db.QueryRowContext(ctx,
		"SELECT value, expires_at FROM kv_store WHERE key = $1", key).
		Scan(&value, &expiresAt)

	if err == sql.ErrNoRows {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", &StorageError{Code: "GET_FAILED", Message: "failed to get key", Err: err}
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		p.Delete(ctx, key)
		return "", ErrKeyNotFound
	}

	return value, nil
}

func (p *PostgresStorage) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	var expiresAt sql.NullTime
	if ttl > 0 {
		expiresAt = sql.NullTime{Time: time.Now().Add(ttl), Valid: true}
	}

	_, err := p.db.ExecContext(ctx,
		`INSERT INTO kv_store (key, value, expires_at) 
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO UPDATE 
		 SET value = EXCLUDED.value, expires_at = EXCLUDED.expires_at`,
		key, value, expiresAt)

	if err != nil {
		return &StorageError{Code: "SET_FAILED", Message: "failed to set key", Err: err}
	}
	return nil
}

func (p *PostgresStorage) Delete(ctx context.Context, key string) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM kv_store WHERE key = $1", key)
	if err != nil {
		return &StorageError{Code: "DELETE_FAILED", Message: "failed to delete key", Err: err}
	}

	_, err = p.db.ExecContext(ctx, "DELETE FROM list_store WHERE key = $1", key)
	_, err = p.db.ExecContext(ctx, "DELETE FROM zset_store WHERE key = $1", key)

	return nil
}

func (p *PostgresStorage) Exists(ctx context.Context, key string) (bool, error) {
	var count int
	err := p.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM kv_store 
		 WHERE key = $1 AND (expires_at IS NULL OR expires_at > NOW())`,
		key).Scan(&count)

	if err != nil {
		return false, &StorageError{Code: "EXISTS_FAILED", Message: "failed to check existence", Err: err}
	}
	return count > 0, nil
}

func (p *PostgresStorage) Increment(ctx context.Context, key string) (int64, error) {
	return p.IncrementBy(ctx, key, 1)
}

func (p *PostgresStorage) Decrement(ctx context.Context, key string) (int64, error) {
	return p.IncrementBy(ctx, key, -1)
}

func (p *PostgresStorage) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, &StorageError{Code: "TX_BEGIN_FAILED", Err: err}
	}
	defer tx.Rollback()

	var currentValue int64
	err = tx.QueryRowContext(ctx,
		"SELECT CAST(value AS BIGINT) FROM kv_store WHERE key = $1 FOR UPDATE",
		key).Scan(&currentValue)

	if err == sql.ErrNoRows {
		currentValue = 0
	} else if err != nil {
		return 0, &StorageError{Code: "INCREMENT_FAILED", Err: err}
	}

	newValue := currentValue + value

	_, err = tx.ExecContext(ctx,
		`INSERT INTO kv_store (key, value, expires_at) 
		 VALUES ($1, $2, NULL)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		key, fmt.Sprintf("%d", newValue))

	if err != nil {
		return 0, &StorageError{Code: "INCREMENT_FAILED", Err: err}
	}

	if err := tx.Commit(); err != nil {
		return 0, &StorageError{Code: "TX_COMMIT_FAILED", Err: err}
	}

	return newValue, nil
}

func (p *PostgresStorage) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return p.Set(ctx, key, value, ttl)
}

func (p *PostgresStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	var expiresAt sql.NullTime
	err := p.db.QueryRowContext(ctx,
		"SELECT expires_at FROM kv_store WHERE key = $1", key).
		Scan(&expiresAt)

	if err == sql.ErrNoRows {
		return -2 * time.Second, nil
	}
	if err != nil {
		return 0, &StorageError{Code: "GET_TTL_FAILED", Err: err}
	}

	if !expiresAt.Valid {
		return -1 * time.Second, nil
	}

	return time.Until(expiresAt.Time), nil
}

func (p *PostgresStorage) Expire(ctx context.Context, key string, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)
	_, err := p.db.ExecContext(ctx,
		"UPDATE kv_store SET expires_at = $1 WHERE key = $2",
		expiresAt, key)

	if err != nil {
		return &StorageError{Code: "EXPIRE_FAILED", Err: err}
	}
	return nil
}

// List operations
func (p *PostgresStorage) ListPush(ctx context.Context, key string, values ...string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return &StorageError{Code: "TX_BEGIN_FAILED", Err: err}
	}
	defer tx.Rollback()

	var currentValues []string
	var valuesJSON []byte

	err = tx.QueryRowContext(ctx,
		"SELECT values FROM list_store WHERE key = $1 FOR UPDATE", key).
		Scan(&valuesJSON)

	if err == sql.ErrNoRows {
		currentValues = []string{}
	} else if err != nil {
		return &StorageError{Code: "LIST_PUSH_FAILED", Err: err}
	} else {
		json.Unmarshal(valuesJSON, &currentValues)
	}

	currentValues = append(currentValues, values...)
	newValuesJSON, _ := json.Marshal(currentValues)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO list_store (key, values, expires_at) 
		 VALUES ($1, $2, NULL)
		 ON CONFLICT (key) DO UPDATE SET values = EXCLUDED.values`,
		key, newValuesJSON)

	if err != nil {
		return &StorageError{Code: "LIST_PUSH_FAILED", Err: err}
	}

	return tx.Commit()
}

func (p *PostgresStorage) ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	var valuesJSON []byte
	err := p.db.QueryRowContext(ctx,
		"SELECT values FROM list_store WHERE key = $1", key).
		Scan(&valuesJSON)

	if err == sql.ErrNoRows {
		return []string{}, nil
	}
	if err != nil {
		return nil, &StorageError{Code: "LIST_RANGE_FAILED", Err: err}
	}

	var values []string
	json.Unmarshal(valuesJSON, &values)

	length := int64(len(values))
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []string{}, nil
	}

	return values[start : stop+1], nil
}

func (p *PostgresStorage) ListTrim(ctx context.Context, key string, start, stop int64) error {
	values, err := p.ListRange(ctx, key, start, stop)
	if err != nil {
		return err
	}

	valuesJSON, _ := json.Marshal(values)
	_, err = p.db.ExecContext(ctx,
		"UPDATE list_store SET values = $1 WHERE key = $2",
		valuesJSON, key)

	return err
}

func (p *PostgresStorage) ListLen(ctx context.Context, key string) (int64, error) {
	var valuesJSON []byte
	err := p.db.QueryRowContext(ctx,
		"SELECT values FROM list_store WHERE key = $1", key).
		Scan(&valuesJSON)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, &StorageError{Code: "LIST_LEN_FAILED", Err: err}
	}

	var values []string
	json.Unmarshal(valuesJSON, &values)
	return int64(len(values)), nil
}

// Sorted Set operations
func (p *PostgresStorage) ZAdd(ctx context.Context, key string, score float64, member string) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO zset_store (key, member, score) 
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key, member) DO UPDATE SET score = EXCLUDED.score`,
		key, member, score)

	if err != nil {
		return &StorageError{Code: "ZADD_FAILED", Err: err}
	}
	return nil
}

func (p *PostgresStorage) ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error) {
	result, err := p.db.ExecContext(ctx,
		"DELETE FROM zset_store WHERE key = $1 AND score >= $2 AND score <= $3",
		key, min, max)

	if err != nil {
		return 0, &StorageError{Code: "ZREM_RANGE_FAILED", Err: err}
	}

	count, _ := result.RowsAffected()
	return count, nil
}

func (p *PostgresStorage) ZCount(ctx context.Context, key string, min, max float64) (int64, error) {
	var count int64
	err := p.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM zset_store WHERE key = $1 AND score >= $2 AND score <= $3",
		key, min, max).Scan(&count)

	if err != nil {
		return 0, &StorageError{Code: "ZCOUNT_FAILED", Err: err}
	}
	return count, nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.Ping()
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

func (p *PostgresStorage) FlushAll(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM kv_store")
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM list_store")
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM zset_store")
	return err
}
