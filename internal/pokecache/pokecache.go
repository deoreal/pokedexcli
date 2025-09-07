// Package pokecache creates a cache for pokedex data
package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	cache    map[string]CacheEntry
	interval time.Duration
	mu       *sync.RWMutex
	stopChan chan struct{}
}

type CacheEntry struct {
	CreatedAt time.Time `json:"created_at"`
	Val       []byte    `json:"val"`
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		cache:    make(map[string]CacheEntry),
		interval: interval,
		mu:       &sync.RWMutex{},
		stopChan: make(chan struct{}),
	}

	// Start the reap loop in a goroutine
	go c.reapLoop()

	return c
}

func (c *Cache) Add(key string, val []byte) {
	ce := CacheEntry{
		CreatedAt: time.Now(),
		Val:       val,
	}

	c.mu.Lock()
	c.cache[key] = ce
	c.mu.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, ok := c.cache[key]
	c.mu.RUnlock()

	if !ok {
		return []byte{}, false
	}

	// Ensure we never return nil, always return empty slice instead
	if entry.Val == nil {
		return []byte{}, true
	}

	return entry.Val, true
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.reapExpired()
		}
	}
}

func (c *Cache) reapExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.cache {
		// If the entry is older than the interval, remove it
		if now.Sub(entry.CreatedAt) > c.interval {
			delete(c.cache, key)
		}
	}
}

func (c *Cache) Stop() {
	close(c.stopChan)
}

// GetInterval returns the cache interval (for testing)
func (c *Cache) GetInterval() time.Duration {
	return c.interval
}

// GetCacheMap returns a copy of the cache map (for testing)
func (c *Cache) GetCacheMap() map[string]CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheCopy := make(map[string]CacheEntry)
	for k, v := range c.cache {
		cacheCopy[k] = v
	}
	return cacheCopy
}

// Len returns the number of entries in the cache (for testing)
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}
