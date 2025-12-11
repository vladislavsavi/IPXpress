package ipxpress

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func init() {
	// Initialize libvips for tests
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 1,
	})
}

func TestResizePreservesAspect(t *testing.T) {
	// create a simple 100x50 PNG
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	// fill with a color
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("encode: %v", err)
	}

	proc := New().FromBytes(buf.Bytes()).Resize(50, 0) // constrain width
	defer proc.Close()

	if err := proc.Err(); err != nil {
		t.Fatalf("processor error: %v", err)
	}
	out, err := proc.ToBytes("png", 0)
	if err != nil {
		t.Fatalf("to bytes: %v", err)
	}

	outImg, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode out: %v", err)
	}
	b := outImg.Bounds()
	w := b.Dx()
	h := b.Dy()

	if w != 50 {
		t.Fatalf("expected width 50, got %d", w)
	}
	if h != 25 { // aspect preserved
		t.Fatalf("expected height 25, got %d", h)
	}
}
