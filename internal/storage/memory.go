// pkg/storage/memory.go
package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type item struct {
	value     string
	expiresAt time.Time
}

type listItem struct {
	values    []string
	expiresAt time.Time
}

type zsetMember struct {
	score  float64
	member string
}

type MemoryStorage struct {
	mu    sync.RWMutex
	data  map[string]*item
	lists map[string]*listItem
	zsets map[string][]zsetMember
}

func NewMemoryStorage() *MemoryStorage {
	s := &MemoryStorage{
		data:  make(map[string]*item),
		lists: make(map[string]*listItem),
		zsets: make(map[string][]zsetMember),
	}

	// Background cleanup
	go s.cleanupExpired()

	return s
}

func (m *MemoryStorage) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()

		// Cleanup data
		for key, item := range m.data {
			if !item.expiresAt.IsZero() && item.expiresAt.Before(now) {
				delete(m.data, key)
			}
		}

		// Cleanup lists
		for key, list := range m.lists {
			if !list.expiresAt.IsZero() && list.expiresAt.Before(now) {
				delete(m.lists, key)
			}
		}

		m.mu.Unlock()
	}
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	if !item.expiresAt.IsZero() && item.expiresAt.Before(time.Now()) {
		return "", ErrKeyNotFound
	}

	return item.value, nil
}

func (m *MemoryStorage) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	m.data[key] = &item{
		value:     value,
		expiresAt: expiresAt,
	}
	return nil
}

func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	delete(m.lists, key)
	delete(m.zsets, key)
	return nil
}

func (m *MemoryStorage) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return false, nil
	}

	if !item.expiresAt.IsZero() && item.expiresAt.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

func (m *MemoryStorage) Increment(ctx context.Context, key string) (int64, error) {
	return m.IncrementBy(ctx, key, 1)
}

func (m *MemoryStorage) Decrement(ctx context.Context, key string) (int64, error) {
	return m.IncrementBy(ctx, key, -1)
}

func (m *MemoryStorage) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dataItem, exists := m.data[key]
	var currentVal int64 = 0

	if exists {
		if !dataItem.expiresAt.IsZero() && dataItem.expiresAt.Before(time.Now()) {
			exists = false
		} else {
			// Parse current value
			var err error
			_, err = parseIntValue(dataItem.value)
			if err != nil {
				return 0, ErrInvalidValue
			}
		}
	}

	newVal := currentVal + value

	expiresAt := time.Time{}
	if exists {
		expiresAt = dataItem.expiresAt
	}

	m.data[key] = &item{
		value:     intToString(newVal),
		expiresAt: expiresAt,
	}

	return newVal, nil
}

func (m *MemoryStorage) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return m.Set(ctx, key, value, ttl)
}

func (m *MemoryStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return -2 * time.Second, nil // Redis convention: -2 for non-existent key
	}

	if item.expiresAt.IsZero() {
		return -1 * time.Second, nil // Redis convention: -1 for no expiration
	}

	ttl := time.Until(item.expiresAt)
	if ttl < 0 {
		return -2 * time.Second, nil
	}

	return ttl, nil
}

func (m *MemoryStorage) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists {
		return ErrKeyNotFound
	}

	item.expiresAt = time.Now().Add(ttl)
	return nil
}

// List operations
func (m *MemoryStorage) ListPush(ctx context.Context, key string, values ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, exists := m.lists[key]
	if !exists {
		list = &listItem{values: []string{}}
		m.lists[key] = list
	}

	list.values = append(list.values, values...)
	return nil
}

func (m *MemoryStorage) ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, exists := m.lists[key]
	if !exists {
		return []string{}, nil
	}

	length := int64(len(list.values))

	// Handle negative indices
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// Bounds checking
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []string{}, nil
	}

	return list.values[start : stop+1], nil
}

func (m *MemoryStorage) ListTrim(ctx context.Context, key string, start, stop int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, exists := m.lists[key]
	if !exists {
		return nil
	}

	length := int64(len(list.values))

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
		list.values = []string{}
	} else {
		list.values = list.values[start : stop+1]
	}

	return nil
}

func (m *MemoryStorage) ListLen(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, exists := m.lists[key]
	if !exists {
		return 0, nil
	}

	return int64(len(list.values)), nil
}

// Sorted Set operations
func (m *MemoryStorage) ZAdd(ctx context.Context, key string, score float64, member string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zset := m.zsets[key]

	// Remove existing member if present
	for i, zm := range zset {
		if zm.member == member {
			zset = append(zset[:i], zset[i+1:]...)
			break
		}
	}

	// Add new member
	zset = append(zset, zsetMember{score: score, member: member})

	// Sort by score
	sort.Slice(zset, func(i, j int) bool {
		return zset[i].score < zset[j].score
	})

	m.zsets[key] = zset
	return nil
}

func (m *MemoryStorage) ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	zset := m.zsets[key]
	var newZset []zsetMember
	removed := int64(0)

	for _, zm := range zset {
		if zm.score < min || zm.score > max {
			newZset = append(newZset, zm)
		} else {
			removed++
		}
	}

	m.zsets[key] = newZset
	return removed, nil
}

func (m *MemoryStorage) ZCount(ctx context.Context, key string, min, max float64) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	zset := m.zsets[key]
	count := int64(0)

	for _, zm := range zset {
		if zm.score >= min && zm.score <= max {
			count++
		}
	}

	return count, nil
}

func (m *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MemoryStorage) Close() error {
	return nil
}

func (m *MemoryStorage) FlushAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]*item)
	m.lists = make(map[string]*listItem)
	m.zsets = make(map[string][]zsetMember)
	return nil
}

// Helper functions
func parseIntValue(s string) (int64, error) {
	var val int64
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

func intToString(i int64) string {
	return fmt.Sprintf("%d", i)
}
