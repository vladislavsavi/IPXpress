// Package main demonstrates advanced IPXpress library usage
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/deadpixel/ipxpress/pkg/ipxpress"
)

func main() {
	// Initialize libvips
	vips.Startup(nil)
	defer vips.Shutdown()

	fmt.Println("IPXpress Advanced Examples")
	fmt.Println("==========================\n")

	// Run all examples
	exampleBlurAndSharpen()
	exampleRotateAndFlip()
	exampleCropAndResize()
	exampleColorOperations()
	exampleFormatConversion()
	exampleChainedOperations()
	exampleRemoteImage()
}

// Example: Blur and Sharpen
func exampleBlurAndSharpen() {
	fmt.Println("Example: Blur and Sharpen")

	data, err := os.ReadFile("test.jpg")
	if err != nil {
		log.Printf("Skip: %v\n\n", err)
		return
	}

	// Apply blur
	proc := ipxpress.New().FromBytes(data).Blur(3.0)
	blurred, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
	proc.Close()
	os.WriteFile("output_blurred.jpg", blurred, 0644)

	// Apply sharpen
	proc2 := ipxpress.New().FromBytes(data).Sharpen(1.5, 1.0, 2.0)
	sharpened, _ := proc2.ToBytes(ipxpress.FormatJPEG, 85)
	proc2.Close()
	os.WriteFile("output_sharpened.jpg", sharpened, 0644)

	fmt.Println("✓ Created blurred and sharpened versions\n")
}

// Example: Rotate and Flip
func exampleRotateAndFlip() {
	fmt.Println("Example: Rotate and Flip")

	data, err := os.ReadFile("test.jpg")
	if err != nil {
		log.Printf("Skip: %v\n\n", err)
		return
	}

	// Rotate 90 degrees
	proc := ipxpress.New().FromBytes(data).Rotate(vips.Angle90)
	rotated, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
	proc.Close()
	os.WriteFile("output_rotated.jpg", rotated, 0644)

	// Flip vertically
	proc2 := ipxpress.New().FromBytes(data).Flip()
	flipped, _ := proc2.ToBytes(ipxpress.FormatJPEG, 85)
	proc2.Close()
	os.WriteFile("output_flipped.jpg", flipped, 0644)

	// Flop horizontally
	proc3 := ipxpress.New().FromBytes(data).Flop()
	flopped, _ := proc3.ToBytes(ipxpress.FormatJPEG, 85)
	proc3.Close()
	os.WriteFile("output_flopped.jpg", flopped, 0644)

	fmt.Println("✓ Created rotated, flipped, and flopped versions\n")
}

// Example: Crop and Resize
func exampleCropAndResize() {
	fmt.Println("Example: Crop and Resize")

	data, err := os.ReadFile("test.jpg")
	if err != nil {
		log.Printf("Skip: %v\n\n", err)
		return
	}

	// Extract 400x400 region starting at (100, 100)
	proc := ipxpress.New().
		FromBytes(data).
		Extract(100, 100, 400, 400).
		Resize(200, 200)

	if err := proc.Err(); err != nil {
		log.Printf("Error: %v\n\n", err)
		proc.Close()
		return
	}

	output, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
	proc.Close()
	os.WriteFile("output_cropped.jpg", output, 0644)

	fmt.Println("✓ Cropped and resized image\n")
}

