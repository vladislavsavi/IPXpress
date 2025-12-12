package ipxpress

import (
	"errors"
	"fmt"
	"io"

	"github.com/davidbyttow/govips/v2/vips"
)

// Processor is a chainable image processor using libvips backend.
type Processor struct {
	img            *vips.ImageRef
	err            error
	originalFormat Format
	originalSize   int
}

// New creates a new Processor instance.
func New() *Processor {
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

// ToBytes encodes the image to bytes in the given format.
// Supports: jpeg, png, gif, webp
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
		params.ReductionEffort = 4 // 0-6: higher = better compression, slower
		buf, _, err := p.img.ExportWebp(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode WebP: %w", err)
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
