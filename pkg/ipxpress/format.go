package ipxpress

import "strings"

// Format represents an image format.
type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
	FormatGIF  Format = "gif"
	FormatWebP Format = "webp"
	FormatAVIF Format = "avif"
)

// String returns the string representation of the format.
func (f Format) String() string {
	return string(f)
}

// ContentType returns the MIME content type for the format.
func (f Format) ContentType() string {
	switch f {
	case FormatPNG:
		return "image/png"
	case FormatWebP:
		return "image/webp"
	case FormatGIF:
		return "image/gif"
	case FormatJPEG:
		return "image/jpeg"
	case FormatAVIF:
		return "image/avif"
	default:
		return "application/octet-stream"
	}
}

// IsValid checks if the format is supported.
func (f Format) IsValid() bool {
	switch f {
	case FormatJPEG, FormatPNG, FormatGIF, FormatWebP, FormatAVIF:
		return true
	default:
		return false
	}
}

// ParseFormat parses a format string and returns a Format.
// Returns empty format if not specified or invalid.
func ParseFormat(s string) Format {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	if s == "jpg" {
		s = "jpeg"
	}

	format := Format(s)
	if format.IsValid() {
		return format
	}
	return ""
}

// DetectFormat detects image format from the first bytes of the image data.
func DetectFormat(data []byte) Format {
	if len(data) < 12 {
		return ""
	}

	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return FormatJPEG
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return FormatPNG
	}

	// GIF: "GIF87a" or "GIF89a"
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return FormatGIF
	}

	// WebP: "RIFF....WEBP"
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return FormatWebP
	}

	// AVIF: starts with "....ftypavif" or "....ftypavis" at bytes 4-11
	if len(data) >= 12 {
		if (data[4] == 0x66 && data[5] == 0x74 && data[6] == 0x79 && data[7] == 0x70) &&
			((data[8] == 0x61 && data[9] == 0x76 && data[10] == 0x69 && data[11] == 0x66) ||
				(data[8] == 0x61 && data[9] == 0x76 && data[10] == 0x69 && data[11] == 0x73)) {
			return FormatAVIF
		}
	}

	return ""
}
