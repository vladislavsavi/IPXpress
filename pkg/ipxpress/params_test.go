package ipxpress

import (
	"net/http"
	"testing"
)

// TestShortParameterAliases tests that short parameter names work correctly
func TestShortParameterAliases(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected ProcessingParams
	}{
		{
			name:  "Short width parameter",
			query: "?url=https://example.com/test.jpg&w=800",
			expected: ProcessingParams{
				URL:     "https://example.com/test.jpg",
				Width:   800,
				Quality: 85,
			},
		},
		{
			name:  "Short height parameter",
			query: "?url=https://example.com/test.jpg&h=600",
			expected: ProcessingParams{
				URL:     "https://example.com/test.jpg",
				Height:  600,
				Quality: 85,
			},
		},
		{
			name:  "Short format parameter",
			query: "?url=https://example.com/test.jpg&f=webp",
			expected: ProcessingParams{
				URL:     "https://example.com/test.jpg",
				Format:  FormatWebP,
				Quality: 85,
			},
		},
		{
			name:  "Short quality parameter",
			query: "?url=https://example.com/test.jpg&q=90",
			expected: ProcessingParams{
				URL:     "https://example.com/test.jpg",
				Quality: 90,
			},
		},
		{
			name:  "Short background parameter",
			query: "?url=https://example.com/test.jpg&b=ff0000",
			expected: ProcessingParams{
				URL:        "https://example.com/test.jpg",
				Background: "#ff0000",
				Quality:    85,
			},
		},
		{
			name:  "Short position parameter",
			query: "?url=https://example.com/test.jpg&pos=center",
			expected: ProcessingParams{
				URL:      "https://example.com/test.jpg",
				Position: "center",
				Quality:  85,
			},
		},
		{
			name:  "Resize parameter (s)",
			query: "?url=https://example.com/test.jpg&s=800x600",
			expected: ProcessingParams{
				URL:     "https://example.com/test.jpg",
				Width:   800,
				Height:  600,
				Quality: 85,
			},
		},
		{
			name:  "Long parameters",
			query: "?url=https://example.com/test.jpg&width=1200&height=800&format=png&quality=95",
			expected: ProcessingParams{
				URL:     "https://example.com/test.jpg",
				Width:   1200,
				Height:  800,
				Format:  FormatPNG,
				Quality: 95,
			},
		},
		{
			name:  "Mixed short and long parameters",
			query: "?url=https://example.com/test.jpg&w=1000&height=500&f=jpeg&quality=80",
			expected: ProcessingParams{
				URL:     "https://example.com/test.jpg",
				Width:   1000,
				Height:  500,
				Format:  FormatJPEG,
				Quality: 80,
			},
		},
		{
			name:  "Short parameters override resize",
			query: "?url=https://example.com/test.jpg&s=400x300&w=800",
			expected: ProcessingParams{
				URL:     "https://example.com/test.jpg",
				Width:   800,
				Height:  300,
				Quality: 85,
			},
		},
		{
			name:  "All short parameters combined",
			query: "?url=https://example.com/test.jpg&w=1000&h=600&f=webp&q=85&b=ffffff&pos=top",
			expected: ProcessingParams{
				URL:        "https://example.com/test.jpg",
				Width:      1000,
				Height:     600,
				Format:     FormatWebP,
				Quality:    85,
				Background: "#ffffff",
				Position:   "top",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://localhost"+tt.query, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			params := ParseProcessingParams(req)

			if params.URL != tt.expected.URL {
				t.Errorf("URL: got %q, want %q", params.URL, tt.expected.URL)
			}
			if params.Width != tt.expected.Width {
				t.Errorf("Width: got %d, want %d", params.Width, tt.expected.Width)
			}
			if params.Height != tt.expected.Height {
				t.Errorf("Height: got %d, want %d", params.Height, tt.expected.Height)
			}
			if params.Quality != tt.expected.Quality {
				t.Errorf("Quality: got %d, want %d", params.Quality, tt.expected.Quality)
			}
			if params.Format != tt.expected.Format {
				t.Errorf("Format: got %q, want %q", params.Format, tt.expected.Format)
			}
			if params.Background != tt.expected.Background {
				t.Errorf("Background: got %q, want %q", params.Background, tt.expected.Background)
			}
			if params.Position != tt.expected.Position {
				t.Errorf("Position: got %q, want %q", params.Position, tt.expected.Position)
			}
		})
	}
}

// TestResizeParameter tests the s=WIDTHxHEIGHT parameter format
func TestResizeParameter(t *testing.T) {
	tests := []struct {
		name           string
		resizeValue    string
		expectedWidth  int
		expectedHeight int
	}{
		{"Valid resize 800x600", "800x600", 800, 600},
		{"Valid resize 1920x1080", "1920x1080", 1920, 1080},
		{"Valid resize 100x100", "100x100", 100, 100},
		{"Invalid format (missing x)", "800", 0, 0},
		{"Invalid format (extra parts)", "800x600x400", 0, 0}, // Should fail to parse, both 0
		{"Non-numeric values", "widthxheight", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost?url=test.jpg&s="+tt.resizeValue, nil)
			params := ParseProcessingParams(req)

			if params.Width != tt.expectedWidth {
				t.Errorf("Width: got %d, want %d", params.Width, tt.expectedWidth)
			}
			if params.Height != tt.expectedHeight {
				t.Errorf("Height: got %d, want %d", params.Height, tt.expectedHeight)
			}
		})
	}
}
