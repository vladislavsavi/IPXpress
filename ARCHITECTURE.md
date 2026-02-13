# IPXpress Architecture

## Overview

IPXpress is a high-performance HTTP image processing service in Go built on libvips. The project is designed for scalability, performance, and code readability.

## Project Structure

```
IPXpress/
├── cmd/
│   └── ipxpress/           # Application entry point
│       └── main.go         # HTTP server with libvips init
├── pkg/
│   └── ipxpress/           # Main library package
│       ├── cache.go        # Caching system
│       ├── config.go       # Service configuration
│       ├── fetcher.go      # Image fetching by URL
│       ├── format.go       # Image formats
│       ├── ipxpress.go     # Image processing core (Processor)
│       ├── params.go       # Request parameter parsing
│       ├── server.go       # HTTP request handler
│       ├── *_test.go       # Tests
│       └── ...
├── static/                 # Static files (if any)
├── go.mod                  # Go module
├── Dockerfile              # Docker image
└── README.md               # Project documentation
```

## System Components

### 1. **Processor** (`ipxpress.go`)

Core image processing pipeline. Uses a chainable API to apply transformations in sequence.

**Key capabilities:**
- Load images from bytes or io.Reader
- Resize with aspect ratio preserved (Lanczos3)
- Encode to various formats (JPEG, PNG, GIF, WebP)
- Auto-detect source image format
- Memory management via libvips

**Example:**
```go
proc := ipxpress.New().
    FromBytes(imageData).
    Resize(800, 600)

if err := proc.Err(); err != nil {
    // handle error
}

output, err := proc.ToBytes(ipxpress.FormatJPEG, 85)
proc.Close() // free memory
```

### 2. **Format** (`format.go`)

Image format utilities.

**Capabilities:**
- Typed format constants (FormatJPEG, FormatPNG, FormatGIF, FormatWebP)
- Auto-detect format by magic bytes
- MIME type lookup for HTTP headers
- Format validation

**Examples:**
```go
// Detect format
format := ipxpress.DetectFormat(imageData)

// Parse format string
format := ipxpress.ParseFormat("jpeg") // returns FormatJPEG

// MIME type
contentType := format.ContentType() // "image/jpeg"
```

### 3. **Cache** (`cache.go`)

Response cache with TTL (Time To Live).

**Architecture:**
- `Cache` interface for multiple implementations
- `InMemoryCache` with sync.RWMutex
- Automatic cleanup of expired entries
- Caches both successful responses and errors

**Entry structure:**
```go
type CacheEntry struct {
    ContentType string    // Response MIME type
    Data        []byte    // Image data
    StatusCode  int       // HTTP status
    ErrorMsg    string    // Error message (if any)
    Timestamp   time.Time // Entry creation time
}
```

### 4. **Fetcher** (`fetcher.go`)

Image download module.

**Capabilities:**
- HTTP/HTTPS support
- Connection pooling for high performance
- URL validation
- Configurable timeouts
- User-Agent for basic restrictions

**HTTP client configuration:**
```go
- Timeout: 20 seconds
- MaxIdleConns: 500
- MaxIdleConnsPerHost: 100
- MaxConnsPerHost: 256
- DialTimeout: 5 seconds
- KeepAlive: 30 seconds
```

### 5. **Params** (`params.go`)

HTTP request parameter parsing and validation.

**Structure:**
```go
type ProcessingParams struct {
    URL     string  // Image URL
    Width   int     // Max width
    Height  int     // Max height
    Quality int     // Quality (1-100)
    Format  Format  // Output format
}
```

**Logic:**
- Automatic parameter validation
- Default quality: 85
- Decide if processing is required
- Choose output format (original or explicit)

### 6. **Server** (`server.go`)

HTTP handler for service requests.

**Handler structure:**
```go
type Handler struct {
    cache           Cache           // Cache system
    fetcher         *Fetcher        // Image fetcher
    config          *Config         // Configuration
    processingLimit chan struct{}   // Semaphore to limit concurrency
}
```

**Request flow:**
1. Parse request parameters
2. Cache lookup (fast path)
3. Fetch image (parallel, I/O bound)
4. Process image (concurrency limited, CPU bound)
5. Cache result
6. Write response

**Optimizations:**
- Two-stage processing: I/O first (parallel), CPU with a semaphore
- Cache errors to avoid repeated fetches
- Connection pooling for outgoing HTTP requests
- Optimized HTTP headers (Cache-Control)

### 7. **Config** (`config.go`)

Service configuration.

