package ipxpress

import (
	"net/http"
	"strings"
)

// Example custom processors and middlewares for extending IPXpress

// WatermarkProcessor adds a watermark to images.
// Example usage:
//
//	handler.UseProcessor(WatermarkProcessor("watermark.png"))
func WatermarkProcessor(watermarkPath string) ProcessorFunc {
	return func(proc *Processor, params *ProcessingParams) *Processor {
		// Check if watermark is requested via custom parameter
		// You can add custom query params like ?watermark=true
		return proc // Implement watermark logic here
	}
}

// AutoOrientProcessor automatically orients images based on EXIF data.
func AutoOrientProcessor() ProcessorFunc {
	return func(proc *Processor, params *ProcessingParams) *Processor {
		if proc.img != nil {
			_ = proc.img.AutoRotate()
		}
		return proc
	}
}

// StripMetadataProcessor removes all metadata from images for privacy.
func StripMetadataProcessor() ProcessorFunc {
	return func(proc *Processor, params *ProcessingParams) *Processor {
		if proc.img != nil {
			_ = proc.img.RemoveMetadata()
		}
		return proc
	}
}

// CompressionOptimizer optimizes images for web delivery.
func CompressionOptimizer() ProcessorFunc {
	return func(proc *Processor, params *ProcessingParams) *Processor {
		if proc.img != nil {
			// Apply optimal settings for web
			if params.Format == "webp" || params.GetOutputFormat(proc.OriginalFormat()) == FormatWebP {
				// Optimize for WebP
				if params.Quality > 90 {
					params.Quality = 90
				}
			}
		}
		return proc
	}
}

// CORSMiddleware adds CORS headers to responses.
func CORSMiddleware(allowedOrigins []string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs all requests.
func LoggingMiddleware(logger func(string, ...interface{})) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger("request: %s %s", r.Method, r.URL.String())
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware limits requests per client.
func RateLimitMiddleware(maxRequests int) MiddlewareFunc {
	// Simple rate limiter - in production use a proper rate limiting library
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Implement rate limiting logic here
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware validates API keys or tokens.
func AuthMiddleware(validTokens []string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")

			if !contains(validTokens, token) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item || s == "*" {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
