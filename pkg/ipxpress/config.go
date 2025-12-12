package ipxpress

import "time"

// Config holds the server configuration.
type Config struct {
	// CacheTTL is the duration to keep cached responses
	CacheTTL time.Duration

	// ProcessingLimit is the maximum number of concurrent image processing operations
	ProcessingLimit int

	// CleanupInterval is the interval for cache cleanup
	CleanupInterval time.Duration
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		CacheTTL:        30 * time.Second,
		ProcessingLimit: 256,
		CleanupInterval: 30 * time.Second,
	}
}
