package ipxpress_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func TestSingleflight(t *testing.T) {
	var backendRequests int32

	// 1. Создаем сервер изображений, который имитирует задержку
	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&backendRequests, 1)
		// Имитируем тяжелую работу или сетевую задержку
		// time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("fake-image-data"))
	}))
	defer imgServer.Close()

	// 2. Создаем IPXpress сервер
	handler := ipxpress.NewHandler(nil)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	imageURL := imgServer.URL + "/test.png"
	testURL := srv.URL + "/?url=" + url.QueryEscape(imageURL) + "&w=100"

	// 3. Делаем много одновременных запросов
	const numRequests = 50
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			resp, err := http.Get(testURL)
			if err != nil {
				t.Errorf("Request failed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}

	wg.Wait()

	// 4. Проверяем, что к бэкенду был сделан только ОДИН запрос благодаря singleflight
	finalBackendRequests := atomic.LoadInt32(&backendRequests)
	if finalBackendRequests != 1 {
		t.Errorf("Singleflight failed: expected 1 backend request, but got %d. This means requests were not grouped correctly.", finalBackendRequests)
	} else {
		t.Logf("Singleflight works: %d requests handled with only 1 backend call.", numRequests)
	}
}
