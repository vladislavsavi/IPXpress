# IPXpress

IPXpress is a high-performance image processing service in Go with support for multiple formats and full access to libvips features.

## Features

- Load images from URLs (HTTP/HTTPS)
- Resize while preserving aspect ratio (Lanczos filter)
- Supported formats: **JPEG, PNG, GIF, WebP, AVIF**
- Compression quality control (1-100)
- Full access to any libvips function via `ImageRef()`, `ApplyFunc()`, and `VipsOperationBuilder`
- REST API service
- Built-in caching
- Extensible architecture with processors and middleware

## Supported Formats

| Format | Decoding | Encoding | Quality |
|--------|---|---|---|
| JPEG | Yes | Yes | Yes |
| PNG | Yes | Yes | No |
| GIF | Yes | Yes | No |
| WebP | Yes | Yes | Yes |
| AVIF | Yes | Yes | Yes |

## Project Structure

```
.
├── cmd/
│   └── ipxpress/          # HTTP server
├── pkg/ipxpress/          # Main library
│   ├── cache.go           # Caching system
│   ├── config.go          # Configuration
│   ├── extensions.go      # libvips extensions (new)
│   ├── fetcher.go         # Image fetching
│   ├── format.go          # Image formats
│   ├── ipxpress.go        # Image Processor
│   ├── params.go          # Request parameters
│   ├── server.go          # HTTP handler
│   └── *_test.go          # Tests
├── ARCHITECTURE.md        # Project architecture
├── API.md                 # API documentation
├── CUSTOM_OPERATIONS.md   # libvips usage (new)
├── CONTRIBUTING.md        # Developer guide
└── README.md              # This file
```

## Quick Start

### Build the server

```bash
go build ./cmd/ipxpress-server
```

### Run the server

```bash
./ipxpress-server -addr :8080
```

The server will be available at `http://localhost:8080/ipx/`

### Request Examples

#### Basic resize request

```bash
# Short parameters (compatible with ipx v2)
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=800&h=600"

# Or using s (resize)
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&s=800x600"
```

#### With quality control

```bash
# Short form: f=format, q=quality
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=1000&h=500&q=85&f=jpeg"

# Long form
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&width=1000&height=500&quality=85&format=jpeg"
```

#### WebP output

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&s=1000x500&q=100&f=webp" -o result.webp
```

#### AVIF output (modern format with better compression)

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=1200&f=avif&q=80" -o result.avif
```

#### Apply blur

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&blur=5.0" -o blurred.jpg
```

#### Sharpen

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&sharpen=1.5_1_2" -o sharp.jpg
```

#### Rotate 90 degrees

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&rotate=90" -o rotated.jpg
```

#### Flip/flop

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&flip=true" -o flipped.jpg
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&flop=true" -o flopped.jpg
```

#### Grayscale

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&grayscale=true" -o grayscale.jpg
```

#### Crop (extract area)

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&extract=100_100_400_400" -o cropped.jpg
```

#### Combine effects

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=800&grayscale=true&sharpen=1.0&quality=90&format=webp" -o processed.webp
```

#### Choose resampling kernel

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=200&kernel=lanczos3" -o resized.jpg
```

#### Allow upscale

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/small.jpg&w=2000&enlarge=true" -o enlarged.jpg
```

## API Parameters

**Core parameters** (compatible with [ipx v2](https://github.com/unjs/ipx)):

| Parameter | Short | Description | Type | Required |
|----------|----------|---------|-----|---|
| `url` | - | Image URL | string | Yes |
| `width` | `w` | Maximum width in pixels | int | No |
| `height` | `h` | Maximum height in pixels | int | No |
| `resize` | `s` | Size in WIDTHxHEIGHT format | string | No |
| `quality` | `q` | Compression quality (1-100) | int | No |
| `format` | `f` | Output format (jpeg, png, gif, webp, avif) | string | No |
| `background` | `b` | Background color (hex without #) | string | No |
| `position` | `pos` | Crop position | string | No |

### Caching and headers

- Internal cache: in-memory, TTL is controlled by `Config.CacheTTL` (default 30s).
- HTTP caching:
	- `Cache-Control`: configured via `Config.ClientMaxAge` and `Config.SMaxAge`.
	- `ETag`: enabled by default (`Config.EnableETag=true`). `If-None-Match` matches return `304`.

Example configuration (as a library):

```go
cfg := ipxpress.NewDefaultConfig()
cfg.ClientMaxAge = 3600 // 1 hour for clients
cfg.SMaxAge = 3600      // 1 hour for CDN/shared cache
cfg.CacheTTL = 10 * time.Minute
cfg.EnableETag = true
handler := ipxpress.NewHandler(cfg)
```

### Resize parameters

| Parameter | Description | Examples |
|----------|---------|---------|
| `fit` | Fit mode | contain, cover, fill, inside, outside |
| `position` / `pos` | Crop position | center, top, bottom, left, right, entropy, attention |
| `kernel` | Resampling algorithm | nearest, cubic, mitchell, lanczos2, lanczos3 |
| `enlarge` | Allow upscaling | true, false |

### Processing operations

| Parameter | Description | Value format |
|----------|---------|-----------------|
| `blur` | Gaussian blur | sigma (float, for example 5.0) |
| `sharpen` | Sharpen | sigma_flat_jagged (for example "1.5_1_2") |
| `rotate` | Image rotation | 0, 90, 180, 270 (degrees) |
| `flip` | Vertical flip | true |
| `flop` | Horizontal flip | true |
| `grayscale` | Convert to grayscale | true |

### Crop and extend

| Parameter | Description | Value format |
|----------|---------|-----------------|
| `extract` | Extract area | left_top_width_height (for example "10_10_200_200") |
| `extend` | Add borders | top_right_bottom_left (for example "10_10_10_10") |

### Color operations

| Parameter | Description | Value format |
|----------|---------|-----------------|
| `background` | Background color | hex without # (for example "ffffff" or "fff") |
| `negate` | Invert colors | true |
| `normalize` | Normalize | true |
| `gamma` | Gamma correction | float (for example 2.2) |
| `modulate` | HSB modulation | brightness_saturation_hue (for example "1.2_0.8_90") |
| `flatten` | Remove alpha channel | true |

**Resize behavior:**
- If only width (`w`) is set, height scales proportionally
- If only height (`h`) is set, width scales proportionally
- If both are set, the image scales to the largest size that fits the rectangle

## Documentation

- **[API.md](API.md)** - Full API documentation with examples
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Architecture and internals
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Developer guide

## Using as a library

```go
package main

import (
	"github.com/deadpixel/ipxpress/pkg/ipxpress"
)

func main() {
	// Load an image from bytes
	proc := ipxpress.New().
		FromBytes(imageBytes).
		Resize(800, 600)
	
	if err := proc.Err(); err != nil {
		panic(err)
	}
	
	// Encode to WebP with quality 85
	output, err := proc.ToBytes("webp", 85)
	if err != nil {
		panic(err)
	}
	
	// Use output...
}
```

## Tests

```bash
go test ./pkg/ipxpress
```

## Dependencies

- `github.com/davidbyttow/govips/v2` - Go bindings for libvips (image processing with native support for JPEG, PNG, GIF, WebP, AVIF)

**Note:** libvips must be installed. See [installation instructions](https://github.com/davidbyttow/govips#prerequisites).

The library automatically initializes libvips on first use, so you do not need to call `vips.Startup()` or `vips.Shutdown()` manually.

## License

MIT

