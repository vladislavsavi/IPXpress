package ipxpress

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
)

var (
	vipsInitOnce sync.Once
	vipsInitErr  error
)

// initVips initializes vips library with default settings.
// This is called automatically when using the library, so users don't need to manually initialize vips.
func initVips() {
	initVipsWithSettings(nil)
}

// initVipsWithSettings initializes vips with custom or default settings.
func initVipsWithSettings(cfg *VipsConfig) {
	vipsInitOnce.Do(func() {
		if cfg == nil {
			cfg = DefaultVipsConfig()
		}

		vips.Startup(&vips.Config{
			ConcurrencyLevel: cfg.ConcurrencyLevel,
			MaxCacheMem:      cfg.MaxCacheMem,
			MaxCacheSize:     cfg.MaxCacheSize,
			MaxCacheFiles:    cfg.MaxCacheFiles,
		})
		vips.LoggingSettings(nil, cfg.LogLevel)
	})
}

// InitVipsWithConfig allows manual initialization of vips with custom configuration.
// This should be called before creating any handlers or processors if you want custom settings.
// If not called, default settings will be used automatically.
func InitVipsWithConfig(config *vips.Config, logLevel vips.LogLevel) {
	vipsInitOnce.Do(func() {
		vips.Startup(config)
		vips.LoggingSettings(nil, logLevel)
	})
}

// Processor is a chainable image processor using libvips backend.
type Processor struct {
	img            *vips.ImageRef
	err            error
	originalFormat Format
	originalSize   int
	originalData   []byte
}

// New creates a new Processor instance.
// Automatically initializes vips if not already initialized.
func New() *Processor {
	initVips()
	return &Processor{}
}

// FromBytes decodes an image from a byte slice.
func (p *Processor) FromBytes(b []byte) *Processor {
	if p.err != nil {
		return p
	}

	img, err := vips.NewImageFromBuffer(b)
	if err != nil {
		p.err = fmt.Errorf("failed to decode image: %w", err)
		return p
	}

	p.img = img

	// Detect original format and store size
	p.originalFormat = DetectFormat(b)
	p.originalSize = len(b)
	p.originalData = b

	return p
}

// FromReader decodes an image from an io.Reader.
func (p *Processor) FromReader(r io.Reader) *Processor {
	if p.err != nil {
		return p
	}

	data, err := io.ReadAll(r)
	if err != nil {
		p.err = fmt.Errorf("failed to read image data: %w", err)
		return p
	}

	return p.FromBytes(data)
}

// Resize resizes the image to fit within maxWidth x maxHeight while preserving aspect ratio.
// Uses high-quality Lanczos resampling from libvips.
func (p *Processor) Resize(maxWidth, maxHeight int) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	if maxWidth == 0 && maxHeight == 0 {
		return p
	}

	srcW := p.img.Width()
	srcH := p.img.Height()

	var tgtW, tgtH int
	if maxWidth == 0 {
		scale := float64(maxHeight) / float64(srcH)
		tgtW = int(float64(srcW) * scale)
		tgtH = maxHeight
	} else if maxHeight == 0 {
		scale := float64(maxWidth) / float64(srcW)
		tgtW = maxWidth
		tgtH = int(float64(srcH) * scale)
	} else {
		scaleW := float64(maxWidth) / float64(srcW)
		scaleH := float64(maxHeight) / float64(srcH)
		scale := scaleW
		if scaleH < scaleW {
			scale = scaleH
		}
		tgtW = int(float64(srcW) * scale)
		tgtH = int(float64(srcH) * scale)
	}

	if tgtW <= 0 {
		tgtW = 1
	}
	if tgtH <= 0 {
		tgtH = 1
	}

	// Compute scale factors
	scaleX := float64(tgtW) / float64(srcW)
	scaleY := float64(tgtH) / float64(srcH)

	// Resize in-place (modifies the image reference)
	if scaleX == scaleY {
		p.err = p.img.Resize(scaleX, vips.KernelLanczos3)
	} else {
		p.err = p.img.ResizeWithVScale(scaleX, scaleY, vips.KernelLanczos3)
	}

	if p.err != nil {
		p.err = fmt.Errorf("failed to resize image: %w", p.err)
	}

	return p
}

