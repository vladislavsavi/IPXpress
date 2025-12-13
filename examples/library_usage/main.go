package main

import (
	"log"
	"net/http"
	"time"

	"github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func main() {
	// Configure IPXpress
	config := &ipxpress.Config{
		ProcessingLimit: 10,
		CacheTTL:        30 * time.Minute,
		CleanupInterval: 5 * time.Minute,
	}

	handler := ipxpress.NewHandler(config)

	// Add processors
	handler.UseProcessor(ipxpress.AutoOrientProcessor())
	handler.UseProcessor(ipxpress.CompressionOptimizer())

	// Add middleware
	handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))

	// Setup server
	mux := http.NewServeMux()
	mux.Handle("/img/", http.StripPrefix("/img/", handler))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
