package ipxpress

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Fetcher is responsible for fetching images from URLs.
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new Fetcher with optimized HTTP client settings.
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 40 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 100,
				MaxConnsPerHost:     256,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 60 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 20 * time.Second,
			},
		},
	}
}

// FetchError represents an error during image fetching.
type FetchError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface.
func (e *FetchError) Error() string {
	return e.Message
}

// Fetch fetches image data from the given URL.
func (f *Fetcher) Fetch(imageURL string) ([]byte, error) {
	if imageURL == "" {
		return nil, &FetchError{
			StatusCode: http.StatusBadRequest,
			Message:    "missing image URL",
		}
	}

	// Validate URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return nil, &FetchError{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("invalid image URL: %v", err),
		}
	}

	if parsedURL.Scheme == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, &FetchError{
			StatusCode: http.StatusBadRequest,
			Message:    "image URL must use http or https",
		}
	}

	// Create request with User-Agent header
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, &FetchError{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("invalid URL: %v", err),
		}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Execute request with simple retries on transient network/DNS errors
	var resp *http.Response
	// reuse existing err variable from above; do not redeclare
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = f.client.Do(req)
		if err == nil {
			break
		}
		// For network errors like timeouts or temporary DNS issues, wait and retry
		if ne, ok := err.(net.Error); ok && (ne.Timeout() || ne.Temporary()) {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}
		// For other errors, no point retrying
		break
	}
	if err != nil {
		return nil, &FetchError{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("failed to fetch image: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &FetchError{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("image fetch failed with status %d", resp.StatusCode),
		}
	}

	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &FetchError{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("failed to read image data: %v", err),
		}
	}

	return imageData, nil
}
