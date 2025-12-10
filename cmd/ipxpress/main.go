package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/deadpixel/ipxpress/pkg/ipxpress"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	// Mount at /ipx/ to handle image processing requests
	mux.Handle("/ipx/", http.StripPrefix("/ipx/", ipxpress.Server()))

	fmt.Printf("starting ipxpress server on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
