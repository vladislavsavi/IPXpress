# IPXpress Developer Guide

## Contents

1. [Getting Started](#getting-started)
2. [Code Structure](#code-structure)
3. [Developing New Features](#developing-new-features)
4. [Testing](#testing)
5. [Performance Optimization](#performance-optimization)
6. [Debugging](#debugging)
7. [Deployment](#deployment)

## Getting Started

### Requirements

- Go 1.24+
- libvips 8.15+
- Git

### Install libvips

#### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install libvips-dev
```

#### macOS

```bash
brew install vips
```

### Clone and run

```bash
# Clone repository
git clone https://github.com/vladislavsavi/IPXpress.git
cd IPXpress

# Install dependencies
go mod download

# Run tests
go test ./pkg/ipxpress/... -v

# Build
go build ./cmd/ipxpress

# Run
./ipxpress -addr :8080
```

## Code Structure

### Core components

```
pkg/ipxpress/
├── cache.go        # Caching system (interface + implementation)
├── config.go       # Service configuration
├── fetcher.go      # HTTP client for image fetching
├── format.go       # Image formats
├── ipxpress.go     # Processing core (Processor)
├── params.go       # HTTP request parameter parsing
├── server.go       # HTTP handler
└── *_test.go       # Tests
```

### Code organization principles

1. **Separation of concerns:** each file owns a specific area
2. **Interfaces:** used for abstractions (Cache, etc.)
3. **Chainable API:** Processor uses a fluent interface
4. **Immutability:** configuration is created once
5. **Explicit resource management:** use Close() to free memory

## Developing New Features

### Adding a new image format

Example: add AVIF support.

#### 1. Update `format.go`

```go
const (
    // ... existing formats
    FormatAVIF Format = "avif"
)

func (f Format) ContentType() string {
    switch f {
    // ... existing cases
    case FormatAVIF:
        return "image/avif"
    // ...
    }
}

func (f Format) IsValid() bool {
    switch f {
    // ... existing cases
    case FormatAVIF:
        return true
    // ...
    }
}

func DetectFormat(data []byte) Format {
    // ... existing checks

    // AVIF: check ftypavif
    if len(data) >= 12 &&
       data[4] == 0x66 && data[5] == 0x74 &&
       data[6] == 0x79 && data[7] == 0x70 {
        return FormatAVIF
    }

    return ""
}
```

#### 2. Update `ipxpress.go`

```go
func (p *Processor) ToBytes(format Format, quality int) ([]byte, error) {
    // ... existing checks

    switch format {
    // ... existing cases

    case FormatAVIF:
        // Check AVIF support in libvips
        params := vips.NewAvifExportParams()
        params.Quality = quality
        params.Speed = 5 // 0-9, lower is higher quality
        buf, _, err := p.img.ExportAvif(params)
        if err != nil {
            return nil, fmt.Errorf("failed to encode AVIF: %w", err)
        }
        return buf, nil

    // ...
    }
}
```

#### 3. Add tests

```go
// ipxpress_test.go
func TestProcessorAVIF(t *testing.T) {
    // Create a test image
    img := createTestImage(100, 50)

    proc := New().FromBytes(img).Resize(50, 0)

    out, err := proc.ToBytes(FormatAVIF, 85)
    if err != nil {
        t.Fatalf("encode AVIF: %v", err)
    }

    // Verify format
    format := DetectFormat(out)
    if format != FormatAVIF {
        t.Errorf("expected AVIF, got %s", format)
    }
}
```

### Adding a new transformation

Example: add a crop operation.

#### 1. Add a method to `Processor`

```go
// ipxpress.go

// Crop crops the image to the specified rectangle.
func (p *Processor) Crop(x, y, width, height int) *Processor {
    if p.err != nil {
        return p
    }
    if p.img == nil {
        p.err = errors.New("no image loaded")
        return p
    }

    // Validation
    if x < 0 || y < 0 || width <= 0 || height <= 0 {
        p.err = errors.New("invalid crop dimensions")
        return p
    }

    // Crop via libvips
    if err := p.img.ExtractArea(x, y, width, height); err != nil {
        p.err = fmt.Errorf("failed to crop: %w", err)
        return p
    }

    return p
}
```

#### 2. Update parameters

```go
// params.go

type ProcessingParams struct {
    URL     string
    Width   int
    Height  int
    Quality int
    Format  Format
    // New crop parameters
    CropX      int
    CropY      int
    CropWidth  int
    CropHeight int
}

func ParseProcessingParams(r *http.Request) *ProcessingParams {
    q := r.URL.Query()

    params := &ProcessingParams{
        // ... existing fields
        CropX:      parseInt(q.Get("crop_x")),
        CropY:      parseInt(q.Get("crop_y")),
        CropWidth:  parseInt(q.Get("crop_w")),
        CropHeight: parseInt(q.Get("crop_h")),
    }

    return params
}
```

#### 3. Use it in server.go

```go
// server.go

func (h *Handler) processImage(imageData []byte, params *ProcessingParams) *CacheEntry {
    proc := New().FromBytes(imageData)

    // Apply crop if specified
    if params.CropWidth > 0 && params.CropHeight > 0 {
        proc = proc.Crop(params.CropX, params.CropY,
                        params.CropWidth, params.CropHeight)
    }

    // Apply resize
    if params.Width > 0 || params.Height > 0 {
        proc = proc.Resize(params.Width, params.Height)
    }

    // ... remaining logic
}
```

### Adding middleware

Example: add request logging.

```go
// middleware.go (new file)

package ipxpress

import (
    "log"
    "net/http"
    "time"
)

// LoggingMiddleware logs all HTTP requests.
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrapper to capture status code
        wrapped := &statusWriter{ResponseWriter: w, status: 200}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start)
        log.Printf("[%s] %s %s - %d (%v)",
            r.Method, r.URL.Path, r.RemoteAddr,
            wrapped.status, duration)
    })
}

type statusWriter struct {
    http.ResponseWriter
    status int
}

func (w *statusWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}

// MetricsMiddleware collects performance metrics.
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Increment request counter
        // Measure latency
        // Report to monitoring system

        next.ServeHTTP(w, r)
    })
}
```

Usage in `main.go`:

```go
handler := ipxpress.NewHandler(config)

