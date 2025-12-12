package ipxpress

import (
	"fmt"
	"net/http"
)

// Handler handles image processing requests.
type Handler struct {
	cache           Cache
	fetcher         *Fetcher
	config          *Config
	processingLimit chan struct{}
}

// NewHandler creates a new Handler with the given configuration.
func NewHandler(config *Config) *Handler {
	if config == nil {
		config = DefaultConfig()
	}

	return &Handler{
		cache:           NewInMemoryCache(config.CacheTTL),
		fetcher:         NewFetcher(),
		config:          config,
		processingLimit: make(chan struct{}, config.ProcessingLimit),
	}
}

// ServeHTTP handles HTTP requests for image processing.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	params := ParseProcessingParams(r)

	// Generate cache key
	cacheKey := GenerateCacheKey(params.URL, params.Width, params.Height, params.Quality, params.Format)

	// Check cache first
	if entry, found := h.cache.Get(cacheKey); found {
		h.writeResponse(w, entry)
		return
	}

	// Cache miss - fetch and process
	// STAGE 1: Fetch image (network I/O - parallel for all requests)
	imageData, err := h.fetcher.Fetch(params.URL)
	if err != nil {
		entry := h.createErrorEntry(err)
		h.cache.Set(cacheKey, entry)
		h.writeResponse(w, entry)
		return
	}

	// STAGE 2: Acquire semaphore for libvips processing
	h.processingLimit <- struct{}{}
	defer func() { <-h.processingLimit }()

	// Process with libvips (serialize via semaphore)
	entry := h.processImage(imageData, params)

	// Cache the result
	h.cache.Set(cacheKey, entry)

	// Write response
	h.writeResponse(w, entry)
}

// Server returns an http.Handler that processes images from URLs.
// Expected query parameters:
// - url: the URL of the image to process (required)
// - w: maximum width
// - h: maximum height
// - quality: output quality (1-100, default 85)
// - format: output format (jpeg, png, gif, webp) - defaults to original format
func Server() http.Handler {
	return NewHandler(DefaultConfig())
}

// ServerWithConfig returns an http.Handler with custom configuration.
func ServerWithConfig(config *Config) http.Handler {
	return NewHandler(config)
}

// createErrorEntry creates a cache entry from an error.
func (h *Handler) createErrorEntry(err error) *CacheEntry {
	if fetchErr, ok := err.(*FetchError); ok {
		return &CacheEntry{
			StatusCode: fetchErr.StatusCode,
			ErrorMsg:   fetchErr.Message,
		}
	}
	return &CacheEntry{
		StatusCode: http.StatusInternalServerError,
		ErrorMsg:   err.Error(),
	}
}

// processImage processes fetched image data with libvips transformations.
func (h *Handler) processImage(imageData []byte, params *ProcessingParams) *CacheEntry {
	proc := New().FromBytes(imageData)
	origFormat := proc.OriginalFormat()

	// If no transformation parameters are specified, return original image
	if !params.NeedsProcessing(origFormat) {
		return &CacheEntry{
			ContentType: "application/octet-stream",
			Data:        imageData,
			StatusCode:  http.StatusOK,
		}
	}

	// Determine output format
	outputFormat := params.GetOutputFormat(origFormat)

	// Apply resize only if dimensions are specified
	if params.Width > 0 || params.Height > 0 {
		proc = proc.Resize(params.Width, params.Height)
	}

	if err := proc.Err(); err != nil {
		proc.Close()
		return &CacheEntry{
			StatusCode: http.StatusInternalServerError,
			ErrorMsg:   fmt.Sprintf("processing: %v", err),
		}
	}

	out, err := proc.ToBytes(outputFormat, params.Quality)
	proc.Close() // Free memory immediately after processing
	if err != nil {
		return &CacheEntry{
			StatusCode: http.StatusInternalServerError,
			ErrorMsg:   fmt.Sprintf("encode: %v", err),
		}
	}

	return &CacheEntry{
		ContentType: outputFormat.ContentType(),
		Data:        out,
		StatusCode:  http.StatusOK,
	}
}

// writeResponse writes a cache entry to the HTTP response writer.
func (h *Handler) writeResponse(w http.ResponseWriter, entry *CacheEntry) {
	if entry.ErrorMsg != "" {
		w.WriteHeader(entry.StatusCode)
		w.Write([]byte(entry.ErrorMsg))
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(entry.Data)))
	if entry.ContentType != "" {
		w.Header().Set("Content-Type", entry.ContentType)
		if entry.ContentType == "application/octet-stream" {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=604800")
		}
	}

	w.WriteHeader(entry.StatusCode)
	w.Write(entry.Data)
}

// CleanupCache removes expired entries from the handler's cache.
func (h *Handler) CleanupCache() {
	h.cache.Cleanup()
}
