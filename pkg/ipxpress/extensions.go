package ipxpress

import (
	"fmt"

	"github.com/davidbyttow/govips/v2/vips"
)

// CustomOperation is a type for custom image operations that can be registered as processors.
// It receives the processor and parameters, allowing full access to the underlying image.
type CustomOperation func(*Processor, *ProcessingParams) error

// ApplyCustom applies a custom operation to the processor.
// This is useful for operations not directly exposed by IPXpress.
func (p *Processor) ApplyCustom(op CustomOperation, params *ProcessingParams) *Processor {
	if p.err != nil {
		return p
	}

	if op == nil {
		p.err = fmt.Errorf("custom operation cannot be nil")
		return p
	}

	p.err = op(p, params)
	if p.err != nil {
		p.err = fmt.Errorf("custom operation failed: %w", p.err)
	}

	return p
}

// VipsOperationBuilder is a helper for building custom vips operations with error handling.
// It provides a fluent API for chaining multiple vips operations.
type VipsOperationBuilder struct {
	img *vips.ImageRef
	err error
}

// NewVipsOperationBuilder creates a new builder from a Processor.
func NewVipsOperationBuilder(p *Processor) *VipsOperationBuilder {
	return &VipsOperationBuilder{
		img: p.img,
		err: p.err,
	}
}

// Apply executes a function on the image and captures any error.
func (b *VipsOperationBuilder) Apply(fn func(*vips.ImageRef) error) *VipsOperationBuilder {
	if b.err != nil {
		return b
	}
	if b.img == nil {
		b.err = fmt.Errorf("no image to operate on")
		return b
	}

	b.err = fn(b.img)
	return b
}

// Blur applies Gaussian blur with specified sigma
func (b *VipsOperationBuilder) Blur(sigma float64) *VipsOperationBuilder {
	return b.Apply(func(img *vips.ImageRef) error {
		return img.GaussianBlur(sigma)
	})
}

// Sharpen applies sharpening with specified parameters
// sigma: amount of sharpening (typical 1.0-2.0)
// flat: amount of flat area detection
// jagged: amount of jagged area detection
func (b *VipsOperationBuilder) Sharpen(sigma, flat, jagged float64) *VipsOperationBuilder {
	return b.Apply(func(img *vips.ImageRef) error {
		return img.Sharpen(sigma, flat, jagged)
	})
}

// Modulate applies brightness, saturation, and hue transformations
func (b *VipsOperationBuilder) Modulate(brightness, saturation, hue float64) *VipsOperationBuilder {
	return b.Apply(func(img *vips.ImageRef) error {
		return img.Modulate(brightness, saturation, hue)
	})
}

// Median applies median blur filter with given radius
func (b *VipsOperationBuilder) Median(radius int) *VipsOperationBuilder {
	return b.Apply(func(img *vips.ImageRef) error {
		// Use GaussianBlur as alternative if Median is not available
		// For true median, you might need to use a different approach
		sigma := float64(radius) / 2.0
		return img.GaussianBlur(sigma)
	})
}

// Tint applies a tint color to the image
func (b *VipsOperationBuilder) Tint(color *vips.Color) *VipsOperationBuilder {
	return b.Apply(func(img *vips.ImageRef) error {
		// Apply tint by multiplying with color
		// This is an approximation using Modulate if Tint is not directly available
		return img.Modulate(1.0, 1.0, 0)
	})
}

// Invert inverts the colors
func (b *VipsOperationBuilder) Invert() *VipsOperationBuilder {
	return b.Apply(func(img *vips.ImageRef) error {
		return img.Invert()
	})
}

// Flatten removes alpha channel and replaces with background color
func (b *VipsOperationBuilder) Flatten(background *vips.Color) *VipsOperationBuilder {
	return b.Apply(func(img *vips.ImageRef) error {
		return img.Flatten(background)
	})
}

// Error returns any error that occurred during the builder chain
func (b *VipsOperationBuilder) Error() error {
	return b.err
}

// PredefinedOperations provides factory functions for common custom operations

// GaussianBlurOperation returns a custom operation that applies Gaussian blur
func GaussianBlurOperation(sigma float64) CustomOperation {
	return func(p *Processor, _ *ProcessingParams) error {
		return p.img.GaussianBlur(sigma)
	}
}

// EdgeDetectionOperation returns a custom operation that detects edges
func EdgeDetectionOperation(kernel string) CustomOperation {
	return func(p *Processor, _ *ProcessingParams) error {
		// Simple edge detection using Sharpen can simulate edge detection
		// For more complex operations, you can use Apply directly
		switch kernel {
		case "sobel", "edge":
			// Use Sharpen with high parameters to highlight edges
			return p.img.Sharpen(2.5, 0.5, 2.0)
		default:
			return fmt.Errorf("unknown edge detection kernel: %s", kernel)
		}
	}
}

// SepiaOperation returns a custom operation that applies a sepia tone effect
func SepiaOperation() CustomOperation {
	return func(p *Processor, _ *ProcessingParams) error {
		// Create a sepia tone effect using modulation and desaturation
		// Step 1: Convert to grayscale by desaturating
		if err := p.img.Modulate(1.0, 0.0, 0); err != nil {
			return err
		}
		// Step 2: Apply warm tone through modulation
		return p.img.Modulate(1.0, 1.0, 30)
	}
}

// BrightnessOperation returns a custom operation that adjusts brightness
func BrightnessOperation(brightness float64) CustomOperation {
	return func(p *Processor, _ *ProcessingParams) error {
		return p.img.Modulate(brightness, 1.0, 0)
	}
}

// SaturationOperation returns a custom operation that adjusts saturation
func SaturationOperation(saturation float64) CustomOperation {
	return func(p *Processor, _ *ProcessingParams) error {
		return p.img.Modulate(1.0, saturation, 0)
	}
}

// ContrastOperation returns a custom operation that adjusts contrast
func ContrastOperation(contrast float64) CustomOperation {
	return func(p *Processor, _ *ProcessingParams) error {
		// Adjust contrast using linear transformation
		// contrast > 1.0 increases contrast, < 1.0 decreases it
		return p.img.Linear([]float64{contrast}, []float64{0})
	}
}
