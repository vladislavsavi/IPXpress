package ipxpress_test

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func TestCacheFunctional(t *testing.T) {
	var backendRequests int32

	// 1. Создаем исходный сервер изображений
	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&backendRequests, 1)
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			}
		}
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer imgServer.Close()

	// 2. Создаем IPXpress сервер с коротким TTL для теста
	config := ipxpress.DefaultConfig()
	config.CacheTTL = 2 * time.Second
	config.CacheMaxCost = 10 * 1024 * 1024 // 10MB
	
	handler := ipxpress.NewHandler(config)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	imageURL := imgServer.URL + "/test.png"
	testURL := srv.URL + "/?url=" + url.QueryEscape(imageURL) + "&w=50"

	// ЗАПРОС 1: Должен быть MISS (запрос к бэкенду)
	resp1, err := http.Get(testURL)
	if err != nil || resp1.StatusCode != 200 {
		t.Fatalf("First request failed: %v, status: %d", err, resp1.StatusCode)
	}
	resp1.Body.Close()

	if atomic.LoadInt32(&backendRequests) != 1 {
		t.Errorf("Expected 1 backend request, got %d", backendRequests)
	}

	// ЗАПРОС 2: Должен быть HIT (бэкенд не опрашивается)
	resp2, err := http.Get(testURL)
	if err != nil || resp2.StatusCode != 200 {
		t.Fatalf("Second request failed: %v, status: %d", err, resp2.StatusCode)
	}
	resp2.Body.Close()

	if atomic.LoadInt32(&backendRequests) != 1 {
		t.Errorf("Expected still 1 backend request (HIT), but got %d (MISS)", backendRequests)
	}

	// ЗАПРОС 3: Другие параметры - должен быть MISS
	testURL2 := testURL + "&format=webp"
	resp3, err := http.Get(testURL2)
	if err != nil || resp3.StatusCode != 200 {
		t.Fatalf("Third request failed: %v, status: %d", err, resp3.StatusCode)
	}
	resp3.Body.Close()

	if atomic.LoadInt32(&backendRequests) != 2 {
		t.Errorf("Expected 2 backend requests after params change, got %d", backendRequests)
	}

	// ЗАПРОС 4: После истечения TTL - должен быть MISS
	time.Sleep(3 * time.Second)
	resp4, err := http.Get(testURL)
	if err != nil || resp4.StatusCode != 200 {
		t.Fatalf("Fourth request failed: %v, status: %d", err, resp4.StatusCode)
	}
	resp4.Body.Close()

	if atomic.LoadInt32(&backendRequests) != 3 {
		t.Errorf("Expected 3 backend requests after TTL expiry, got %d", backendRequests)
	}

	t.Log("Functional cache test passed!")
}
