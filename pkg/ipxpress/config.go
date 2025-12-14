package ipxpress

import (
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

// VipsConfig holds vips-specific configuration.
type VipsConfig struct {
	// ConcurrencyLevel controls the number of threads libvips uses (0 = use all cores)
	ConcurrencyLevel int

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
		ConcurrencyLevel: 0,    // Use all available cores
		MaxCacheMem:      2048, // 2GB
		MaxCacheSize:     5000,
		MaxCacheFiles:    0,
		LogLevel:         vips.LogLevelWarning,
	}
}

// Config holds the server configuration.
type Config struct {
	// CacheTTL is the duration to keep cached responses
	CacheTTL time.Duration

	// ProcessingLimit is the maximum number of concurrent image processing operations
	ProcessingLimit int

	// CleanupInterval is the interval for cache cleanup
	CleanupInterval time.Duration

	// VipsConfig holds libvips-specific configuration
	// If nil, default vips settings will be used
	VipsConfig *VipsConfig
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		CacheTTL:        30 * time.Second,
		ProcessingLimit: 256,
		CleanupInterval: 30 * time.Second,
		VipsConfig:      nil, // Will use default vips settings
	}
}

// NewDefaultConfig is an alias for DefaultConfig to improve discoverability
// for library clients who look for a constructor-style helper.
func NewDefaultConfig() *Config { return DefaultConfig() }
