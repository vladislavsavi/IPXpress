# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build server
go build ./cmd/ipxpress

# Run server
./ipxpress -addr :8080

# Run all tests
go test ./test/ipxpress/... -v

# Run single test
go test ./test/ipxpress/... -run TestName -v

# Coverage
go test ./test/ipxpress/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Benchmarks
go test ./test/ipxpress/... -bench=. -benchmem

# Docker
docker build -t ipxpress .
docker run -p 8080:8080 ipxpress
```

**Prerequisite:** libvips 8.15+ must be installed (`brew install vips` on macOS, `apt-get install libvips-dev` on Ubuntu).

## Architecture

IPXpress is both a standalone HTTP image server and an importable Go library.

### Package layout

- `cmd/ipxpress/` — HTTP server entry point; mounts handler at `/ipx/`, adds `/health`
- `pkg/ipxpress/` — library package (all core types live here)
- `test/ipxpress/` — external test package (`package ipxpress_test`); tests are here, not in `pkg/`
- `examples/library_usage/` — demonstrates library usage

### Request pipeline (`server.go`)

```
HTTP request
→ ParseProcessingParams (params.go)
→ cache lookup (Otter W-TinyLFU, cost-based by byte size)
→ singleflight.Do (deduplicates concurrent requests for same key)
  → semaphore acquire (ProcessingLimit=256, limits concurrent fetches+processing)
  → Fetcher.Fetch (fetcher.go) — HTTP client with connection pooling
  → processImage → applyBuiltInTransformations → custom ProcessorFuncs
  → cache.Set
  → semaphore release
→ writeResponse (ETag/304, Cache-Control headers)
```

Cache key: MD5 of all request parameters (url, width, height, quality, format, fit, blur, etc.).

### Processor (`ipxpress.go`)

Chainable fluent API wrapping `govips/v2`. Every method returns `*Processor`; errors are sticky — the chain short-circuits on first error.

```go
proc := ipxpress.New().FromBytes(data).Resize(800, 0).Blur(2.0)
out, err := proc.ToBytes(ipxpress.FormatWebP, 85)
proc.Close() // always required — frees libvips memory
```

libvips is initialized once via `sync.Once` on first `New()` call. libvips internal cache is disabled (`MaxCacheMem=0`, `MaxCacheSize=0`) — caching is handled at the application level.

### Extension points

**Custom processors** — `ProcessorFunc` applied after built-in transforms:
```go
handler.UseProcessor(func(p *Processor, params *ProcessingParams) *Processor {
    return p.ApplyFunc(func(img *vips.ImageRef) error { ... })
})
```

**Custom middleware** — `MiddlewareFunc` wraps the `http.Handler`:
```go
handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))
```

**Direct libvips access** — `proc.ImageRef()` returns the raw `*vips.ImageRef`, and `proc.ApplyFunc(fn)` / `VipsOperationBuilder` allow arbitrary libvips calls within the chain.

### Cache (`cache.go`)

`InMemoryCache` backed by [Otter](https://github.com/maypok86/otter) (W-TinyLFU). Cost = byte size of cached data. Default max cost = 512 MB, TTL = 10 minutes. Both successful responses and errors are cached.

### Config defaults

| Field | Default | Notes |
|---|---|---|
| `CacheTTL` | 10m | In-memory cache TTL |
| `CacheMaxCost` | 512 MB | Otter cost-based eviction |
| `ProcessingLimit` | 256 | Semaphore depth |
| `ClientMaxAge` | 604800s (7d) | `Cache-Control: max-age` |
| `EnableETag` | true | ETag precomputed at encode time |

### Built-in transformation order

Extract → Resize → Extend → Rotate → Flip/Flop → Blur → Sharpen → Grayscale/Negate/Normalize/Gamma/Modulate → Flatten

### Adding a new format

1. Add constant in `format.go`, update `ContentType()`, `IsValid()`, `DetectFormat()`
2. Add `case FormatXXX:` in `Processor.ToBytes()` (`ipxpress.go`)
3. Add tests in `test/ipxpress/`
