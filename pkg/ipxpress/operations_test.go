package ipxpress

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

// TestBlurOperation tests the Blur method
func TestBlurOperation(t *testing.T) {
	// Create a simple test image
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Blur(3.0)

	if proc.Err() != nil {
		t.Errorf("Blur operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestSharpenOperation tests the Sharpen method
func TestSharpenOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Sharpen(1.5, 1.0, 2.0)

	if proc.Err() != nil {
		t.Errorf("Sharpen operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestRotateOperation tests rotation
func TestRotateOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Rotate(vips.Angle90)

	if proc.Err() != nil {
		t.Errorf("Rotate operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestFlipOperation tests vertical flip
func TestFlipOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Flip()

	if proc.Err() != nil {
		t.Errorf("Flip operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestFlopOperation tests horizontal flip
func TestFlopOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Flop()

	if proc.Err() != nil {
		t.Errorf("Flop operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestGrayscaleOperation tests grayscale conversion
func TestGrayscaleOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Grayscale()

	if proc.Err() != nil {
		t.Errorf("Grayscale operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestExtractOperation tests region extraction
func TestExtractOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Extract(10, 10, 50, 50)

	if proc.Err() != nil {
		t.Errorf("Extract operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestNegateOperation tests color inversion
func TestNegateOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Negate()

	if proc.Err() != nil {
		t.Errorf("Negate operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestGammaOperation tests gamma correction
func TestGammaOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Gamma(2.2)

	if proc.Err() != nil {
		t.Errorf("Gamma operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestModulateOperation tests HSB modulation
func TestModulateOperation(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	proc.Modulate(1.2, 0.8, 90.0)

	if proc.Err() != nil {
		t.Errorf("Modulate operation failed: %v", proc.Err())
	}

	proc.Close()
}

// TestChainedOperations tests multiple operations in sequence
func TestChainedOperations(t *testing.T) {
	img := createTestImage(200, 200)

	proc := New().
		FromBytes(img).
		Resize(100, 100).
		Blur(2.0).
		Sharpen(1.0, 1.0, 2.0).
		Grayscale()

	if proc.Err() != nil {
		t.Errorf("Chained operations failed: %v", proc.Err())
	}

	output, err := proc.ToBytes(FormatJPEG, 85)
	proc.Close()

	if err != nil {
		t.Errorf("Failed to encode after chained operations: %v", err)
	}

	if len(output) == 0 {
		t.Error("Output is empty after chained operations")
	}
}

// TestResizeWithOptions tests advanced resize
func TestResizeWithOptions(t *testing.T) {
	img := createTestImage(100, 100)

	tests := []struct {
		name    string
		kernel  vips.Kernel
		enlarge bool
	}{
		{"Lanczos3", vips.KernelLanczos3, false},
		{"Cubic", vips.KernelCubic, false},
		{"Nearest", vips.KernelNearest, false},
		{"WithEnlarge", vips.KernelLanczos3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc := New().FromBytes(img)
			proc.ResizeWithOptions(200, 200, tt.kernel, tt.enlarge)

			if proc.Err() != nil {
				t.Errorf("ResizeWithOptions failed: %v", proc.Err())
			}

			proc.Close()
		})
	}
}

// TestNoUpscaleEnsured verifies that output dimensions never exceed original when enlarge=false
func TestNoUpscaleEnsured(t *testing.T) {
	// Original image is 100x100
	img := createTestImage(100, 100)

	// Request a larger size but with enlarge=false
	proc := New().FromBytes(img)
	proc.ResizeWithOptions(300, 300, vips.KernelLanczos3, false)

	if err := proc.Err(); err != nil {
		t.Fatalf("processing error: %v", err)
	}

	// Encode to PNG and check dimensions
	out, err := proc.ToBytes(FormatPNG, 0)
	proc.Close()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	decoded, format, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = format

	b := decoded.Bounds()
	if b.Dx() > 100 || b.Dy() > 100 {
		t.Fatalf("upscale detected: got %dx%d, original 100x100", b.Dx(), b.Dy())
	}
}

// TestAVIFFormat tests AVIF encoding
func TestAVIFFormat(t *testing.T) {
	img := createTestImage(100, 100)

	proc := New().FromBytes(img)
	output, err := proc.ToBytes(FormatAVIF, 85)
	proc.Close()

	if err != nil {
		t.Errorf("AVIF encoding failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("AVIF output is empty")
	}

	// Verify AVIF magic bytes
	if len(output) >= 12 {
		// AVIF should have 'ftyp' at bytes 4-7
		if output[4] != 0x66 || output[5] != 0x74 || output[6] != 0x79 || output[7] != 0x70 {
			t.Error("Output doesn't appear to be valid AVIF format")
		}
	}
}

// TestFormatDetection tests format detection for all supported formats
func TestFormatDetection(t *testing.T) {
	tests := []struct {
		format   Format
		expected Format
	}{
		{FormatJPEG, FormatJPEG},
		{FormatPNG, FormatPNG},
		{FormatWebP, FormatWebP},
		{FormatGIF, FormatGIF},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			img := createTestImage(50, 50)
			proc := New().FromBytes(img)

			output, err := proc.ToBytes(tt.format, 85)
			proc.Close()

			if err != nil {
				t.Fatalf("Failed to encode as %s: %v", tt.format, err)
			}

			detected := DetectFormat(output)
			if detected != tt.expected {
				t.Errorf("Format detection mismatch: got %s, want %s", detected, tt.expected)
			}
		})
	}
}

// createTestImage creates a simple test image with gradient
func createTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create a simple gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 128,
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	// Convert to JPEG bytes using standard library
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil
	}

	return buf.Bytes()
}
