package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	// Create handler with custom config including vips settings
	config := ipxpress.DefaultConfig()
	config.VipsConfig = &ipxpress.VipsConfig{
		ConcurrencyLevel: 0, // Use all available CPU cores
		MaxCacheMem:      0, // Disable libvips caching (we manage cache at application level)
		MaxCacheSize:     0, // Disable libvips caching
		MaxCacheFiles:    0, // No file cache
		LogLevel:         vips.LogLevelWarning,
	}

	handler := ipxpress.NewHandler(config)

	// Add custom processors (optional - examples)
	handler.UseProcessor(ipxpress.AutoOrientProcessor())
	handler.UseProcessor(ipxpress.StripMetadataProcessor())

	// Add middlewares (optional - examples)
	handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))

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
