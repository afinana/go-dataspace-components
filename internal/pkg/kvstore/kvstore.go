package kvstore

import (
	"context"
	"sync"
	"time"
)

// KVStore defines the explicit port interface for high-performance key-value operations.
type KVStore interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}

type item struct {
	value     []byte
	expiresAt time.Time
}

// MemoryKVStore is a thread-safe, high-performance in-memory key-value cache with TTL eviction.
type MemoryKVStore struct {
	mu          sync.RWMutex
	store       map[string]item
	stopCleanup chan struct{}
}

// NewMemoryKVStore initializes a MemoryKVStore instance and starts background eviction cleanup.
func NewMemoryKVStore(cleanupInterval time.Duration) *MemoryKVStore {
	kv := &MemoryKVStore{
		store:       make(map[string]item),
		stopCleanup: make(chan struct{}),
	}

	if cleanupInterval > 0 {
		go kv.startCleanup(cleanupInterval)
	}

	return kv
}

// Set stores a key-value pair with an optional time-to-live (TTL).
func (m *MemoryKVStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	valCopy := make([]byte, len(value))
	copy(valCopy, value)

	m.store[key] = item{
		value:     valCopy,
		expiresAt: expiresAt,
	}

	return nil
}

// Get retrieves a key from the cache. Returns value, found status, and error.
func (m *MemoryKVStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	m.mu.RLock()
	it, exists := m.store[key]
	m.mu.RUnlock()

	if !exists {
		return nil, false, nil
	}

	if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
		// Expired item: delete lazily
		m.mu.Lock()
		delete(m.store, key)
		m.mu.Unlock()
		return nil, false, nil
	}

	valCopy := make([]byte, len(it.value))
	copy(valCopy, it.value)
	return valCopy, true, nil
}

// Delete removes a key from the key-value store.
func (m *MemoryKVStore) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, key)
	return nil
}

// Clear removes all keys from the key-value store.
func (m *MemoryKVStore) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = make(map[string]item)
	return nil
}

// Close stops the background eviction goroutine.
func (m *MemoryKVStore) Close() {
	close(m.stopCleanup)
}

func (m *MemoryKVStore) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.evictExpired()
		case <-m.stopCleanup:
			return
		}
	}
}

func (m *MemoryKVStore) evictExpired() {
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.store {
		if !v.expiresAt.IsZero() && now.After(v.expiresAt) {
			delete(m.store, k)
		}
	}
}
