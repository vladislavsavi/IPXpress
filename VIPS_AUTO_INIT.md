# Automatic libvips Initialization

## Changes

IPXpress now initializes libvips automatically on first use. Users no longer need to call `vips.Startup()` and `vips.Shutdown()` manually.

## What changed

### 1. Automatic initialization

The `initVips()` function was added using `sync.Once`, and is called automatically when:
- Creating a new handler via `NewHandler()`
- Creating a new processor via `New()`

### 2. Default configuration

Automatic initialization uses the following settings:
```go
vips.Config{
    ConcurrencyLevel: 0,    // Use all available CPU cores
    MaxCacheMem:      2048, // 2GB cache memory
    MaxCacheSize:     5000, // Up to 5000 files in cache
}
```

### 3. Custom configuration (optional)

For production environments with high load, use `InitVipsWithConfig()`:

```go
ipxpress.InitVipsWithConfig(&vips.Config{
    ConcurrencyLevel: 0,
    MaxCacheMem:      4096,
    MaxCacheSize:     10000,
    MaxCacheFiles:    0,
}, vips.LogLevelWarning)
```

**Important:** Call this **before** creating any handlers or processors.

## Usage examples

### Simple usage (recommended)

```go
package main

import (
    "net/http"
    "github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func main() {
    // vips initializes automatically
    handler := ipxpress.NewHandler(nil)
    http.Handle("/img/", http.StripPrefix("/img/", handler))
    http.ListenAndServe(":8080", nil)
}
```

### With custom vips settings

```go
package main

import (
    "net/http"
    "github.com/davidbyttow/govips/v2/vips"
    "github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func main() {
    // Optional: configure vips before use
    ipxpress.InitVipsWithConfig(&vips.Config{
        ConcurrencyLevel: 0,
        MaxCacheMem:      4096,
        MaxCacheSize:     10000,
    }, vips.LogLevelWarning)

    handler := ipxpress.NewHandler(nil)
    http.Handle("/img/", http.StripPrefix("/img/", handler))
    http.ListenAndServe(":8080", nil)
}
```

### Direct image processing

```go
func processImage(data []byte) ([]byte, error) {
    // vips initializes automatically on first call
    proc := ipxpress.New().
        FromBytes(data).
        Resize(800, 600)

    if err := proc.Err(); err != nil {
        return nil, err
    }

    result, err := proc.ToBytes(ipxpress.FormatJpeg, 85)
    proc.Close()

    return result, err
}
```

## Backward compatibility

These changes are fully backward compatible. If you explicitly call `vips.Startup()` in your code, it will keep working. The automatic initialization detects that vips is already running and will not initialize it again.

## Updated files

1. `pkg/ipxpress/ipxpress.go` - added automatic initialization
2. `pkg/ipxpress/server.go` - `NewHandler` now calls `initVips()`
3. `cmd/ipxpress/main.go` - updated to use `InitVipsWithConfig()`
4. `examples/library_usage/main.go` - removed manual vips init
5. `pkg/ipxpress/ipxpress_test.go` - removed manual init from tests
6. Documentation (README.md, LIBRARY_USAGE.md, etc.) - updated

## Benefits

- Easier for new users - no need to learn libvips init details
- Less boilerplate code
- Impossible to forget vips initialization
- Safe init via `sync.Once`
- Custom configuration for production

## Requirements

libvips must still be installed on the system:

**Ubuntu/Debian:**
```bash
apt-get install libvips-dev
```

**macOS:**
```bash
brew install vips
```

See [govips prerequisites](https://github.com/davidbyttow/govips#prerequisites) for other platforms.