```go
type Config struct {
    CacheTTL        time.Duration // Cache TTL (30 seconds)
    ProcessingLimit int           // Max concurrent processing (256)
    CleanupInterval time.Duration // Cache cleanup interval (30 seconds)
}
```

## Data Flow

### Request processing

```
HTTP Request
    |
[ParseParams] -> ProcessingParams
    |
[Cache Check] -> Hit? -> [Write Response]
    | Miss
[Fetch Image] -> imageData (parallel, no semaphore)
    |
[Acquire Semaphore] -> limit concurrent processing
    |
[Process Image] -> Processor chain
    |
[Encode Output] -> output bytes
    |
[Release Semaphore]
    |
[Cache Result]
    |
[Write Response]
```

### Image processing (Processor)

```
Image Bytes
    |
[Detect Format] -> Format
    |
[Decode with libvips] -> vips.ImageRef
    |
[Resize (optional)] -> transformed ImageRef
    |
[Encode to Format] -> output bytes
    |
[Close/Free Memory]
    |
Output Bytes
```

## Concurrency and Performance

### Processing strategy

1. **I/O phase (no limits):**
   - Fetch images via HTTP
   - Cache lookup
   - Parallel handling of many requests

2. **CPU phase (with semaphore):**
   - Process images via libvips
   - Limit: 256 concurrent operations
   - Avoid memory pressure

### Caching

- **Cache key:** MD5(url|width|height|quality|format)
- **TTL:** 30 seconds (configurable)
- **Cleanup:** periodic (every 30 seconds)
- **Storage:** in-memory (fast access)

### Memory management

- Immediate release after processing (`proc.Close()`)
- libvips settings:
  - MaxCacheMem: 0 MB (application-level caching)
  - MaxCacheSize: 0 images
  - ConcurrencyLevel: 0 (use all CPU cores)

## API

### HTTP endpoint

**URL:** `/ipx/`

**Query parameters:**

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `url` | string | Yes | Image URL (HTTP/HTTPS) |
| `w` | int | No | Max width in pixels |
| `h` | int | No | Max height in pixels |
| `quality` | int | No | Compression quality (1-100, default: 85) |
| `format` | string | No | Output format (jpeg, png, gif, webp) |

**Examples:**

```bash
# Resize
GET /ipx/?url=https://example.com/image.jpg&w=800&h=600

# Convert to WebP
GET /ipx/?url=https://example.com/image.jpg&format=webp&quality=90

# Resize only (keep format)
GET /ipx/?url=https://example.com/image.png&w=500
```

## Extending the system

### Add a new format

1. Add a constant in `format.go`:
```go
const FormatAVIF Format = "avif"
```

2. Update `ContentType()`:
```go
case FormatAVIF:
    return "image/avif"
```

3. Add handling in `Processor.ToBytes()` (`ipxpress.go`)

### Replace the cache system

Implement the `Cache` interface:
```go
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Get(key string) (*CacheEntry, bool) { ... }
func (c *RedisCache) Set(key string, entry *CacheEntry) { ... }
func (c *RedisCache) Cleanup() { ... }
```

Use it in Handler:
```go
handler := &Handler{
    cache: NewRedisCache(...),
    // ...
}
```

## Testing

Run tests:
```bash
go test ./pkg/ipxpress/... -v
```

Coverage:
```bash
go test ./pkg/ipxpress/... -cover
```

## Monitoring and logging

### Current logs
- libvips logs (WARNING+ level)
- HTTP requests (standard log)

### Production recommendations
- Add structured logging (zap, zerolog)
- Prometheus metrics (latency, cache hit rate, error rate)
- Distributed tracing (OpenTelemetry)
- Health check endpoint (`/health`)

## Performance

### Targets
- **Throughput:** 3000+ req/sec
- **Latency:** <50ms (cached), <200ms (processed)
- **Concurrency:** 256 concurrent operations

### Optimizations
- Connection pooling (500 idle connections)
- Response caching (30 sec TTL)
- Efficient memory management (immediate cleanup)
- Vector operations in libvips (SIMD)

## Security

### Current measures
- URL validation (HTTP/HTTPS only)
- Timeouts for all operations
- Concurrency limit (DoS protection)

### Recommendations
- Rate limiting by IP
- Domain allowlist for URLs
- Maximum file size
- Authentication/authorization

## Deployment

### Docker
```bash
docker build -t ipxpress .
docker run -p 8080:8080 ipxpress
```

### Native build
```bash
go build -o ipxpress ./cmd/ipxpress
./ipxpress -addr :8080
```

## Dependencies

- **libvips:** Fast image processing library
- **govips:** Go bindings for libvips
- Go standard library

## License

See LICENSE in the project root.
