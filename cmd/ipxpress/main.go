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

	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	// Create handler with default config
	config := ipxpress.DefaultConfig()
	handler := ipxpress.NewHandler(config)

	// Start cache cleanup goroutine
	go func() {
		ticker := time.NewTicker(config.CleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			handler.CleanupCache()
		}
	}()

	mux := http.NewServeMux()
	// Mount at /ipx/ to handle image processing requests
	mux.Handle("/ipx/", http.StripPrefix("/ipx/", handler))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("starting ipxpress server on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