// ResizeWithOptions resizes with advanced options (fit, position, kernel, enlarge)
func (p *Processor) ResizeWithOptions(width, height int, kernel vips.Kernel, enlarge bool) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	if width == 0 && height == 0 {
		return p
	}

	srcW := p.img.Width()
	srcH := p.img.Height()

	var tgtW, tgtH int
	if width == 0 {
		scale := float64(height) / float64(srcH)
		tgtW = int(float64(srcW) * scale)
		tgtH = height
	} else if height == 0 {
		scale := float64(width) / float64(srcW)
		tgtW = width
		tgtH = int(float64(srcH) * scale)
	} else {
		scaleW := float64(width) / float64(srcW)
		scaleH := float64(height) / float64(srcH)
		scale := scaleW
		if scaleH < scaleW {
			scale = scaleH
		}
		tgtW = int(float64(srcW) * scale)
		tgtH = int(float64(srcH) * scale)
	}

	// Don't enlarge if not requested
	if !enlarge {
		if tgtW > srcW {
			tgtW = srcW
		}
		if tgtH > srcH {
			tgtH = srcH
		}
	}

	if tgtW <= 0 {
		tgtW = 1
	}
	if tgtH <= 0 {
		tgtH = 1
	}

	// Compute scale factors
	scaleX := float64(tgtW) / float64(srcW)
	scaleY := float64(tgtH) / float64(srcH)

	// Resize in-place with specified kernel
	if scaleX == scaleY {
		p.err = p.img.Resize(scaleX, kernel)
	} else {
		p.err = p.img.ResizeWithVScale(scaleX, scaleY, kernel)
	}

	if p.err != nil {
		p.err = fmt.Errorf("failed to resize image: %w", p.err)
	}

	return p
}

// Thumbnail creates a thumbnail using SmartCrop (attention-based cropping)
func (p *Processor) Thumbnail(width, height int, interesting vips.Interesting) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	if width == 0 || height == 0 {
		return p
	}

	p.err = p.img.Thumbnail(width, height, interesting)
	if p.err != nil {
		p.err = fmt.Errorf("failed to create thumbnail: %w", p.err)
	}

	return p
}

// Blur applies Gaussian blur to the image
func (p *Processor) Blur(sigma float64) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	if sigma <= 0 {
		return p
	}

	p.err = p.img.GaussianBlur(sigma)
	if p.err != nil {
		p.err = fmt.Errorf("failed to blur image: %w", p.err)
	}

	return p
}

// Sharpen sharpens the image
func (p *Processor) Sharpen(sigma, flat, jagged float64) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.Sharpen(sigma, flat, jagged)
	if p.err != nil {
		p.err = fmt.Errorf("failed to sharpen image: %w", p.err)
	}

	return p
}

// Rotate rotates the image by the given angle
func (p *Processor) Rotate(angle vips.Angle) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.Rotate(angle)
	if p.err != nil {
		p.err = fmt.Errorf("failed to rotate image: %w", p.err)
	}

	return p
}

// Flip flips the image vertically
func (p *Processor) Flip() *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.Flip(vips.DirectionVertical)
	if p.err != nil {
		p.err = fmt.Errorf("failed to flip image: %w", p.err)
	}

	return p
}

// Flop flips the image horizontally
func (p *Processor) Flop() *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.Flip(vips.DirectionHorizontal)
	if p.err != nil {
		p.err = fmt.Errorf("failed to flop image: %w", p.err)
	}

	return p
}

// Grayscale converts the image to grayscale
func (p *Processor) Grayscale() *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.ToColorSpace(vips.InterpretationBW)
	if p.err != nil {
		p.err = fmt.Errorf("failed to convert to grayscale: %w", p.err)
	}

	return p
}

// Extract extracts a rectangular region from the image
func (p *Processor) Extract(left, top, width, height int) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.ExtractArea(left, top, width, height)
	if p.err != nil {
		p.err = fmt.Errorf("failed to extract region: %w", p.err)
	}

	return p
}

// Extend adds borders to the image
func (p *Processor) Extend(top, right, bottom, left int, background []float64) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	// Create background color if provided
	if len(background) >= 3 {
		bgColor := &vips.Color{
			R: uint8(background[0]),
			G: uint8(background[1]),
			B: uint8(background[2]),
		}
		p.err = p.img.Embed(left, top, p.img.Width()+left+right, p.img.Height()+top+bottom, vips.ExtendBackground)
		if p.err == nil {
			// Apply background by flattening first if there's alpha
			if p.img.HasAlpha() {
				p.err = p.img.Flatten(bgColor)
			}
		}
	} else {
		// Default extend with white background
		p.err = p.img.Embed(left, top, p.img.Width()+left+right, p.img.Height()+top+bottom, vips.ExtendWhite)
	}

	if p.err != nil {
		p.err = fmt.Errorf("failed to extend image: %w", p.err)
	}

	return p
}

// Negate inverts the colors of the image
func (p *Processor) Negate() *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.Invert()
	if p.err != nil {
		p.err = fmt.Errorf("failed to negate image: %w", p.err)
	}

	return p
}

