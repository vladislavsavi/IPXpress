package ipxpress_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func TestMemoryLeakAndCacheHit(t *testing.T) {
	// Pre-encode a 1000x1000 PNG with some "noise" to make it less compressible
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	for y := 0; y < 1000; y++ {
		for x := 0; x < 1000; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: uint8((x + y) % 256), A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgData := buf.Bytes()

	// Create a test image server
	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(imgData)
	}))
	defer imgServer.Close()

	// Configuration with 512MB cache limit
	cfg := ipxpress.DefaultConfig()
	cfg.CacheMaxCost = 512 * 1024 * 1024 
	handler := ipxpress.NewHandler(cfg)
	
	getMem := func() (uint64, int64) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		var vipsMem vips.MemoryStats
		vips.ReadVipsMemStats(&vipsMem)
		return m.Alloc / 1024 / 1024, int64(vipsMem.Mem) / 1024 / 1024 // MB
	}

	goMem, vMem := getMem()
	fmt.Printf("=== СТАРТ ТЕСТА ===\nНачальная память: Go=%d MB, Vips=%d MB\n", goMem, vMem)

	// Simulate 1000 different "photos"
	for i := 0; i < 1000; i++ {
		count := 1
		if i%10 == 0 { count = 3 }

		for j := 0; j < count; j++ {
			u := fmt.Sprintf("%s/img%d.png", imgServer.URL, i)
			// Apply some random-ish operations to prevent trivial caching/optimization
			req := httptest.NewRequest("GET", fmt.Sprintf("/ipx/?url=%s&w=%d&blur=%d", url.QueryEscape(u), 800+i%10, i%2+1), nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
		
		if i%100 == 0 {
			gm, vm := getMem()
			fmt.Printf("Обработано %d/1000... Go=%d MB, Vips=%d MB\n", i, gm, vm)
		}
	}

	// Final GC and measurement
	runtime.GC()
	time.Sleep(2 * time.Second)
	gm, vm := getMem()
	fmt.Printf("Финальная память после GC: Go=%d MB, Vips=%d MB\n", gm, vm)
}
