package ipxpress

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

// ProcessingParams contains all parameters for image processing.
type ProcessingParams struct {
	URL     string
	Width   int
	Height  int
	Quality int
	Format  Format

	// Resize options
	Fit      string // contain, cover, fill, inside, outside
	Position string // top, bottom, left, right, centre, etc.
	Kernel   string // nearest, cubic, mitchell, lanczos2, lanczos3
	Enlarge  bool   // allow upscaling

	// Operations
	Blur      float64 // blur sigma
	Sharpen   string  // sigma_flat_jagged (e.g., "1.5_1_2")
	Rotate    int     // rotation angle
	Flip      bool    // flip vertically
	Flop      bool    // flip horizontally
	Grayscale bool    // convert to grayscale

	// Cropping and extending
	Extract string // left_top_width_height
	Trim    int    // trim threshold
	Extend  string // top_right_bottom_left

	// Color operations
	Background string  // background color (hex)
	Negate     bool    // invert colors
	Normalize  bool    // normalize image
	Threshold  int     // threshold value
	Tint       string  // tint color (hex)
	Gamma      float64 // gamma correction
	Median     int     // median filter size
	Modulate   string  // brightness_saturation_hue
	Flatten    bool    // remove alpha channel
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

		// Resize options
		Fit:      q.Get("fit"),
		Position: q.Get("position"),
		Kernel:   q.Get("kernel"),
		Enlarge:  parseBool(q.Get("enlarge")),

		// Operations
		Blur:      parseFloat(q.Get("blur")),
		Sharpen:   q.Get("sharpen"),
		Rotate:    parseInt(q.Get("rotate")),
		Flip:      parseBool(q.Get("flip")),
		Flop:      parseBool(q.Get("flop")),
		Grayscale: parseBool(q.Get("grayscale")),

		// Cropping and extending
		Extract: q.Get("extract"),
		Trim:    parseInt(q.Get("trim")),
		Extend:  q.Get("extend"),

		// Color operations
		Background: q.Get("background"),
		Negate:     parseBool(q.Get("negate")),
		Normalize:  parseBool(q.Get("normalize")),
		Threshold:  parseInt(q.Get("threshold")),
		Tint:       q.Get("tint"),
		Gamma:      parseFloat(q.Get("gamma")),
		Median:     parseInt(q.Get("median")),
		Modulate:   q.Get("modulate"),
		Flatten:    parseBool(q.Get("flatten")),
	}

	// Set default quality if not specified or invalid
	if params.Quality <= 0 || params.Quality > 100 {
		params.Quality = 85
	}

	// Normalize background color
	if params.Background != "" {
		params.Background = normalizeHexColor(params.Background)
	}

	// Normalize tint color
	if params.Tint != "" {
		params.Tint = normalizeHexColor(params.Tint)
	}

	return params
}

// NeedsProcessing returns true if any transformation is requested.
func (p *ProcessingParams) NeedsProcessing(originalFormat Format) bool {
	return p.Width > 0 || p.Height > 0 || p.Quality != 85 ||
		(p.Format != "" && p.Format != originalFormat) ||
		p.Blur > 0 || p.Sharpen != "" || p.Rotate != 0 ||
		p.Flip || p.Flop || p.Grayscale ||
		p.Extract != "" || p.Trim > 0 || p.Extend != "" ||
		p.Background != "" || p.Negate || p.Normalize ||
		p.Threshold > 0 || p.Tint != "" || p.Gamma > 0 ||
		p.Median > 0 || p.Modulate != "" || p.Flatten ||
		p.Fit != "" || p.Position != "" || p.Kernel != "" || p.Enlarge
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

// GetVipsKernel converts kernel string to vips.Kernel
func (p *ProcessingParams) GetVipsKernel() vips.Kernel {
	switch strings.ToLower(p.Kernel) {
	case "nearest":
		return vips.KernelNearest
	case "cubic":
		return vips.KernelCubic
	case "mitchell":
		return vips.KernelMitchell
	case "lanczos2":
		return vips.KernelLanczos2
	case "lanczos3":
		return vips.KernelLanczos3
	default:
		return vips.KernelLanczos3 // default
	}
}

// GetVipsInteresting converts position string to vips.Interesting
func (p *ProcessingParams) GetVipsInteresting() vips.Interesting {
	switch strings.ToLower(p.Position) {
	case "centre", "center":
		return vips.InterestingCentre
	case "entropy":
		return vips.InterestingEntropy
	case "attention":
		return vips.InterestingAttention
	case "low":
		return vips.InterestingLow
	case "high":
		return vips.InterestingHigh
	default:
		return vips.InterestingNone
	}
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

// parseFloat is a helper function to parse float from string.
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// parseBool is a helper function to parse boolean from string.
func parseBool(s string) bool {
	if s == "" {
		return false
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return s == "1" || strings.ToLower(s) == "true"
	}
	return v
}

// normalizeHexColor normalizes hex color string
func normalizeHexColor(color string) string {
	// Remove # if present
	color = strings.TrimPrefix(color, "#")

	// Validate hex format
	if len(color) == 3 || len(color) == 6 {
		return "#" + color
	}

	return color
}