// Example: Color Operations
func exampleColorOperations() {
	fmt.Println("Example: Color Operations")

	data, err := os.ReadFile("test.jpg")
	if err != nil {
		log.Printf("Skip: %v\n\n", err)
		return
	}

	// Grayscale
	proc1 := ipxpress.New().FromBytes(data).Grayscale()
	gray, _ := proc1.ToBytes(ipxpress.FormatJPEG, 85)
	proc1.Close()
	os.WriteFile("output_grayscale.jpg", gray, 0644)

	// Gamma correction
	proc2 := ipxpress.New().FromBytes(data).Gamma(2.2)
	gamma, _ := proc2.ToBytes(ipxpress.FormatJPEG, 85)
	proc2.Close()
	os.WriteFile("output_gamma.jpg", gamma, 0644)

	// Negate (invert colors)
	proc3 := ipxpress.New().FromBytes(data).Negate()
	negated, _ := proc3.ToBytes(ipxpress.FormatJPEG, 85)
	proc3.Close()
	os.WriteFile("output_negate.jpg", negated, 0644)

	// Modulate (HSB adjustment)
	proc4 := ipxpress.New().FromBytes(data).Modulate(1.2, 0.8, 30.0)
	modulated, _ := proc4.ToBytes(ipxpress.FormatJPEG, 85)
	proc4.Close()
	os.WriteFile("output_modulated.jpg", modulated, 0644)

	fmt.Println("✓ Applied various color operations\n")
}

// Example: Format Conversion
func exampleFormatConversion() {
	fmt.Println("Example: Format Conversion")

	data, err := os.ReadFile("test.jpg")
	if err != nil {
		log.Printf("Skip: %v\n\n", err)
		return
	}

	// Convert to WebP
	proc1 := ipxpress.New().FromBytes(data)
	webp, _ := proc1.ToBytes(ipxpress.FormatWebP, 85)
	proc1.Close()
	os.WriteFile("output.webp", webp, 0644)

	// Convert to PNG
	proc2 := ipxpress.New().FromBytes(data)
	png, _ := proc2.ToBytes(ipxpress.FormatPNG, 85)
	proc2.Close()
	os.WriteFile("output.png", png, 0644)

	// Convert to AVIF
	proc3 := ipxpress.New().FromBytes(data)
	avif, _ := proc3.ToBytes(ipxpress.FormatAVIF, 85)
	proc3.Close()
	os.WriteFile("output.avif", avif, 0644)

	fmt.Printf("✓ Converted to WebP (%d bytes), PNG (%d bytes), AVIF (%d bytes)\n\n",
		len(webp), len(png), len(avif))
}

// Example: Chained Operations
func exampleChainedOperations() {
	fmt.Println("Example: Chained Operations")

	data, err := os.ReadFile("test.jpg")
	if err != nil {
		log.Printf("Skip: %v\n\n", err)
		return
	}

	// Chain multiple operations
	proc := ipxpress.New().
		FromBytes(data).
		Resize(800, 600).
		Grayscale().
		Sharpen(1.0, 1.0, 2.0).
		Gamma(1.5)

	if err := proc.Err(); err != nil {
		log.Printf("Error: %v\n\n", err)
		proc.Close()
		return
	}

	output, err := proc.ToBytes(ipxpress.FormatJPEG, 90)
	proc.Close()

	if err != nil {
		log.Printf("Error encoding: %v\n\n", err)
		return
	}

	os.WriteFile("output_chained.jpg", output, 0644)
	fmt.Println("✓ Applied chained operations: resize → grayscale → sharpen → gamma\n")
}

// Example: Process Remote Image
func exampleRemoteImage() {
	fmt.Println("Example: Process Remote Image")

	// Fetch image from URL
	resp, err := http.Get("https://picsum.photos/800/600")
	if err != nil {
		log.Printf("Skip: %v\n\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Skip: HTTP %d\n\n", resp.StatusCode)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error: %v\n\n", err)
		return
	}

	// Process and save
	proc := ipxpress.New().
		FromBytes(data).
		Resize(400, 300).
		Sharpen(1.0, 1.0, 2.0)

	output, err := proc.ToBytes(ipxpress.FormatWebP, 85)
	proc.Close()

	if err != nil {
		log.Printf("Error: %v\n\n", err)
		return
	}

	os.WriteFile("output_remote.webp", output, 0644)
	fmt.Printf("✓ Fetched and processed remote image (%d bytes)\n\n", len(output))
}
