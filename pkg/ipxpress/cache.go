package ipxpress

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

// CacheEntry represents a cached response.
type CacheEntry struct {
	ContentType string
	Data        []byte
	StatusCode  int
	ErrorMsg    string
	Timestamp   time.Time
}

// InMemoryCache is an in-memory cache implementation backed by golang-lru with expirable TTL support.
type InMemoryCache struct {
	lru *expirable.LRU[string, *CacheEntry]
}

// NewInMemoryCache creates a new in-memory cache with the given TTL and capacity using golang-lru expirable.
// Entries are automatically evicted when they expire or when the cache reaches capacity (LRU eviction).
func NewInMemoryCache(ttl time.Duration, capacity int) *InMemoryCache {
	if capacity <= 0 {
		capacity = 10000 // Fallback to default if invalid capacity is provided
	}
	cache := expirable.NewLRU[string, *CacheEntry](capacity, nil, ttl)
	return &InMemoryCache{
		lru: cache,
	}
}

// Get retrieves a cache entry by key. Returns the entry and true if found and not expired.
// The expirable cache automatically removes expired entries.
func (c *InMemoryCache) Get(key string) (*CacheEntry, bool) {
	entry, exists := c.lru.Get(key)
	if !exists {
		return nil, false
	}
	return entry, true
}

// Set stores a cache entry with the given key.
// The entry will be automatically removed after the TTL expires.
func (c *InMemoryCache) Set(key string, entry *CacheEntry) {
	// Stamp the entry time for reference (not used for expiration)
	entry.Timestamp = time.Now()
	c.lru.Add(key, entry)
}

// GenerateCacheKey generates a cache key from request parameters.
func GenerateCacheKey(imageURL string, width, height, quality int, format Format) string {
	h := md5.Sum([]byte(fmt.Sprintf("%s|%d|%d|%d|%s", imageURL, width, height, quality, format)))
	return fmt.Sprintf("%x", h)
}
