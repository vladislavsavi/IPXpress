package ipxpress_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

// BenchmarkCacheGet benchmarks cache read performance
func BenchmarkCacheGet(b *testing.B) {
	cache := ipxpress.NewInMemoryCache(10*time.Minute, 1000)

	entry := &ipxpress.CacheEntry{
		ContentType: "image/jpeg",
		Data:        make([]byte, 10240),
		StatusCode:  200,
	}

	cache.Set("test-key", entry)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, ok := cache.Get("test-key")
			if !ok {
				b.Fatal("cache miss on existing key")
			}
		}
	})
}

// BenchmarkCacheSet benchmarks cache write performance
func BenchmarkCacheSet(b *testing.B) {
	cache := ipxpress.NewInMemoryCache(10*time.Minute, 10000)

	entry := &ipxpress.CacheEntry{
		ContentType: "image/jpeg",
		Data:        make([]byte, 10240),
		StatusCode:  200,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i)
			cache.Set(key, entry)
			i++
		}
	})
}

// BenchmarkCacheMixedLoad benchmarks mixed read/write operations
func BenchmarkCacheMixedLoad(b *testing.B) {
	cache := ipxpress.NewInMemoryCache(10*time.Minute, 1000)

	for i := 0; i < 100; i++ {
		entry := &ipxpress.CacheEntry{
			ContentType: "image/jpeg",
			Data:        make([]byte, 10240),
			StatusCode:  200,
		}
		cache.Set(fmt.Sprintf("key-%d", i), entry)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			keyNum := r.Intn(150)
			key := fmt.Sprintf("key-%d", keyNum)

			if r.Float32() < 0.7 {
				cache.Get(key)
			} else {
				entry := &ipxpress.CacheEntry{
					ContentType: "image/jpeg",
					Data:        make([]byte, 10240),
					StatusCode:  200,
				}
				cache.Set(key, entry)
			}
		}
	})
}

