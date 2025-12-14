package ipxpress

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestServerResize(t *testing.T) {
	// Create a test image server that serves a PNG image
	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 100, 50))
		for y := 0; y < 50; y++ {
			for x := 0; x < 100; x++ {
				img.Set(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
			}
		}
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer imgServer.Close()

	// Create the ipxpress server
	mux := http.NewServeMux()
	mux.Handle("/ipx/", http.StripPrefix("/ipx/", Server()))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Make request with image URL and resize parameter
	resp, err := http.Get(srv.URL + "/ipx/?url=" + url.QueryEscape(imgServer.URL+"/image.png") + "&w=50")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct == "" {
		t.Fatalf("missing content-type")
	}

	outImg, _, err := image.Decode(resp.Body)
	if err != nil {
		t.Fatalf("decode out: %v", err)
	}
	b := outImg.Bounds()
	if b.Dx() != 50 || b.Dy() != 25 {
		t.Fatalf("unexpected size: %vx%v", b.Dx(), b.Dy())
	}
}

func TestServerDefaultsWithNilConfig(t *testing.T) {
	// Create a test image server that serves a PNG image
	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 40, 20))
		for y := 0; y < 20; y++ {
			for x := 0; x < 40; x++ {
				img.Set(x, y, color.RGBA{R: 5, G: 15, B: 25, A: 255})
			}
		}
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer imgServer.Close()

	// Ensure NewHandler(nil) uses default config without requiring settings
	handler := NewHandler(nil)

	mux := http.NewServeMux()
	mux.Handle("/ipx/", http.StripPrefix("/ipx/", handler))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Make a simple request that triggers processing
	resp, err := http.Get(srv.URL + "/ipx/?url=" + url.QueryEscape(imgServer.URL+"/image.png") + "&w=20")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct == "" {
		t.Fatalf("missing content-type")
	}

	outImg, _, err := image.Decode(resp.Body)
	if err != nil {
		t.Fatalf("decode out: %v", err)
	}
	b := outImg.Bounds()
	if b.Dx() != 20 || b.Dy() != 10 {
		t.Fatalf("unexpected size: %vx%v", b.Dx(), b.Dy())
	}
}
