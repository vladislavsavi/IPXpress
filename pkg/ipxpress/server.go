package ipxpress

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

// ProcessorFunc is a function that processes an image.
// It receives the processor and params, and can apply custom transformations.
type ProcessorFunc func(*Processor, *ProcessingParams) *Processor

// MiddlewareFunc is a function that wraps the handler with additional functionality.
type MiddlewareFunc func(http.Handler) http.Handler

// Handler handles image processing requests.
type Handler struct {
	cache           Cache
	fetcher         *Fetcher
	config          *Config
	processingLimit chan struct{}
	processors      []ProcessorFunc
	middlewares     []MiddlewareFunc
}

// NewHandler creates a new Handler with the given configuration.
// Automatically initializes vips if not already initialized.
// If config.VipsConfig is provided, vips will be initialized with those settings.
func NewHandler(config *Config) *Handler {
	if config == nil {
		config = DefaultConfig()
	}

	// Initialize vips with custom config if provided
	if config.VipsConfig != nil {
		initVipsWithSettings(config.VipsConfig)
	} else {
		initVips()
	}

	return &Handler{
		cache:           NewInMemoryCache(config.CacheTTL),
		fetcher:         NewFetcher(),
		config:          config,
		processingLimit: make(chan struct{}, config.ProcessingLimit),
		processors:      []ProcessorFunc{},
		middlewares:     []MiddlewareFunc{},
	}
}

// UseProcessor adds a custom processor function to the processing pipeline.
// Processors are executed after the built-in transformations.
func (h *Handler) UseProcessor(processor ProcessorFunc) *Handler {
	h.processors = append(h.processors, processor)
	return h
}

// UseMiddleware adds a middleware to wrap the handler.
func (h *Handler) UseMiddleware(middleware MiddlewareFunc) *Handler {
	h.middlewares = append(h.middlewares, middleware)
	return h
}

// applyMiddlewares wraps the handler with all registered middlewares.
func (h *Handler) applyMiddlewares(handler http.Handler) http.Handler {
	// Apply middlewares in reverse order so they execute in the order they were added
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handler = h.middlewares[i](handler)
	}
	return handler
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

	// Apply built-in operations in order (order matters for image processing)
	proc = h.applyBuiltInTransformations(proc, params)

	// Apply custom processors
	for _, processor := range h.processors {
		proc = processor(proc, params)
	}

	// Check for errors
	if err := proc.Err(); err != nil {
		proc.Close()
		return &CacheEntry{
			StatusCode: http.StatusInternalServerError,
			ErrorMsg:   fmt.Sprintf("processing: %v", err),
		}
	}

	// Encode to output format
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

// applyBuiltInTransformations applies the standard image transformations.
func (h *Handler) applyBuiltInTransformations(proc *Processor, params *ProcessingParams) *Processor {

	// 1. Extract/Crop (do this first to reduce data to process)
	if params.Extract != "" {
		parts := strings.Split(params.Extract, "_")
		if len(parts) == 4 {
			left, _ := strconv.Atoi(parts[0])
			top, _ := strconv.Atoi(parts[1])
			width, _ := strconv.Atoi(parts[2])
			height, _ := strconv.Atoi(parts[3])
			proc = proc.Extract(left, top, width, height)
		}
	}

	// 2. Resize
	if params.Width > 0 || params.Height > 0 {
		kernel := params.GetVipsKernel()
		proc = proc.ResizeWithOptions(params.Width, params.Height, kernel, params.Enlarge)
	}

	// 3. Extend (add borders)
	if params.Extend != "" {
		parts := strings.Split(params.Extend, "_")
		if len(parts) == 4 {
			top, _ := strconv.Atoi(parts[0])
			right, _ := strconv.Atoi(parts[1])
			bottom, _ := strconv.Atoi(parts[2])
			left, _ := strconv.Atoi(parts[3])

			var bgColor []float64
			if params.Background != "" {
				bgColor = hexToRGB(params.Background)
			}
			proc = proc.Extend(top, right, bottom, left, bgColor)
		}
	}

	// 4. Rotate
	if params.Rotate != 0 {
		angle := angleToVips(params.Rotate)
		proc = proc.Rotate(angle)
	}

	// 5. Flip/Flop
	if params.Flip {
		proc = proc.Flip()
	}
	if params.Flop {
		proc = proc.Flop()
	}

	// 6. Blur
	if params.Blur > 0 {
		proc = proc.Blur(params.Blur)
	}

	// 7. Sharpen
	if params.Sharpen != "" {
		parts := strings.Split(params.Sharpen, "_")
		sigma, flat, jagged := 1.0, 1.0, 2.0
		if len(parts) >= 1 {
			if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
				sigma = v
			}
		}
		if len(parts) >= 2 {
			if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
				flat = v
			}
		}
		if len(parts) >= 3 {
			if v, err := strconv.ParseFloat(parts[2], 64); err == nil {
				jagged = v
			}
		}
		proc = proc.Sharpen(sigma, flat, jagged)
	}

	// 8. Color operations
	if params.Grayscale {
		proc = proc.Grayscale()
	}

	if params.Negate {
		proc = proc.Negate()
	}

	if params.Normalize {
		proc = proc.Normalize()
	}

	if params.Gamma > 0 {
		proc = proc.Gamma(params.Gamma)
	}

	if params.Modulate != "" {
		parts := strings.Split(params.Modulate, "_")
		brightness, saturation, hue := 1.0, 1.0, 0.0
		if len(parts) >= 1 {
			if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
				brightness = v
			}
		}
		if len(parts) >= 2 {
			if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
				saturation = v
			}
		}
		if len(parts) >= 3 {
			if v, err := strconv.ParseFloat(parts[2], 64); err == nil {
				hue = v
			}
		}
		proc = proc.Modulate(brightness, saturation, hue)
	}

	// 9. Flatten (remove alpha)
	if params.Flatten {
		var bgColor *vips.Color
		if params.Background != "" {
			rgb := hexToRGB(params.Background)
			if len(rgb) >= 3 {
				bgColor = &vips.Color{
					R: uint8(rgb[0]),
					G: uint8(rgb[1]),
					B: uint8(rgb[2]),
				}
			}
		}
		proc = proc.Flatten(bgColor)
	}

	return proc
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

// hexToRGB converts hex color string to RGB values
func hexToRGB(hex string) []float64 {
	hex = strings.TrimPrefix(hex, "#")

	// Handle 3-digit hex
	if len(hex) == 3 {
		hex = string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
	}

	if len(hex) != 6 {
		return []float64{255, 255, 255} // default to white
	}

	r, err1 := strconv.ParseUint(hex[0:2], 16, 8)
	g, err2 := strconv.ParseUint(hex[2:4], 16, 8)
	b, err3 := strconv.ParseUint(hex[4:6], 16, 8)

	if err1 != nil || err2 != nil || err3 != nil {
		return []float64{255, 255, 255} // default to white
	}

	return []float64{float64(r), float64(g), float64(b)}
}

// angleToVips converts rotation angle to vips.Angle
func angleToVips(angle int) vips.Angle {
	// Normalize angle to 0-359
	angle = angle % 360
	if angle < 0 {
		angle += 360
	}

	switch angle {
	case 90:
		return vips.Angle90
	case 180:
		return vips.Angle180
	case 270:
		return vips.Angle270
	default:
		return vips.Angle0
	}
}
