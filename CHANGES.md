# IPXpress Changes

## v0.3.0 (2026-02-11)

**Memory and caching optimization:**

- **Disable libvips internal cache**: `DefaultVipsConfig` now disables libvips caching by default (`MaxCacheMem = 0`, `MaxCacheSize = 0`).
- **Single source of truth for caching**: all caching responsibility is moved to the application layer (`InMemoryCache`), removing duplicate memory usage and making resource consumption predictable.
- **Config update**: default values for `VipsConfig` in `DefaultVipsConfig()` changed to disable libvips cache.
- **Documentation**: architecture docs updated to reflect the caching strategy change.

These changes prevent images from being stored twice (application cache and libvips cache) and allow more precise memory usage control via `CacheTTL` and `CleanupInterval`.

## v0.2.0 (2025-12-15)

**Expanded functionality for using any libvips function:**

- **Direct ImageRef access**: `ImageRef()` returns a `vips.ImageRef` so any libvips function can be used.
- **ApplyFunc method**: apply custom processing functions with automatic error handling and chaining.
- **VipsOperationBuilder (fluent API)**: chain operations with `Blur()`, `Sharpen()`, `Modulate()`, `Median()`, `Flatten()`, `Invert()`.
- **CustomOperation type**: reusable custom operations.
- **Built-in operations**: `GaussianBlurOperation()`, `EdgeDetectionOperation()`, `SepiaOperation()`, `BrightnessOperation()`, `SaturationOperation()`, `ContrastOperation()`.
- **Documentation**: new `CUSTOM_OPERATIONS.md` with full examples.
- **Unit tests**: full coverage in `extensions_test.go`.

Usage examples:
```go
// Direct access: img := proc.ImageRef()
// ApplyFunc: proc.ApplyFunc(func(img *vips.ImageRef) error { ... })
// Builder: builder.Blur(2.0).Sharpen(1.5, 0.5, 1.0)
```

## v0.1.0 (2025-12-14)

Key improvements and fixes:

- **Default configuration**: `NewDefaultConfig()` and auto-apply for `NewHandler(nil)`. Library users do not need to set it manually.
- **Internal cache fix**: `Timestamp` is set in `InMemoryCache.Set` for stable TTL and consistent cache hits.
- **HTTP caching**: configurable `Cache-Control` headers (`ClientMaxAge`, `SMaxAge`) and `ETag`/`If-None-Match` support (enabled by default).
- **Short parameter aliases**: compatibility with ipx v2 (`w,h,f,q,s,b,pos`), plus parsing `s=WxH`.
- **Content headers**: correct `Content-Type` and forced `Content-Disposition: inline` for display without download.
- **Encoder tuning**: WebP `ReductionEffort=4` (fast), AVIF `Speed=6` for speed/compression balance; no unexpected format downgrades.

Stability:

- All tests pass: `go test ./...` OK.
- Comparison with external ipx showed matching output sizes and faster local performance.

Configuration (new fields):

- `ClientMaxAge` - `Cache-Control: max-age` (seconds), default `604800`.
- `SMaxAge` - `Cache-Control: s-maxage` for CDNs, default `0` (disabled).
- `EnableETag` - enables `ETag` and `304 Not Modified`, default `true`.

Example:

```go
cfg := ipxpress.NewDefaultConfig()
cfg.ClientMaxAge = 3600 // 1 hour
cfg.SMaxAge = 3600      // 1 hour for CDN
cfg.CacheTTL = 10 * time.Minute
cfg.EnableETag = true
handler := ipxpress.NewHandler(cfg)
```

Documentation:

- Updated `README.md`, `README.library.md`, `API.md` - caching and default settings sections.

### 1. Extensible handler architecture

**New types:**
- `ProcessorFunc` - function for custom image processing
- `MiddlewareFunc` - function for adding middleware

**New Handler methods:**
```go
handler.UseProcessor(processorFunc)  // Add custom processor
handler.UseMiddleware(middleware)     // Add middleware
```

### 2. Built-in processors (`pkg/ipxpress/examples.go`)

- `AutoOrientProcessor()` - auto-orient via EXIF
- `StripMetadataProcessor()` - remove metadata for privacy
- `CompressionOptimizer()` - optimize compression for web

### 3. Built-in middleware

- `CORSMiddleware(origins)` - CORS headers
- `LoggingMiddleware(logger)` - request logging
- `RateLimitMiddleware(maxRequests)` - request rate limiting
- `AuthMiddleware(tokens)` - token authentication

### 4. Documentation

**New files:**
- `LIBRARY_USAGE.md` - full library usage documentation
- `README.library.md` - short library README
- `examples/library_usage/main.go` - working example

## Usage examples

### Simple integration

```go
handler := ipxpress.NewHandler(nil)
http.Handle("/img/", http.StripPrefix("/img/", handler))
```

### With custom processors

```go
handler := ipxpress.NewHandler(nil)
handler.UseProcessor(ipxpress.AutoOrientProcessor())
handler.UseProcessor(ipxpress.StripMetadataProcessor())

// Custom processor
customProc := func(proc *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    return proc.Sharpen(1.5, 1.0, 2.0)
}
handler.UseProcessor(customProc)
```

### With middleware

```go
handler := ipxpress.NewHandler(nil)
handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))
handler.UseMiddleware(ipxpress.AuthMiddleware([]string{"secret-token"}))
```

### Multiple handlers

```go
// Public handler with limits
publicHandler := ipxpress.NewHandler(config1)
publicHandler.UseMiddleware(ipxpress.RateLimitMiddleware(100))

// Private handler with auth
privateHandler := ipxpress.NewHandler(config2)
privateHandler.UseMiddleware(ipxpress.AuthMiddleware(tokens))

http.Handle("/public/img/", http.StripPrefix("/public/img/", publicHandler))
http.Handle("/private/img/", http.StripPrefix("/private/img/", privateHandler))
```

## Architectural improvements

### Before:
- Rigid structure without extensibility
- Only built-in transformations
- No middleware support
- Single handler for the entire server

### After:
- Flexible architecture with ProcessorFunc and MiddlewareFunc
- Add custom processors to the pipeline
- Middleware support at the HTTP layer
- Multiple handlers with different settings
- Easy integration into existing projects

## Backward compatibility

- All existing functions work without changes
- Old code continues to work
- New features are fully optional

## Usage in other projects

```bash
go get github.com/vladislavsavi/ipxpress/pkg/ipxpress
```

```go
import "github.com/vladislavsavi/ipxpress/pkg/ipxpress"

handler := ipxpress.NewHandler(nil)
// Mount in your router
```

## Testing

```bash
# Build
go build ./...

# Run example
go run examples/library_usage/main.go

# Test API
curl "http://localhost:8080/img/?url=https://example.com/image.jpg&w=800"
```

## Next steps

1. Add `s-maxage` and `ETag` settings to CLI flags.
2. Add metrics (profiling, p50/p95 processing time).
3. Configurable cache store (for example, Redis) for horizontal scaling.
4. Document quality and performance recommendations.
