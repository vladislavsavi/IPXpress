package ipxpress

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

// Processor is a chainable image processor.
type Processor struct {
	img image.Image
	err error
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

	// Try to decode using standard formats first
	r := bytes.NewReader(b)
	img, _, err := image.Decode(r)
	if err == nil {
		p.img = img
		return p
	}

	// Try WebP
	r = bytes.NewReader(b)
	img, err = webp.Decode(r)
	if err == nil {
		p.img = img
		return p
	}

	// If all fail, return error
	p.err = fmt.Errorf("unsupported image format")
	return p
}

// Resize resizes the image to fit within maxWidth x maxHeight while preserving aspect ratio.
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

	srcBounds := p.img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

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

	p.img = imaging.Resize(p.img, tgtW, tgtH, imaging.Lanczos)
	return p
}

// ToBytes encodes the image to bytes in the given format.
// Supports: jpeg, jpg, png, gif, webp
func (p *Processor) ToBytes(format string, quality int) ([]byte, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.img == nil {
		return nil, errors.New("no image to encode")
	}

	format = strings.ToLower(format)

	if quality <= 0 || quality > 100 {
		quality = 85
	}

	buf := &bytes.Buffer{}
	var err error

	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, p.img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(buf, p.img)
	case "gif":
		err = gif.Encode(buf, p.img, nil)
	case "webp":
		err = webp.Encode(buf, p.img, &webp.Options{Quality: float32(quality)})
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Err returns the processor's error (if any).
func (p *Processor) Err() error { return p.err }
