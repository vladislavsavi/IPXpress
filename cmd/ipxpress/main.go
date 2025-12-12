package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/deadpixel/ipxpress/pkg/ipxpress"
)

func main() {
	// Initialize libvips on startup
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 0,   // 0 = use all available CPU cores
		MaxCacheMem:      100, // 100MB cache for better performance
		MaxCacheSize:     200, // 200 images in cache
		MaxCacheFiles:    0,   // No file cache
	})

	// Suppress VIPS info logs, only show warnings and errors
	vips.LoggingSettings(nil, vips.LogLevelWarning)

	// Ensure cleanup happens on shutdown
	defer vips.Shutdown()

	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	// Mount at /ipx/ to handle image processing requests
	mux.Handle("/ipx/", http.StripPrefix("/ipx/", ipxpress.Server()))

	fmt.Printf("starting ipxpress server on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
