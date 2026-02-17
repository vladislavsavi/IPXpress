package ipxpress_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

// TestJPEGToWebPCompression tests that WebP encoding produces smaller files than JPEG
func TestJPEGToWebPCompression(t *testing.T) {
	// Create a test image with some detail
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Fill with gradient pattern to simulate real photo
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			r := uint8((x * 255) / 800)
			g := uint8((y * 255) / 600)
			b := uint8(((x + y) * 255) / 1400)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Encode as JPEG with quality 85
	jpegBuf := &bytes.Buffer{}
	if err := jpeg.Encode(jpegBuf, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("failed to encode JPEG: %v", err)
	}
	jpegBytes := jpegBuf.Bytes()
	jpegSize := len(jpegBytes)

	t.Logf("Original JPEG size: %d bytes", jpegSize)

	// Convert JPEG to WebP with same quality
	proc := ipxpress.New().FromBytes(jpegBytes)
	if err := proc.Err(); err != nil {
		t.Fatalf("failed to load JPEG: %v", err)
	}

	webpBytes, err := proc.ToBytes(ipxpress.FormatWebP, 85)
	proc.Close()
	if err != nil {
		t.Fatalf("failed to encode WebP: %v", err)
	}
	webpSize := len(webpBytes)

	t.Logf("WebP size (q=85): %d bytes", webpSize)
	t.Logf("Size difference: %d bytes (%.1f%%)", jpegSize-webpSize, float64(jpegSize-webpSize)*100.0/float64(jpegSize))

	// WebP should typically be smaller or similar in size
	if webpSize > int(float64(jpegSize)*1.2) {
		t.Errorf("WebP is significantly larger than JPEG: %d vs %d (%.1f%% larger)",
			webpSize, jpegSize, float64(webpSize-jpegSize)*100.0/float64(jpegSize))
	}

	// Test with high quality
	proc2 := ipxpress.New().FromBytes(jpegBytes)
	webpBytes100, err := proc2.ToBytes(ipxpress.FormatWebP, 100)
	proc2.Close()
	if err != nil {
		t.Fatalf("failed to encode WebP q=100: %v", err)
	}
	webp100Size := len(webpBytes100)

	t.Logf("WebP size (q=100): %d bytes", webp100Size)
	t.Logf("Size difference from JPEG: %d bytes (%.1f%%)", jpegSize-webp100Size, float64(jpegSize-webp100Size)*100.0/float64(jpegSize))

	// Even at q=100, WebP shouldn't be dramatically larger
	if webp100Size > int(float64(jpegSize)*1.5) {
		t.Errorf("WebP q=100 is too large compared to JPEG: %d vs %d (%.1f%% larger)",
			webp100Size, jpegSize, float64(webp100Size-jpegSize)*100.0/float64(jpegSize))
	}
}

// TestFormatConversionOnly tests pure format conversion without reprocessing
func TestFormatConversionOnly(t *testing.T) {
	// Create simple test image
	rgbaImg := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			rgbaImg.Set(x, y, color.RGBA{
				R: uint8((x * 255) / 400),
				G: uint8((y * 255) / 300),
				B: 128,
				A: 255,
			})
		}
	}

	jpegBuf := &bytes.Buffer{}
	jpeg.Encode(jpegBuf, rgbaImg, &jpeg.Options{Quality: 85})
	jpegBytes := jpegBuf.Bytes()

	t.Logf("Test JPEG size: %d bytes", len(jpegBytes))

	// Convert to WebP with various quality levels
	qualities := []int{70, 80, 85, 90, 95, 100}

	for _, q := range qualities {
		proc := ipxpress.New().FromBytes(jpegBytes)
		webpBytes, err := proc.ToBytes(ipxpress.FormatWebP, q)
		proc.Close()

		if err != nil {
			t.Errorf("failed at q=%d: %v", q, err)
			continue
		}

		ratio := float64(len(webpBytes)) / float64(len(jpegBytes))
		t.Logf("WebP q=%d: %d bytes (%.1f%% of JPEG)", q, len(webpBytes), ratio*100)

		// For typical photos, WebP should be smaller or comparable
		if q <= 85 && len(webpBytes) > len(jpegBytes) {
			t.Logf("WARNING: WebP q=%d is larger than JPEG (%d vs %d)", q, len(webpBytes), len(jpegBytes))
		}
	}
}
