package ipxpress

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// Cache defines the interface for response caching.
type Cache interface {
	Get(key string) (*CacheEntry, bool)
	Set(key string, entry *CacheEntry)
	Cleanup()
}

// CacheEntry represents a cached response.
type CacheEntry struct {
	ContentType string
	Data        []byte
	StatusCode  int
	ErrorMsg    string
	Timestamp   time.Time
}

// InMemoryCache is a simple in-memory cache implementation with TTL.
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewInMemoryCache creates a new in-memory cache with the given TTL.
func NewInMemoryCache(ttl time.Duration) *InMemoryCache {
	return &InMemoryCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cache entry by key. Returns the entry and true if found and not expired.
func (c *InMemoryCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(entry.Timestamp) > c.ttl {
		return nil, false
	}

	return entry, true
}

// Set stores a cache entry with the given key.
func (c *InMemoryCache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Stamp the entry time so TTL can be enforced correctly
	entry.Timestamp = time.Now()
	c.entries[key] = entry
}

// Cleanup removes expired entries from the cache.
func (c *InMemoryCache) Cleanup() {
	c.mu.RLock()
	now := time.Now()
	keysToDelete := make([]string, 0)

	for key, entry := range c.entries {
		if now.Sub(entry.Timestamp) > c.ttl {
			keysToDelete = append(keysToDelete, key)
		}
	}
	c.mu.RUnlock()

	// Only lock for writing if there are keys to delete
	if len(keysToDelete) > 0 {
		c.mu.Lock()
		for _, key := range keysToDelete {
			delete(c.entries, key)
		}
		c.mu.Unlock()
	}
}

// GenerateCacheKey generates a cache key from request parameters.
func GenerateCacheKey(imageURL string, width, height, quality int, format Format) string {
	h := md5.Sum([]byte(fmt.Sprintf("%s|%d|%d|%d|%s", imageURL, width, height, quality, format)))
	return fmt.Sprintf("%x", h)
}