// Wrap with middleware
wrappedHandler := ipxpress.LoggingMiddleware(handler)
wrappedHandler = ipxpress.MetricsMiddleware(wrappedHandler)

mux.Handle("/ipx/", http.StripPrefix("/ipx/", wrappedHandler))
```

## Testing

### Running tests

```bash
# All tests
go test ./pkg/ipxpress/... -v

# With coverage
go test ./pkg/ipxpress/... -cover -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out
```

### Writing tests

#### Unit test for Processor

```go
func TestProcessorResize(t *testing.T) {
    // Create a test image
    img := createTestRGBA(200, 100)

    proc := New().FromBytes(img)

    // Apply resize
    proc = proc.Resize(100, 0)

    // Ensure no errors
    if err := proc.Err(); err != nil {
        t.Fatalf("resize failed: %v", err)
    }

    // Encode and verify size
    out, err := proc.ToBytes(FormatPNG, 85)
    proc.Close()

    if err != nil {
        t.Fatalf("encode failed: %v", err)
    }

    // Decode result
    decoded, _, err := image.Decode(bytes.NewReader(out))
    if err != nil {
        t.Fatalf("decode result: %v", err)
    }

    bounds := decoded.Bounds()
    if bounds.Dx() != 100 || bounds.Dy() != 50 {
        t.Errorf("unexpected size: %dx%d", bounds.Dx(), bounds.Dy())
    }
}
```

#### Integration test

```go
func TestServerIntegration(t *testing.T) {
    // Create a test image server
    imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Return a test image
        w.Header().Set("Content-Type", "image/jpeg")
        jpeg.Encode(w, createTestImage(1000, 500), &jpeg.Options{Quality: 90})
    }))
    defer imgServer.Close()

    // Create IPXpress handler
    handler := NewHandler(DefaultConfig())
    server := httptest.NewServer(handler)
    defer server.Close()

    // Make a request
    resp, err := http.Get(server.URL + "?url=" +
        url.QueryEscape(imgServer.URL+"/test.jpg") + "&w=500")

    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()

    // Check status
    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected 200, got %d", resp.StatusCode)
    }

    // Check Content-Type
    ct := resp.Header.Get("Content-Type")
    if !strings.HasPrefix(ct, "image/") {
        t.Errorf("invalid content-type: %s", ct)
    }
}
```

#### Benchmark tests

```go
func BenchmarkProcessorResize(b *testing.B) {
    img := createTestRGBA(2000, 1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        proc := New().FromBytes(img).Resize(800, 400)
        out, _ := proc.ToBytes(FormatJPEG, 85)
        proc.Close()
        _ = out
    }
}

