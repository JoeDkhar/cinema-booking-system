package cache

import (
	"sync"
	"time"
)

// Cache is a generic in-memory cache with expiration
type Cache[T any] struct {
	items map[string]cacheItem[T]
	mutex sync.RWMutex
}

// cacheItem represents a single item in the cache
type cacheItem[T any] struct {
	value      T
	expiration time.Time
}

// NewCache creates a new generic cache
func NewCache[T any]() *Cache[T] {
	cache := &Cache[T]{
		items: make(map[string]cacheItem[T]),
	}

	// Start garbage collection in the background
	go cache.startGC()

	return cache
}

// Set adds an item to the cache with expiration duration
func (c *Cache[T]) Set(key string, value T, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(duration)
	c.items[key] = cacheItem[T]{
		value:      value,
		expiration: expiration,
	}
}

// Get retrieves an item from the cache
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found {
		var zero T
		return zero, false
	}

	// Check if the item has expired
	if time.Now().After(item.expiration) {
		var zero T
		return zero, false
	}

	return item.value, true
}

// Delete removes an item from the cache
func (c *Cache[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]cacheItem[T])
}

// startGC starts the garbage collection process
func (c *Cache[T]) startGC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()

		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiration) {
				delete(c.items, key)
			}
		}

		c.mutex.Unlock()
	}
}
