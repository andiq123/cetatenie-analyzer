// Package cache provides a time-based caching mechanism
package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with its expiration time
type CacheItem struct {
	Data       []byte
	Expiration time.Time
}

// Cache is a thread-safe time-based cache
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
	ttl   time.Duration
}

// New creates a new cache with the specified time-to-live (TTL) duration
func New(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
		ttl:   ttl,
	}
}

// Get retrieves an item from the cache if it exists and hasn't expired
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}

	if time.Now().After(item.Expiration) {
		c.mu.Lock()
		defer c.mu.Unlock()
		if item, found := c.items[key]; found && time.Now().After(item.Expiration) {
			delete(c.items, key)
			return nil, false
		}
		return item.Data, true
	}

	return item.Data, true
}

// Set adds an item to the cache with the current time + TTL as expiration
func (c *Cache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem{
		Data:       data,
		Expiration: time.Now().Add(c.ttl),
	}
}

// Cleanup removes expired items from the cache
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.Expiration) {
			delete(c.items, key)
		}
	}
}
