# IPXpress Library Usage

IPXpress is a minimalist, extensible image processing library built on libvips. It can be easily integrated into your Go projects.

## Installation

```bash
go get github.com/vladislavsavi/ipxpress/pkg/ipxpress
```

## Basic Usage

### Simple Handler

The simplest way to use IPXpress is with the default handler:

```go
package main

import (
    "log"
    "net/http"
    "github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func main() {
    // Create a simple handler with default settings
    handler := ipxpress.NewHandler(nil)
    
    // Mount it
    http.Handle("/img/", http.StripPrefix("/img/", handler))
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Custom Configuration

```go
config := &ipxpress.Config{
    ProcessingLimit: 10,              // Max 10 concurrent image processing operations
    CacheTTL:        5 * time.Minute, // Cache images for 5 minutes
    CleanupInterval: 1 * time.Minute, // Clean cache every minute
}

handler := ipxpress.NewHandler(config)
```

## Extensibility

### Adding Custom Processors

Processors allow you to add custom image transformations:

```go
// Create a custom processor
watermarkProcessor := func(proc *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    // Add your custom logic here
    // For example, add a watermark to all images
    return proc.Sharpen(1.5, 1.0, 2.0)
}

handler := ipxpress.NewHandler(nil)
handler.UseProcessor(watermarkProcessor)
```

### Built-in Custom Processors

```go
handler := ipxpress.NewHandler(nil)

// Automatically rotate images based on EXIF orientation
handler.UseProcessor(ipxpress.AutoOrientProcessor())

// Strip all metadata for privacy
handler.UseProcessor(ipxpress.StripMetadataProcessor())

// Optimize compression settings
handler.UseProcessor(ipxpress.CompressionOptimizer())
```

### Adding Middleware

Middleware wraps the HTTP handler with additional functionality:

```go
handler := ipxpress.NewHandler(nil)

// Add CORS support
handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))

// Add authentication
validTokens := []string{"secret-token-1", "secret-token-2"}
handler.UseMiddleware(ipxpress.AuthMiddleware(validTokens))

// Add logging
logger := func(format string, args ...interface{}) {
    log.Printf(format, args...)
}
handler.UseMiddleware(ipxpress.LoggingMiddleware(logger))
```

### Custom Middleware Example

```go
// Create custom middleware
metricsMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
    })
}

handler.UseMiddleware(metricsMiddleware)
```

## Multiple Handlers in One Application

You can create multiple handlers with different configurations:

```go
func main() {
    // Public handler with rate limiting
    publicHandler := ipxpress.NewHandler(&ipxpress.Config{
        ProcessingLimit: 5,
        CacheTTL:        10 * time.Minute,
    })
    publicHandler.UseMiddleware(ipxpress.RateLimitMiddleware(100))
    
    // Private handler with authentication
    privateHandler := ipxpress.NewHandler(&ipxpress.Config{
        ProcessingLimit: 20,
        CacheTTL:        1 * time.Hour,
    })
    privateHandler.UseMiddleware(ipxpress.AuthMiddleware([]string{"admin-token"}))
    privateHandler.UseProcessor(ipxpress.StripMetadataProcessor())
    
    // Mount both handlers
    http.Handle("/public/img/", http.StripPrefix("/public/img/", publicHandler))
    http.Handle("/private/img/", http.StripPrefix("/private/img/", privateHandler))
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Integration with Existing Applications

### With existing http.ServeMux

```go
mux := http.NewServeMux()

// Your existing routes
mux.HandleFunc("/", homeHandler)
mux.HandleFunc("/api/users", usersHandler)

// Add IPXpress
imgHandler := ipxpress.NewHandler(nil)
mux.Handle("/images/", http.StripPrefix("/images/", imgHandler))

http.ListenAndServe(":8080", mux)
```

### With popular routers (chi, gorilla/mux, etc.)

```go
// Chi router example
r := chi.NewRouter()
r.Use(middleware.Logger)

// Your routes
r.Get("/", homeHandler)

// Mount IPXpress
imgHandler := ipxpress.NewHandler(nil)
r.Mount("/img", imgHandler)

http.ListenAndServe(":8080", r)
```

## Direct Image Processing (without HTTP)

You can also use IPXpress for direct image processing:

```go
import "github.com/vladislavsavi/ipxpress/pkg/ipxpress"

func processImage(inputData []byte) ([]byte, error) {
    proc := ipxpress.New().
        FromBytes(inputData).
        Resize(800, 600).
        Sharpen(1.0, 1.0, 2.0).
        Blur(1.5)
    
    if err := proc.Err(); err != nil {
        return nil, err
    }
    
    output, err := proc.ToBytes(ipxpress.FormatJpeg, 85)
    proc.Close()
    
    return output, err
}
```

## API Endpoints

Once integrated, your handler responds to:

```
GET /ipx/?url=https://example.com/image.jpg&w=800&h=600&quality=85&format=webp
```

Query parameters:
- `url` (required): Image URL to process
- `w`: Maximum width
- `h`: Maximum height
- `quality`: Output quality (1-100, default: 85)
- `format`: Output format (jpeg, png, gif, webp)

See [API.md](../../API.md) for complete API documentation.

## Performance Tips

1. **Set appropriate ProcessingLimit**: Match to your server's CPU cores
2. **Use caching**: Enable and configure cache TTL based on your use case
3. **Add custom processors wisely**: Each processor adds processing time
4. **Consider middleware order**: Place auth/rate-limiting before heavy processing

## Example: Complete Production Setup

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/davidbyttow/govips/v2/vips"
    "github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func main() {
    // Initialize vips
    vips.Startup(&vips.Config{
        ConcurrencyLevel: 0,
        MaxCacheMem:      2048,
        MaxCacheSize:     5000,
    })
    defer vips.Shutdown()
    
    // Configure IPXpress
    config := &ipxpress.Config{
        ProcessingLimit: 10,
        CacheTTL:        30 * time.Minute,
        CleanupInterval: 5 * time.Minute,
    }
    
    handler := ipxpress.NewHandler(config)
    
    // Add features
    handler.UseProcessor(ipxpress.AutoOrientProcessor())
    handler.UseProcessor(ipxpress.CompressionOptimizer())
    handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))
    
    // Setup server
    mux := http.NewServeMux()
    mux.Handle("/img/", http.StripPrefix("/img/", handler))
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```
