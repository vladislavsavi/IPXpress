package ipxpress

import (
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Global HTTP client with connection pooling for image fetching
var httpClient = &http.Client{
	Timeout: 20 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     256,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
}

// Response cache entry
type cacheEntry struct {
	contentType string
	data        []byte
	statusCode  int
	errMsg      string
	timestamp   time.Time
}

// Response cache with TTL
var responseCache = struct {
	sync.RWMutex
	entries map[string]*cacheEntry
}{
	entries: make(map[string]*cacheEntry),
}

// Cache TTL - how long to keep cached responses (30 seconds)
const cacheTTL = 30 * time.Second

// getCacheKey generates a cache key from request parameters
func getCacheKey(imageURL, width, height, quality, format string) string {
	h := md5.Sum([]byte(imageURL + "|" + width + "|" + height + "|" + quality + "|" + format))
	return fmt.Sprintf("%x", h)
}

// Server returns an http.Handler that processes images from URLs and applies
// transformations using the ipxpress Processor.
// Expected query parameters:
// - url: the URL of the image to process (required)
// - w: maximum width
// - h: maximum height
// - quality: output quality (for JPEG)
// - format: output format (jpeg, png) - defaults to jpeg
func Server() http.Handler {
	// Semaphore to limit concurrent libvips operations
	// For 3000 req/sec target, need aggressive parallelism
	// Allow up to 256 concurrent operations (sufficient for heavy load)
	processingLimit := make(chan struct{}, 256)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract parameters for cache key
		q := r.URL.Query()
		imageURL := q.Get("url")
		widthStr := q.Get("w")
		heightStr := q.Get("h")
		qualityStr := q.Get("quality")
		formatParam := q.Get("format")

		// Check cache first
		cacheKey := getCacheKey(imageURL, widthStr, heightStr, qualityStr, formatParam)
		responseCache.RLock()
		if entry, exists := responseCache.entries[cacheKey]; exists && time.Since(entry.timestamp) < cacheTTL {
			responseCache.RUnlock()
			// Cache hit - serve from cache
			if entry.errMsg != "" {
				w.WriteHeader(entry.statusCode)
				w.Write([]byte(entry.errMsg))
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(entry.data)))
			if entry.contentType != "" {
				w.Header().Set("Content-Type", entry.contentType)
				if entry.contentType == "application/octet-stream" {
					w.Header().Set("Cache-Control", "public, max-age=31536000")
				} else {
					w.Header().Set("Cache-Control", "public, max-age=604800")
				}
			}
			w.WriteHeader(entry.statusCode)
			w.Write(entry.data)
			return
		}
		responseCache.RUnlock()

		// Cache miss - fetch and process
		// STAGE 1: Fetch image (network I/O - parallel for all requests)
		imageData, err := fetchImageData(r)
		if err != nil {
			w.WriteHeader(err.statusCode)
			w.Write([]byte(err.message))
			return
		}

		// STAGE 2: Acquire semaphore for libvips processing
		processingLimit <- struct{}{}
		defer func() { <-processingLimit }()

		// Process with libvips (serialize via semaphore)
		contentType, data, statusCode, errMsg := processImageData(r, imageData)

		// Cache the result
		responseCache.Lock()
		responseCache.entries[cacheKey] = &cacheEntry{
			contentType: contentType,
			data:        data,
			statusCode:  statusCode,
			errMsg:      errMsg,
			timestamp:   time.Now(),
		}
		responseCache.Unlock()

		// Write response
		if errMsg != "" {
			w.WriteHeader(statusCode)
			w.Write([]byte(errMsg))
			return
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
			if contentType == "application/octet-stream" {
				w.Header().Set("Cache-Control", "public, max-age=31536000")
			} else {
				w.Header().Set("Cache-Control", "public, max-age=604800")
			}
		}

		w.WriteHeader(statusCode)
		w.Write(data)
	})
}

