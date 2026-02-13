# IPXpress API Documentation

## Overview

IPXpress provides a REST API for real-time image processing. The service fetches images by URL, applies transformations, and returns the result.

## Base URL

```
http://localhost:8080/ipx/
```

## Endpoint

### GET /ipx/

Processes an image with the specified parameters.

#### Query parameters

**Core parameters:**

| Parameter | Short | Type | Required | Default | Description |
|----------|----------|-----|--------------|--------------|----------|
| `url` | - | string | **Yes** | - | Image URL to process (HTTP/HTTPS) |
| `width` | `w` | integer | No | - | Max width in pixels |
| `height` | `h` | integer | No | - | Max height in pixels |
| `resize` | `s` | string | No | - | Size in `WIDTHxHEIGHT` format (for example, `800x600`) |
| `quality` | `q` | integer | No | 85 | Compression quality for JPEG/WebP/AVIF (1-100) |
| `format` | `f` | string | No | original | Output format: `jpeg`, `png`, `gif`, `webp`, `avif` |

**Resize parameters:**

| Parameter | Short | Type | Default | Description |
|----------|----------|-----|--------------|----------|
| `fit` | - | string | - | Fit mode: `contain`, `cover`, `fill`, `inside`, `outside` |
| `position` | `pos` | string | - | Crop position: `top`, `bottom`, `left`, `right`, `centre`, `entropy`, `attention` |
| `kernel` | - | string | `lanczos3` | Resampling algorithm: `nearest`, `cubic`, `mitchell`, `lanczos2`, `lanczos3` |
| `enlarge` | - | boolean | `false` | Allow upscaling above original size |

**Crop and extend operations:**

| Parameter | Description | Example |
|----------|----------|--------|
| `extract` | Extract region: `left_top_width_height` | `extract=10_10_200_200` |
| `trim` | Trim edges by threshold | `trim=10` |
| `extend` | Add border: `top_right_bottom_left` | `extend=10_10_10_10` |
| `background` | `b` | Background color (hex) | `background=ff0000` or `b=ff0000` |

**Effects and filters:**

| Parameter | Description | Example |
|----------|----------|--------|
| `blur` | Blur (sigma) | `blur=5` |
| `sharpen` | Sharpen: `sigma_flat_jagged` | `sharpen=1.5_1_2` |
| `rotate` | Rotate in degrees (90/180/270) | `rotate=90` |
| `flip` | Vertical flip | `flip=true` |
| `flop` | Horizontal flip | `flop=true` |
| `grayscale` | Convert to grayscale | `grayscale=true` |
| `negate` | Invert colors | `negate=true` |
| `normalize` | Normalize | `normalize=true` |
| `gamma` | Gamma correction | `gamma=2.2` |
| `median` | Median filter | `median=3` |
| `threshold` | Threshold for binarization | `threshold=128` |
| `tint` | Tint (hex) | `tint=00ff00` |
| `modulate` | Modulate: `brightness_saturation_hue` | `modulate=1.2_0.8_90` |
| `flatten` | Remove transparency | `flatten=true` |

#### Response headers

- `Content-Type`: image MIME type (`image/jpeg`, `image/png`, etc.)
- `Content-Length`: size in bytes
- `Cache-Control`: caching directives (configurable)
- `ETag`: content hash for conditional requests (if enabled)

#### Response codes

| Code | Description |
|-----|----------|
| 200 | Image processed successfully |
| 400 | Invalid request parameters |
| 500 | Internal server error |

## Usage examples

### 1. Basic resize

Resize with aspect ratio preserved:

```bash
# Short parameter
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&w=800" -o resized.jpg

# Long parameter
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&width=800" -o resized.jpg
```

**Behavior:**
- Width will be 800px
- Height is calculated automatically
- Original format is preserved

### 2. Resize with both dimensions

Fit the image into a 1000x600 box:

```bash
# Short form
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&w=1000&h=600" -o fitted.jpg

# Or using s (resize)
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&s=1000x600" -o fitted.jpg
```

**Behavior:**
- Image scales to fit within 1000x600
- Aspect ratio is preserved
- Final size can be smaller than requested

### 3. Format conversion

Convert JPEG to WebP:

```bash
# Short form (f)
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&f=webp" -o photo.webp

# Long form (format)
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&format=webp" -o photo.webp
```

### 4. Convert to PNG without compression

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&f=png" -o photo.png
```

**Note:** PNG does not support `quality`, it is ignored.

### 5. Resize with quality control

```bash
# Short form
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&w=1200&q=95" -o high-quality.jpg

# Long form
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&width=1200&quality=95" -o high-quality.jpg
```

**Quality guidance:**
- `70-80`: good quality, smaller file size
- `85` (default): best balance
- `90-100`: high quality, larger file size

### 6. Web optimization (WebP + quality)

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/large.jpg&w=1200&f=webp&q=80" -o optimized.webp
```

### 7. Thumbnail creation

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&s=200x200&q=75" -o thumbnail.jpg
```

### 8. Blur and other effects

```bash
# Blur
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&blur=5" -o blurred.jpg

# Grayscale
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&grayscale=true" -o bw.jpg

# Rotate 90 degrees
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&rotate=90" -o rotated.jpg
```

### 9. Combine parameters

```bash
# Resize + format + quality + effect
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&s=800x600&f=webp&q=85&sharpen=1.5_1_2" -o processed.webp
```

### 10. Get original image

If no transform parameters are set, the original is returned:

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg" -o original.jpg
```

## Caching

### Internal cache

