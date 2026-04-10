package ipxpress

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/maypok86/otter"
)

// CacheEntry represents a cached response.
type CacheEntry struct {
	ContentType string
	Data        []byte
	StatusCode  int
	ErrorMsg    string
	ETag        string
	Timestamp   time.Time
}

// InMemoryCache is an in-memory cache implementation backed by otter (W-TinyLFU algorithm).
// It supports cost-based eviction (by data size) and high-concurrency access.
type InMemoryCache struct {
	cache otter.Cache[string, *CacheEntry]
}

// NewInMemoryCache creates a new in-memory cache with the given TTL and capacity.
// It uses W-TinyLFU for high hit rates and low memory overhead.
// Capacity is treated as the number of items by default, but can be scaled for bytes.
func NewInMemoryCache(ttl time.Duration, capacity int) *InMemoryCache {
	if capacity <= 0 {
		capacity = 10000
	}

	// Build the cache with W-TinyLFU and cost-based eviction
	cache, err := otter.MustBuilder[string, *CacheEntry](capacity).
		CollectStats().
		Cost(func(key string, entry *CacheEntry) uint32 {
			// Cost is based on the data size plus metadata strings and overhead
			// This allows the cache to evict based on actual memory usage
			cost := uint32(len(entry.Data) + len(entry.ContentType) + len(entry.ErrorMsg) + len(entry.ETag)) + 256 // 256 bytes struct/node overhead estimate
			if cost == 0 {
				return 1 // Minimum cost must be 1
			}
			return cost
		}).
		WithTTL(ttl).
		Build()

	if err != nil {
		// Should not happen with MustBuilder unless something is fundamentally wrong
		panic(fmt.Sprintf("failed to build otter cache: %v", err))
	}

	return &InMemoryCache{
		cache: cache,
	}
}

// Get retrieves a cache entry by key. Returns the entry and true if found and not expired.
func (c *InMemoryCache) Get(key string) (*CacheEntry, bool) {
	return c.cache.Get(key)
}

// Set stores a cache entry with the given key.
// The entry will be automatically removed after the TTL expires.
func (c *InMemoryCache) Set(key string, entry *CacheEntry) {
	// Stamp the entry time for reference
	entry.Timestamp = time.Now()
	c.cache.Set(key, entry)
}

// Close closes the cache and releases resources.
func (c *InMemoryCache) Close() {
	c.cache.Close()
}

// GenerateCacheKey generates a cache key from all request parameters to avoid collisions.
func GenerateCacheKey(p *ProcessingParams) string {
	// Include all parameters that affect the output image to ensure correct caching.
	// We use | as separator to avoid ambiguity between parameter values.
	key := fmt.Sprintf("%s|%d|%d|%d|%s|%s|%s|%s|%t|%f|%s|%d|%t|%t|%t|%s|%d|%s|%s|%t|%t|%d|%s|%f|%d|%s|%t",
		p.URL, p.Width, p.Height, p.Quality, p.Format,
		p.Fit, p.Position, p.Kernel, p.Enlarge,
		p.Blur, p.Sharpen, p.Rotate, p.Flip, p.Flop, p.Grayscale,
		p.Extract, p.Trim, p.Extend,
		p.Background, p.Negate, p.Normalize, p.Threshold, p.Tint, p.Gamma, p.Median, p.Modulate, p.Flatten)

	h := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", h)
}