type fetchError struct {
	statusCode int
	message    string
}

// fetchImageData fetches image data from URL (network I/O only, no libvips)
func fetchImageData(r *http.Request) ([]byte, *fetchError) {
	q := r.URL.Query()
	imageURL := q.Get("url")
	if imageURL == "" {
		return nil, &fetchError{http.StatusBadRequest, "missing image URL"}
	}

	// validate URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return nil, &fetchError{http.StatusBadRequest, fmt.Sprintf("invalid image URL: %v", err)}
	}
	if parsedURL.Scheme == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, &fetchError{http.StatusBadRequest, "image URL must use http or https"}
	}

	// fetch image from URL with User-Agent header
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, &fetchError{http.StatusBadRequest, fmt.Sprintf("invalid URL: %v", err)}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Use global HTTP client with connection pooling
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, &fetchError{http.StatusBadRequest, fmt.Sprintf("failed to fetch image: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &fetchError{http.StatusBadRequest, fmt.Sprintf("image fetch failed with status %d", resp.StatusCode)}
	}

	// read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &fetchError{http.StatusInternalServerError, fmt.Sprintf("failed to read image data: %v", err)}
	}

	return imageData, nil
}

// processImageData processes fetched image data with libvips transformations
func processImageData(r *http.Request, imageData []byte) (contentType string, data []byte, statusCode int, errMsg string) {
	q := r.URL.Query()
	wv := parseInt(q.Get("w"))
	hv := parseInt(q.Get("h"))
	quality := parseInt(q.Get("quality"))
	formatParam := q.Get("format")

	// If no transformation parameters are specified, return original image
	if wv == 0 && hv == 0 && quality == 0 && formatParam == "" {
		return "application/octet-stream", imageData, http.StatusOK, ""
	}

	proc := New().FromBytes(imageData)

	origFormat := proc.OriginalFormat()

	// Determine format: use specified format, or original format as fallback
	format := formatParam
	if format == "" {
		// Use original format if no format specified
		format = origFormat
		if format == "" {
			format = "jpeg" // fallback
		}
	}

	// normalize format
	format = strings.ToLower(format)
	if format == "jpg" {
		format = "jpeg"
	}
	// Validate format - allow jpeg, png, gif, webp
	if format != "jpeg" && format != "png" && format != "gif" && format != "webp" {
		format = "jpeg"
	}

	// Set quality: if not specified, use good quality
	if quality <= 0 || quality > 100 {
		quality = 85 // Good default
	}

	// Apply resize only if dimensions are specified
	if wv > 0 || hv > 0 {
		proc = proc.Resize(wv, hv)
	}

	if err := proc.Err(); err != nil {
		proc.Close()
		return "", nil, http.StatusInternalServerError, fmt.Sprintf("processing: %v", err)
	}

	out, err := proc.ToBytes(format, quality)
	proc.Close() // Free memory immediately after processing
	if err != nil {
		return "", nil, http.StatusInternalServerError, fmt.Sprintf("encode: %v", err)
	}

	// content type
	var ctype string
	switch format {
	case "png":
		ctype = "image/png"
	case "webp":
		ctype = "image/webp"
	case "gif":
		ctype = "image/gif"
	default:
		ctype = "image/jpeg"
	}

	return ctype, out, http.StatusOK, ""
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

// CleanupCache removes expired entries from response cache
func CleanupCache() {
	responseCache.RLock()
	now := time.Now()
	keysToDelete := []string{}

	for key, entry := range responseCache.entries {
		if now.Sub(entry.timestamp) > cacheTTL {
			keysToDelete = append(keysToDelete, key)
		}
	}
	responseCache.RUnlock()

	// Only lock for writing if there are keys to delete
	if len(keysToDelete) > 0 {
		responseCache.Lock()
		for _, key := range keysToDelete {
			delete(responseCache.entries, key)
		}
		responseCache.Unlock()
	}
}
