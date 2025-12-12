package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/deadpixel/ipxpress/pkg/ipxpress"
)

func main() {
	// Initialize libvips for high-concurrency server workloads
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 0,    // Use all available CPU cores
		MaxCacheMem:      2048, // 2GB cache - more space for concurrent images
		MaxCacheSize:     5000, // 5000 images in cache - handle more concurrent requests
		MaxCacheFiles:    0,    // No file cache (memory only is faster)
	})

	// Suppress VIPS info logs, only show warnings and errors
	vips.LoggingSettings(nil, vips.LogLevelWarning)

	// Ensure cleanup happens on shutdown
	defer vips.Shutdown()

	// Start cache cleanup goroutine (less frequently to reduce contention)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			ipxpress.CleanupCache()
		}
	}()

	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	// Mount at /ipx/ to handle image processing requests
	mux.Handle("/ipx/", http.StripPrefix("/ipx/", ipxpress.Server()))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("starting ipxpress server on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
