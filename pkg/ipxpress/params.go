package ipxpress

import (
	"net/http"
	"strconv"
)

// ProcessingParams contains all parameters for image processing.
type ProcessingParams struct {
	URL     string
	Width   int
	Height  int
	Quality int
	Format  Format
}

// ParseProcessingParams extracts processing parameters from HTTP request.
func ParseProcessingParams(r *http.Request) *ProcessingParams {
	q := r.URL.Query()

	params := &ProcessingParams{
		URL:     q.Get("url"),
		Width:   parseInt(q.Get("w")),
		Height:  parseInt(q.Get("h")),
		Quality: parseInt(q.Get("quality")),
		Format:  ParseFormat(q.Get("format")),
	}

	// Set default quality if not specified or invalid
	if params.Quality <= 0 || params.Quality > 100 {
		params.Quality = 85
	}

	return params
}

// NeedsProcessing returns true if any transformation is requested.
func (p *ProcessingParams) NeedsProcessing(originalFormat Format) bool {
	return p.Width > 0 || p.Height > 0 || p.Quality != 85 || (p.Format != "" && p.Format != originalFormat)
}

// GetOutputFormat returns the output format, using original format if not specified.
func (p *ProcessingParams) GetOutputFormat(originalFormat Format) Format {
	if p.Format == "" {
		if originalFormat != "" {
			return originalFormat
		}
		return FormatJPEG
	}
	return p.Format
}

// parseInt is a helper function to parse integer from string.
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
