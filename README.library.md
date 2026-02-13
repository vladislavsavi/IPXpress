# IPXpress - Minimal, Extensible Image Processing Library

IPXpress is a fast and flexible image processing library for Go, built on libvips.

## Key features

- Minimal API - easy to use in any project
- Full extensibility - use any libvips function
- High performance - libvips-based processing
- Caching - built-in result caching
- Flexibility - use as a library or as a ready server
- Direct access - full ImageRef access for any operations

## Quick start

### As a library

```go
import "github.com/vladislavsavi/ipxpress/pkg/ipxpress"

func main() {
    // Simplest setup: default config
    handler := ipxpress.NewHandler(nil)
    http.Handle("/img/", http.StripPrefix("/img/", handler))
    http.ListenAndServe(":8080", nil)
}
```

### As a standalone server

```bash
go build -o ipxpress ./cmd/ipxpress
./ipxpress -addr :8080
```

## Usage

### Basic integration

```go
// Create a handler with default settings
handler := ipxpress.NewHandler(nil)

// Or with custom configuration
config := &ipxpress.Config{
    ProcessingLimit: 10,
    CacheTTL:        5 * time.Minute,
}
handler := ipxpress.NewHandler(config)

// Explicit way to get the default config
cfg := ipxpress.NewDefaultConfig()
handler2 := ipxpress.NewHandler(cfg)

// Mount in your router
http.Handle("/images/", http.StripPrefix("/images/", handler))
```

### Extending functionality

#### Adding custom processors

```go
handler := ipxpress.NewHandler(nil)

// Auto-orient images based on EXIF
handler.UseProcessor(ipxpress.AutoOrientProcessor())

// Strip metadata for privacy
handler.UseProcessor(ipxpress.StripMetadataProcessor())

// Optimize for web delivery
handler.UseProcessor(ipxpress.CompressionOptimizer())
```

#### Creating your own processor

```go
customProcessor := func(proc *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    // Your processing logic
    return proc.Sharpen(1.5, 1.0, 2.0)
}

handler.UseProcessor(customProcessor)
```

#### Adding middleware

```go
// CORS
handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))

// Authentication
handler.UseMiddleware(ipxpress.AuthMiddleware([]string{"secret-token"}))

// Logging
logger := func(format string, args ...interface{}) {
    log.Printf(format, args...)
}
handler.UseMiddleware(ipxpress.LoggingMiddleware(logger))

// Custom middleware
customMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your logic
        next.ServeHTTP(w, r)
    })
}
handler.UseMiddleware(customMiddleware)
```

### Multiple handlers

```go
// Public handler with limits
publicHandler := ipxpress.NewHandler(&ipxpress.Config{
    ProcessingLimit: 5,
})
publicHandler.UseMiddleware(ipxpress.RateLimitMiddleware(100))

// Private handler with auth
privateHandler := ipxpress.NewHandler(&ipxpress.Config{
    ProcessingLimit: 20,
})
privateHandler.UseMiddleware(ipxpress.AuthMiddleware([]string{"admin-token"}))

// Mount both
http.Handle("/public/img/", http.StripPrefix("/public/img/", publicHandler))
http.Handle("/private/img/", http.StripPrefix("/private/img/", privateHandler))
```

## API

```
GET /ipx/?url=https://example.com/image.jpg&w=800&h=600&quality=85&format=webp
```

**Parameters:**
- `url` (required) - image URL to process
- `w` - max width
- `h` - max height
- `quality` - quality (1-100, default 85)
- `format` - output format (jpeg, png, gif, webp)

Full API documentation: [API.md](API.md)

## Direct image processing

You can use IPXpress without HTTP:

```go
proc := ipxpress.New().
    FromBytes(imageData).
    Resize(800, 600).
    Sharpen(1.0, 1.0, 2.0)

output, err := proc.ToBytes(ipxpress.FormatJPEG, 85)
proc.Close()
```

### Using any libvips functions

IPXpress provides full access to any libvips function through several mechanisms:

#### 1. Direct ImageRef access

```go
proc := ipxpress.New().FromBytes(imageData)

// Get direct access to vips.ImageRef
img := proc.ImageRef()
if img != nil {
    img.Blur(2.0)
    img.Sharpen(1.5, 0.5, 1.0)
    img.Modulate(1.1, 1.2, 0)
}

output, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
```

#### 2. ApplyFunc for custom operations

```go
proc := ipxpress.New().
    FromBytes(imageData).
    ApplyFunc(func(img *vips.ImageRef) error {
        if err := img.Blur(2.0); err != nil {
            return err
        }
        return img.Sharpen(1.5, 0.5, 1.0)
    })

output, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
```

#### 3. VipsOperationBuilder for operation chains

```go
proc := ipxpress.New().FromBytes(imageData)

builder := ipxpress.NewVipsOperationBuilder(proc)
err := builder.
    Blur(2.0).
    Sharpen(1.5, 0.5, 1.0).
    Modulate(1.1, 1.2, 0).
    Error()

output, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
```

Details: [CUSTOM_OPERATIONS.md](CUSTOM_OPERATIONS.md)

## Documentation

- [LIBRARY_USAGE.md](LIBRARY_USAGE.md) - Detailed library usage
- [API.md](API.md) - Full API description
- [ARCHITECTURE.md](ARCHITECTURE.md) - Project architecture
- [CUSTOM_OPERATIONS.md](CUSTOM_OPERATIONS.md) - Extending and using any libvips functions

## Usage examples

### Integration with Chi Router

```go
r := chi.NewRouter()
r.Get("/", homeHandler)

imgHandler := ipxpress.NewHandler(nil)
r.Mount("/img", imgHandler)
```

### Integration with Gorilla Mux

```go
r := mux.NewRouter()
r.HandleFunc("/", homeHandler)

imgHandler := ipxpress.NewHandler(nil)
r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", imgHandler))
```

### Production setup

```go
// The library initializes vips automatically
// You do not need to call vips.Startup() or vips.Shutdown()

// Basic setup
handler := ipxpress.NewHandler(nil)

// Custom settings if needed
config := ipxpress.NewDefaultConfig()
config.ProcessingLimit = 10
config.CacheTTL = 30 * time.Minute

handler = ipxpress.NewHandler(config)
handler.UseProcessor(ipxpress.AutoOrientProcessor())
handler.UseProcessor(ipxpress.CompressionOptimizer())
handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))
```

## Requirements

- Go 1.21+
- libvips 8.12+

**Note:** The library automatically initializes libvips on first use. You do not need to call `vips.Startup()` or `vips.Shutdown()` manually.

## Installing libvips

### Ubuntu/Debian
```bash
apt-get install libvips-dev
```

### macOS
```bash
brew install vips
```

## License

MIT License - see [LICENSE](LICENSE)