// TestCacheConcurrency tests cache under concurrent load
func TestCacheConcurrency(t *testing.T) {
	cache := ipxpress.NewInMemoryCache(5*time.Second, 500)

	const (
		numGoroutines = 100
		numOperations = 1000
	)

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	testData := createTestImageData(200, 200)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for j := 0; j < numOperations; j++ {
				keyNum := r.Intn(100)
				key := fmt.Sprintf("image-%d", keyNum)

				op := r.Float32()

				if op < 0.5 {
					entry, ok := cache.Get(key)
					if ok && entry == nil {
						errors <- fmt.Errorf("worker %d: got nil entry for key %s", workerID, key)
						return
					}
				} else {
					entry := &ipxpress.CacheEntry{
						ContentType: "image/jpeg",
						Data:        testData,
						StatusCode:  200,
					}
					cache.Set(key, entry)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	t.Logf("Cache concurrency test completed: %d goroutines, %d operations each", numGoroutines, numOperations)
}

// TestCacheTTLExpiration tests that cache entries expire correctly under load
func TestCacheTTLExpiration(t *testing.T) {
	ttl := 200 * time.Millisecond
	cache := ipxpress.NewInMemoryCache(ttl, 100)

	for i := 0; i < 10; i++ {
		entry := &ipxpress.CacheEntry{
			ContentType: "image/jpeg",
			Data:        []byte(fmt.Sprintf("data-%d", i)),
			StatusCode:  200,
		}
		cache.Set(fmt.Sprintf("key-%d", i), entry)
	}

	for i := 0; i < 10; i++ {
		_, ok := cache.Get(fmt.Sprintf("key-%d", i))
		if !ok {
			t.Errorf("Entry %d not found immediately after set", i)
		}
	}

	time.Sleep(ttl + 50*time.Millisecond)

	expiredCount := 0
	for i := 0; i < 10; i++ {
		_, ok := cache.Get(fmt.Sprintf("key-%d", i))
		if !ok {
			expiredCount++
		}
	}

	if expiredCount == 0 {
		t.Error("No entries expired after TTL")
	}

	t.Logf("Expired %d out of 10 entries after TTL", expiredCount)
}

// TestCacheLRUEviction tests LRU eviction under capacity pressure
func TestCacheLRUEviction(t *testing.T) {
	capacity := 50
	cache := ipxpress.NewInMemoryCache(10*time.Minute, capacity)

	for i := 0; i < capacity; i++ {
		entry := &ipxpress.CacheEntry{
			ContentType: "image/jpeg",
			Data:        []byte(fmt.Sprintf("data-%d", i)),
			StatusCode:  200,
		}
		cache.Set(fmt.Sprintf("key-%d", i), entry)
	}

	for i := 0; i < 25; i++ {
		cache.Get(fmt.Sprintf("key-%d", i))
	}

	for i := capacity; i < capacity+30; i++ {
		entry := &ipxpress.CacheEntry{
			ContentType: "image/jpeg",
			Data:        []byte(fmt.Sprintf("data-%d", i)),
			StatusCode:  200,
		}
		cache.Set(fmt.Sprintf("key-%d", i), entry)
	}

	recentlyAccessedStillPresent := 0
	for i := 0; i < 25; i++ {
		if _, ok := cache.Get(fmt.Sprintf("key-%d", i)); ok {
			recentlyAccessedStillPresent++
		}
	}

	lruEvicted := 0
	for i := 25; i < capacity; i++ {
		if _, ok := cache.Get(fmt.Sprintf("key-%d", i)); !ok {
			lruEvicted++
		}
	}

	t.Logf("Recently accessed entries still present: %d/25", recentlyAccessedStillPresent)
	t.Logf("LRU entries evicted: %d/25", lruEvicted)

	if recentlyAccessedStillPresent < 15 {
		t.Errorf("Too few recently accessed entries preserved: %d/25", recentlyAccessedStillPresent)
	}
}

// TestCacheHighThroughput tests cache under high throughput scenario
func TestCacheHighThroughput(t *testing.T) {
	cache := ipxpress.NewInMemoryCache(30*time.Second, 1000)

	const (
		numWorkers = 50
		duration   = 2 * time.Second
		uniqueKeys = 200
	)

	var (
		wg          sync.WaitGroup
		totalOps    int64
		cacheHits   int64
		cacheMisses int64
		mu          sync.Mutex
	)

	testData := createTestImageData(150, 150)
	done := make(chan struct{})

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			localOps := 0
			localHits := 0
			localMisses := 0

			for {
				select {
				case <-done:
					mu.Lock()
					totalOps += int64(localOps)
					cacheHits += int64(localHits)
					cacheMisses += int64(localMisses)
					mu.Unlock()
					return
				default:
					keyNum := r.Intn(uniqueKeys)
					key := fmt.Sprintf("img-%d-%d-%d", keyNum, r.Intn(5), r.Intn(3))

					if r.Float32() < 0.6 {
						_, ok := cache.Get(key)
						if ok {
							localHits++
						} else {
							localMisses++
						}
					} else {
						entry := &ipxpress.CacheEntry{
							ContentType: "image/jpeg",
							Data:        testData,
							StatusCode:  200,
						}
						cache.Set(key, entry)
					}
					localOps++
				}
			}
		}(i)
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	opsPerSecond := float64(totalOps) / duration.Seconds()
	hitRate := float64(cacheHits) / float64(cacheHits+cacheMisses) * 100

	t.Logf("High throughput test results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Operations/second: %.0f", opsPerSecond)
	t.Logf("  Cache hits: %d", cacheHits)
	t.Logf("  Cache misses: %d", cacheMisses)
	t.Logf("  Hit rate: %.2f%%", hitRate)

	if totalOps == 0 {
		t.Error("No operations completed")
	}
}

// TestCacheGenerateCacheKey tests cache key generation under load
func TestCacheGenerateCacheKey(t *testing.T) {
	const numKeys = 10000

	keys := make(map[string]bool)
	collisions := 0

	for i := 0; i < numKeys; i++ {
		url := fmt.Sprintf("https://example.com/image%d.jpg", i)
		width := (i % 10) * 100
		height := (i % 8) * 100
		quality := 70 + ((i % 4) * 10)
		format := ipxpress.Format([]string{"jpeg", "png", "webp"}[i%3])

		key := ipxpress.GenerateCacheKey(url, width, height, quality, format)

		if keys[key] {
			collisions++
		} else {
			keys[key] = true
		}
	}

	uniqueKeys := len(keys)
	t.Logf("Generated %d keys, %d unique, %d collisions", numKeys, uniqueKeys, collisions)

	if collisions > 0 {
		t.Errorf("Cache key collisions detected: %d", collisions)
	}
}

func createTestImageData(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

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

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	return buf.Bytes()
}
