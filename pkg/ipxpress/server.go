package ipxpress

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Server returns an http.Handler that processes images from URLs and applies
// transformations using the ipxpress Processor.
// Expected query parameters:
// - url: the URL of the image to process (required)
// - w: maximum width
// - h: maximum height
// - quality: output quality (for JPEG)
// - format: output format (jpeg, png) - defaults to jpeg
func Server() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parse query params
		q := r.URL.Query()
		imageURL := q.Get("url")
		if imageURL == "" {
			http.Error(w, "missing image URL", http.StatusBadRequest)
			return
		}

		// validate URL
		parsedURL, err := url.Parse(imageURL)
		if err != nil {
			http.Error(w, "invalid image URL", http.StatusBadRequest)
			return
		}
		if parsedURL.Scheme == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			http.Error(w, "image URL must use http or https", http.StatusBadRequest)
			return
		}

		// fetch image from URL with User-Agent header
		req, err := http.NewRequest("GET", imageURL, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid URL: %v", err), http.StatusBadRequest)
			return
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch image: %v", err), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("image fetch failed with status %d", resp.StatusCode), http.StatusBadRequest)
			return
		}

		// read image data
		imageData, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read image data: %v", err), http.StatusInternalServerError)
			return
		}

		wv := parseInt(q.Get("w"))
		hv := parseInt(q.Get("h"))
		quality := parseInt(q.Get("quality"))
		formatParam := q.Get("format")

		// If no transformation parameters are specified, return original image
		if wv == 0 && hv == 0 && quality == 0 && formatParam == "" {
			// Return original without processing
			proc := New().FromBytes(imageData)
			format := proc.OriginalFormat()
			if format == "" {
				format = "jpeg"
			}

			var ctype string
			switch format {
			case "png":
				ctype = "image/png"
			case "webp":
				ctype = "image/webp"
			case "gif":
				ctype = "image/gif"
			default:
				ctype = "image/jpeg"
			}

			w.Header().Set("Content-Type", ctype)
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(imageData)
			return
		}

		proc := New().FromBytes(imageData)

		origFormat := proc.OriginalFormat()

		// Determine format: use specified format, or original format as fallback
		format := formatParam
		if format == "" {
			// Use original format if no format specified
			format = origFormat
			if format == "" {
				format = "jpeg" // fallback
			}
		}

		// normalize format
		format = strings.ToLower(format)
		if format == "jpg" {
			format = "jpeg"
		}
		// Validate format - allow jpeg, png, gif, webp
		if format != "jpeg" && format != "png" && format != "gif" && format != "webp" {
			format = "jpeg"
		}

		// Set quality: if not specified, use high quality for optimization without visible loss
		if quality <= 0 || quality > 100 {
			if wv == 0 && hv == 0 {
				// No resize - use high quality to preserve original as much as possible
				quality = 90
			} else {
				// Resize applied - use good quality
				quality = 85
			}
		}

		// Note: quality is validated to not exceed original size below

		// Apply resize only if dimensions are specified
		if wv > 0 || hv > 0 {
			proc = proc.Resize(wv, hv)
		}

		if err := proc.Err(); err != nil {
			http.Error(w, fmt.Sprintf("processing: %v", err), http.StatusInternalServerError)
			return
		}

		originalSize := proc.OriginalSize()
		out, err := proc.ToBytes(format, quality)
		if err != nil {
			http.Error(w, fmt.Sprintf("encode: %v", err), http.StatusInternalServerError)
			return
		}

		// If no resize, try to ensure the output is not larger than the original
		// by gradually reducing quality for lossy formats (jpeg/webp).
		if wv == 0 && hv == 0 && (format == "jpeg" || format == "webp") && len(out) > originalSize {
			tryQuality := quality
			best := out
			for tryQuality > 50 {
				tryQuality -= 5
				tmp, err := proc.ToBytes(format, tryQuality)
				if err != nil {
					break
				}
				if len(tmp) <= originalSize {
					best = tmp
					break
				}
				best = tmp
			}

			// If still larger than original and format not forced, return original
			if len(best) > originalSize {
				if formatParam == "" || format == origFormat {
					out = imageData
				} else {
					out = best
				}
			} else {
				out = best
			}
		}

		// content type
		var ctype string
		switch format {
		case "png":
			ctype = "image/png"
		case "webp":
			ctype = "image/webp"
		case "gif":
			ctype = "image/gif"
		default:
			ctype = "image/jpeg"
		}

		// caching headers
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
	})
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}