- In-memory cache at the server level. TTL is set by `Config.CacheTTL` (default `30s`).

### HTTP caching

- Managed by handler config:

```go
cfg := ipxpress.NewDefaultConfig()
cfg.ClientMaxAge = 3600 // Cache-Control: max-age
cfg.SMaxAge = 3600      // Cache-Control: s-maxage (CDN)
cfg.EnableETag = true   // ETag + If-None-Match => 304
handler := ipxpress.NewHandler(cfg)
```

- Client side / caching proxies:
  - If `If-None-Match` matches `ETag`, server returns `304 Not Modified`.
  - With `s-maxage`, a CDN can cache independently of client `max-age`.

## Resize behavior

### Width only (w)

```bash
?url=https://example.com/1000x500.jpg&w=500
# Result: 500x250
```

Height scales proportionally.

### Height only (h)

```bash
?url=https://example.com/1000x500.jpg&h=100
# Result: 200x100
```

Width scales proportionally.

### Width and height (w + h or s)

```bash
?url=https://example.com/1000x500.jpg&w=600&h=400
# or
?url=https://example.com/1000x500.jpg&s=600x400
# Result: 600x300 (fits within 600x400)
```

Image scales to fit within the specified rectangle while preserving aspect ratio.

## Supported formats

### Input formats

- JPEG / JPG
- PNG (including transparency)
- GIF (static)
- WebP

### Output formats

| Format | Value | Quality | Transparency | Notes |
|--------|----------|----------|--------------|------------|
| JPEG | `jpeg` or `jpg` | Yes | No | Best compression for photos |
| PNG | `png` | No | Yes | Lossless, for graphics |
| GIF | `gif` | No | Yes | Limited palette |
| WebP | `webp` | Yes | Yes | Modern format, good compression |
| AVIF | `avif` | Yes | Yes | Newest format, best compression |

## Performance and caching

### Caching

The service caches processed images for **30 seconds**. Repeat requests with the same parameters are served instantly.

**Cache key:**
```
MD5(url + width + height + quality + format)
```

### Cache-Control headers

```
Cache-Control: public, max-age=604800  # Processed images (7 days)
Cache-Control: public, max-age=31536000 # Original images (1 year)
```

### Recommendations

1. **CDN:** Put IPXpress behind a CDN for better performance
2. **Stable URLs:** Use stable image URLs
3. **Batch processing:** Send requests in parallel

## Limits

### Current limits

- Maximum 256 concurrent processing operations
- Fetch timeout: 20 seconds
- Connect timeout: 5 seconds
- HTTP/HTTPS URLs only

### Recommended practices

1. **Image sizes:**
   - Input: up to 20-30 MP
   - Output: reasonable sizes (up to 4000px on the longer side)

2. **Rate limiting:**
   - Limit requests per client
   - Use nginx/haproxy for rate limiting

3. **Monitoring:**
   - Track latency and error rate
   - Configure alerts for 5xx errors

## Integration

### JavaScript / Fetch API

```javascript
const imageUrl = encodeURIComponent('https://example.com/photo.jpg');
const apiUrl = `http://localhost:8080/ipx/?url=${imageUrl}&w=800&format=webp`;

fetch(apiUrl)
  .then(response => response.blob())
  .then(blob => {
    const img = document.createElement('img');
    img.src = URL.createObjectURL(blob);
    document.body.appendChild(img);
  });
```

### Python / requests

```python
import requests

params = {
    'url': 'https://example.com/photo.jpg',
    'w': 800,
    'format': 'webp',
    'quality': 85
}

response = requests.get('http://localhost:8080/ipx/', params=params)

with open('output.webp', 'wb') as f:
    f.write(response.content)
```

### Go

```go
package main

import (
    "io"
    "net/http"
    "net/url"
    "os"
)

func main() {
    params := url.Values{}
    params.Add("url", "https://example.com/photo.jpg")
    params.Add("w", "800")
    params.Add("format", "webp")

    apiURL := "http://localhost:8080/ipx/?" + params.Encode()

    resp, err := http.Get(apiURL)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    out, err := os.Create("output.webp")
    if err != nil {
        panic(err)
    }
    defer out.Close()

    io.Copy(out, resp.Body)
}
```

### HTML (direct usage)

```html
<img src="http://localhost:8080/ipx/?url=https://example.com/photo.jpg&w=400&format=webp"
     alt="Processed image">
```

## Error handling

### Error examples

#### Missing URL

```bash
curl "http://localhost:8080/ipx/"
# HTTP 400: missing image URL
```

#### Invalid URL

```bash
curl "http://localhost:8080/ipx/?url=not-a-valid-url"
# HTTP 400: invalid image URL: ...
```

#### Unavailable image

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/404.jpg"
# HTTP 400: image fetch failed with status 404
```

#### Processing error

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/corrupted.jpg&w=800"
# HTTP 500: processing: failed to decode image
```

### Handling in code

```javascript
fetch(apiUrl)
  .then(response => {
    if (!response.ok) {
      return response.text().then(text => {
        throw new Error(`Server error: ${text}`);
      });
    }
    return response.blob();
  })
  .catch(error => {
    console.error('Image processing failed:', error);
  });
```

## Health check

### Endpoint

```
GET /health
```

### Example

```bash
curl http://localhost:8080/health
# OK
```

Use this endpoint to monitor service availability.

## Additional resources

- [README.md](README.md) - Project overview
- [ARCHITECTURE.md](ARCHITECTURE.md) - Architecture and internals
- [Examples](examples/) - More integration examples
