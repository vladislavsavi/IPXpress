# IPXpress

IPXpress is a high-performance image processing service written in Go with support for multiple formats.

## Features

- ✅ Load images from URLs (HTTP/HTTPS)
- ✅ Resize images while preserving aspect ratio (Lanczos filter)
- ✅ Support for formats: **JPEG, PNG, GIF, WebP**
- ✅ Compression quality control (1-100)
- ✅ REST API service
- ✅ Chainable API for library usage

## Supported Formats

| Format | Decoding | Encoding | Quality |
|--------|---|---|---|
| JPEG | ✅ | ✅ | ✅ |
| PNG | ✅ | ✅ | ❌ |
| GIF | ✅ | ✅ | ❌ |
| WebP | ✅ | ✅ | ✅ |

## Project Structure

```
.
├── cmd/
│   ├── ipxpress/          # CLI utility (future implementation)
│   └── ipxpress-server/   # REST API server
├── pkg/ipxpress/          # Main library
│   ├── ipxpress.go        # Image Processor
│   ├── server.go          # HTTP handler
│   └── server_test.go     # Tests
└── README.md
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

#### Basic request with resizing

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=800&h=600"
```

#### With quality control

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=1000&h=500&quality=85&format=jpeg"
```

#### In WebP format

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=1000&h=500&quality=100&format=webp" -o result.webp
```

#### In PNG format

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&format=png" -o result.png
```

## API Parameters

| Parameter | Description | Type | Required |
|-----------|-------------|------|---|
| `url` | Image URL | string | ✅ |
| `w` | Maximum width in pixels | int | ❌ |
| `h` | Maximum height in pixels | int | ❌ |
| `quality` | Compression quality (1-100) | int | ❌ |
| `format` | Output format (jpeg, png, gif, webp) | string | ❌ |

**Resize behavior:**
- If only width (`w`) is specified — height scales proportionally
- If only height (`h`) is specified — width scales proportionally
- If both are specified — image scales to the largest size that fits within the rectangle

## Using as a Library

```go
package main

import (
	"github.com/deadpixel/ipxpress/pkg/ipxpress"
)

func main() {
	// Load image from bytes
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

- `github.com/chai2010/webp` — WebP encoding/decoding
- `github.com/disintegration/imaging` — High-quality scaling (Lanczos)
- `golang.org/x/image` — JPEG, PNG, GIF support

## License

MIT