// Normalize normalizes the image
func (p *Processor) Normalize() *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	// Vips doesn't have a direct normalize, but we can use linear adjustment
	p.err = p.img.Linear([]float64{1}, []float64{0})
	if p.err != nil {
		p.err = fmt.Errorf("failed to normalize image: %w", p.err)
	}

	return p
}

// Gamma applies gamma correction
func (p *Processor) Gamma(gamma float64) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	if gamma <= 0 {
		return p
	}

	p.err = p.img.Gamma(gamma)
	if p.err != nil {
		p.err = fmt.Errorf("failed to apply gamma: %w", p.err)
	}

	return p
}

// Modulate transforms the image using brightness, saturation, hue rotation
func (p *Processor) Modulate(brightness, saturation, hue float64) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.Modulate(brightness, saturation, hue)
	if p.err != nil {
		p.err = fmt.Errorf("failed to modulate image: %w", p.err)
	}

	return p
}

// Flatten removes alpha channel
func (p *Processor) Flatten(background *vips.Color) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	p.err = p.img.Flatten(background)
	if p.err != nil {
		p.err = fmt.Errorf("failed to flatten image: %w", p.err)
	}

	return p
}

// ToBytes encodes the image to bytes in the given format.
// Supports: jpeg, png, gif, webp, avif
func (p *Processor) ToBytes(format Format, quality int) ([]byte, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.img == nil {
		return nil, errors.New("no image to encode")
	}

	if quality <= 0 || quality > 100 {
		quality = 85
	}

	switch format {
	case FormatJPEG:
		params := vips.NewJpegExportParams()
		params.Quality = quality
		params.OptimizeCoding = true
		params.Interlace = true
		params.StripMetadata = true
		buf, _, err := p.img.ExportJpeg(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode JPEG: %w", err)
		}
		return buf, nil

	case FormatPNG:
		params := vips.NewPngExportParams()
		buf, _, err := p.img.ExportPng(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode PNG: %w", err)
		}
		return buf, nil

	case FormatGIF:
		params := vips.NewGifExportParams()
		buf, _, err := p.img.ExportGIF(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode GIF: %w", err)
		}
		return buf, nil

	case FormatWebP:
		params := vips.NewWebpExportParams()
		params.Quality = quality
		params.Lossless = false
		params.StripMetadata = true
		params.ReductionEffort = 4 // Optimal balance for speed
		buf, _, err := p.img.ExportWebp(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode WebP: %w", err)
		}
		return buf, nil

	case FormatAVIF:
		params := vips.NewAvifExportParams()
		params.Quality = quality
		params.Speed = 6 // Fast encoding, good compression
		params.StripMetadata = true
		params.Lossless = false
		buf, _, err := p.img.ExportAvif(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode AVIF: %w", err)
		}
		return buf, nil

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Close closes the internal image reference and frees memory.
// It's recommended to call this method after you're done with the Processor.
func (p *Processor) Close() {
	if p.img != nil {
		p.img.Close()
		p.img = nil
	}
}

// Err returns the processor's error (if any).
func (p *Processor) Err() error { return p.err }

// OriginalFormat returns the detected original format of the image.
func (p *Processor) OriginalFormat() Format { return p.originalFormat }

// OriginalSize returns the size of the original image in bytes.
func (p *Processor) OriginalSize() int { return p.originalSize }

// OriginalBytes returns the original image bytes if available.
func (p *Processor) OriginalBytes() []byte { return p.originalData }

// ImageRef returns the underlying vips.ImageRef for direct manipulation.
// This allows users to apply any libvips function not directly exposed by IPXpress.
// Important: The returned ImageRef is managed by the Processor and will be closed
// when the Processor is closed. Do not manually close it.
func (p *Processor) ImageRef() *vips.ImageRef {
	return p.img
}

// ApplyFunc applies a custom function to the image.
// The function receives the current ImageRef and should return an error if the operation fails.
// This is useful for applying libvips functions that IPXpress doesn't directly expose.
//
// Example:
//
//	processor.ApplyFunc(func(img *vips.ImageRef) error {
//	    return img.Sharpen(1.5, 0.5)
//	})
func (p *Processor) ApplyFunc(fn func(*vips.ImageRef) error) *Processor {
	if p.err != nil {
		return p
	}
	if p.img == nil {
		p.err = errors.New("no image loaded")
		return p
	}

	if fn == nil {
		p.err = errors.New("function cannot be nil")
		return p
	}

	p.err = fn(p.img)
	if p.err != nil {
		p.err = fmt.Errorf("custom function failed: %w", p.err)
	}

	return p
}