func BenchmarkServerHandler(b *testing.B) {
    handler := NewHandler(DefaultConfig())

    // Create a test request
    req := httptest.NewRequest("GET", "/ipx/?url=...&w=800", nil)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
    }
}
```

## Performance Optimization

### Profiling

#### CPU profiling

```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

In interactive mode:
```
(pprof) top10      # Top 10 functions
(pprof) list ProcessImage  # Function details
(pprof) web        # Visualization
```

#### Memory profiling

```bash
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### Optimization tips

1. **Use pprof during load testing:**

```go
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // ... main code
}
```

2. **Minimize allocations:**
   - Reuse buffers
   - Avoid string concatenation in hot paths
   - Use sync.Pool for temporary objects

3. **libvips optimization:**
   - Tune ConcurrencyLevel to your hardware
   - Experiment with MaxCacheMem
   - Use vector operations (automatic)

## Debugging

### Logging

#### Enable VIPS logs

```go
// main.go
vips.LoggingSettings(func(domain string, level vips.LogLevel, msg string) {
    log.Printf("[%s:%s] %s", domain, level, msg)
}, vips.LogLevelInfo)
```

#### Structured logging

```go
import "github.com/rs/zerolog/log"

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Info().
        Str("url", r.URL.String()).
        Str("remote", r.RemoteAddr).
        Msg("processing request")

    // ... processing

    log.Info().
        Str("url", r.URL.String()).
        Int("status", statusCode).
        Dur("duration", duration).
        Msg("request completed")
}
```

### Debugging with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
dlv debug ./cmd/ipxpress

# In delve:
(dlv) break server.go:123
(dlv) continue
```

## Deployment

### Docker

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache vips-dev build-base

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . .
RUN go build -o ipxpress ./cmd/ipxpress

FROM alpine:latest
RUN apk add --no-cache vips

COPY --from=builder /app/ipxpress /usr/local/bin/

EXPOSE 8080
CMD ["ipxpress", "-addr", ":8080"]
```

Build and run:

```bash
docker build -t ipxpress:latest .
docker run -p 8080:8080 ipxpress:latest
```

### Kubernetes

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ipxpress
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ipxpress
  template:
    metadata:
      labels:
        app: ipxpress
    spec:
      containers:
      - name: ipxpress
        image: ipxpress:latest
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: ipxpress-service
spec:
  selector:
    app: ipxpress
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### Systemd (Linux)

```ini
# /etc/systemd/system/ipxpress.service
[Unit]
Description=IPXpress Image Processing Service
After=network.target

[Service]
Type=simple
User=ipxpress
WorkingDirectory=/opt/ipxpress
ExecStart=/opt/ipxpress/ipxpress -addr :8080
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Manage the service:

```bash
sudo systemctl enable ipxpress
sudo systemctl start ipxpress
sudo systemctl status ipxpress
```

## Best Practices

1. **Always call Close()** after using Processor
2. **Use context.Context** to cancel long-running operations
3. **Validate inputs** before processing
4. **Log errors** with enough context
5. **Write tests** for new features
6. **Document APIs** in code comments
7. **Use semantic versioning** (SemVer)

## Additional resources

- [libvips documentation](https://www.libvips.org/API/current/)
- [govips examples](https://github.com/davidbyttow/govips)
- [Go testing best practices](https://go.dev/doc/tutorial/add-a-test)
