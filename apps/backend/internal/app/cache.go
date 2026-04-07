package app

import (
	"sync"
	"time"
)

type cacheEntry[T any] struct {
	value     T
	expiresAt time.Time
}

type MemCache[T any] struct {
	mu    sync.RWMutex
	items map[string]cacheEntry[T]
	ttl   time.Duration
}

func NewMemCache[T any](ttl time.Duration) *MemCache[T] {
	c := &MemCache[T]{
		items: make(map[string]cacheEntry[T]),
		ttl:   ttl,
	}
	go c.evictLoop()
	return c
}

func (c *MemCache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		var zero T
		return zero, false
	}
	return entry.value, true
}

func (c *MemCache[T]) Set(key string, value T) {
	c.mu.Lock()
	c.items[key] = cacheEntry[T]{value: value, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func (c *MemCache[T]) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *MemCache[T]) evictLoop() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, v := range c.items {
			if now.After(v.expiresAt) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
