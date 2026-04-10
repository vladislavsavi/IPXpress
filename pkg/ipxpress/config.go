package ipxpress

import (
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

// VipsConfig holds vips-specific configuration.
type VipsConfig struct {
	// MaxCacheMem is the maximum memory to use for caching (in MB)
	MaxCacheMem int

	// MaxCacheSize is the maximum number of operations to cache
	MaxCacheSize int

	// MaxCacheFiles is the maximum number of files to have open
	MaxCacheFiles int

	// LogLevel controls vips logging verbosity
	LogLevel vips.LogLevel
}

// DefaultVipsConfig returns default vips configuration.
func DefaultVipsConfig() *VipsConfig {
	return &VipsConfig{
		MaxCacheMem:      0, // Disable libvips caching (we manage cache at application level)
		MaxCacheSize:     0, // Disable libvips caching
		MaxCacheFiles:    0,
		LogLevel:         vips.LogLevelWarning,
	}
}

// Config holds the server configuration.
type Config struct {
	// CacheTTL is the duration to keep cached responses
	CacheTTL time.Duration

	// CacheMaxCost is the maximum total cost of the cache (usually in bytes).
	// Otter uses this to perform cost-based eviction.
	CacheMaxCost int

	// ProcessingLimit is the maximum number of concurrent image processing operations
	ProcessingLimit int

	// CleanupInterval is the interval for cache cleanup (maintained for compatibility,
	// though Otter manages cleanup internally).
	CleanupInterval time.Duration

	// VipsConfig holds libvips-specific configuration
	// If nil, default vips settings will be used
	VipsConfig *VipsConfig

	// ClientMaxAge controls Cache-Control max-age for clients (in seconds)
	ClientMaxAge int

	// SMaxAge controls Cache-Control s-maxage for shared caches/CDNs (in seconds). 0 disables.
	SMaxAge int

	// EnableETag enables ETag generation and If-None-Match handling
	EnableETag bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		CacheTTL:        10 * time.Minute,
		CacheMaxCost:    512 * 1024 * 1024, // 512 MB
		ProcessingLimit: 256,
		CleanupInterval: 30 * time.Second,
		VipsConfig:      nil,    // Will use default vips settings
		ClientMaxAge:    604800, // 7 days
		SMaxAge:         0,
		EnableETag:      true,
	}
}

// NewDefaultConfig is an alias for DefaultConfig to improve discoverability
// for library clients who look for a constructor-style helper.
func NewDefaultConfig() *Config { return DefaultConfig() }
