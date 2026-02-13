# Extending IPXpress: Using Any libvips Function

## Overview

IPXpress provides several ways to use any libvips function without being limited to built-in methods. This lets you apply arbitrary image transformations.

## Extension methods

### 1. Direct ImageRef Access - `ImageRef()`

Get direct access to the underlying `vips.ImageRef` and call any libvips function.

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

// Access ImageRef and use any libvips function
imgRef := processor.ImageRef()
if imgRef != nil {
    // Call any libvips methods directly
    imgRef.Blur(2.5)
    imgRef.Sharpen(1.5, 0.5, 1.0)
    imgRef.Modulate(1.1, 1.2, 0)
}

result, err := processor.ToBytes(ipxpress.FormatJPEG, 85)
```

### 2. ApplyFunc - Callback functions

Use `ApplyFunc` to run custom processing functions with automatic error handling.

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

// Apply a custom function
processor.ApplyFunc(func(img *vips.ImageRef) error {
    // You can use any libvips functions
    if err := img.Blur(2.0); err != nil {
        return err
    }
    return img.Sharpen(1.5, 0.5, 1.0)
})

if processor.Err() != nil {
    log.Fatal(processor.Err())
}

result, err := processor.ToBytes(ipxpress.FormatWebP, 80)
```

### 3. VipsOperationBuilder - Fluent API

Build an operation chain with a convenient interface and error handling:

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

builder := ipxpress.NewVipsOperationBuilder(processor)
err := builder.
    Blur(2.0).
    Sharpen(1.5, 0.5, 1.0).
    Modulate(1.1, 1.2, 0).
    Median(3).
    Error()

if err != nil {
    log.Fatal(err)
}

result, err := processor.ToBytes(ipxpress.FormatJPEG, 85)
```

### 4. CustomOperation - Custom operations as processors

Create reusable custom operations:

```go
// Define a custom operation
applySepiaEffect := func(p *ipxpress.Processor, params *ipxpress.ProcessingParams) error {
    img := p.ImageRef()
    if img == nil {
        return errors.New("no image loaded")
    }

    // Apply sepia effect
    if err := img.Modulate(1.0, 0.0, 0); err != nil {
        return err
    }
    sepiaColor := &vips.Color{R: 255, G: 200, B: 124}
    return img.Tint(sepiaColor)
}

// Use the operation
processor := ipxpress.New()
processor.FromBytes(imageData)
processor.ApplyCustom(applySepiaEffect, nil)
```

### 5. Processors in the handler

Add custom operations to the request processing pipeline:

```go
config := &ipxpress.Config{ProcessingLimit: 10}
handler := ipxpress.NewHandler(config)

// Add a custom processor with any libvips operation
handler.UseProcessor(func(p *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    return p.ApplyFunc(func(img *vips.ImageRef) error {
        // Apply any libvips operation
        return img.Blur(1.5)
    })
})

// Or use built-in operations
handler.UseProcessor(func(p *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    return p.ApplyFunc(func(img *vips.ImageRef) error {
        // Apply a more complex operation
        if err := img.Sharpen(2.0, 0.5, 1.0); err != nil {
            return err
        }
        return img.Modulate(1.05, 1.1, 0)
    })
})

mux := http.NewServeMux()
mux.Handle("/img/", http.StripPrefix("/img/", handler))
http.ListenAndServe(":8080", mux)
```

## Usage examples

### Example 1: Blur with sharpening

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    ApplyFunc(func(img *vips.ImageRef) error {
        if err := img.Blur(0.5); err != nil {
            return err
        }
        return img.Sharpen(2.0, 0.5, 1.0)
    })

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 90)
```

### Example 2: Sepia effect

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    ApplyFunc(func(img *vips.ImageRef) error {
        // Convert to grayscale
        if err := img.Modulate(1.0, 0.0, 0); err != nil {
            return err
        }
        // Apply sepia tint
        sepiaColor := &vips.Color{R: 255, G: 200, B: 124}
        return img.Tint(sepiaColor)
    })

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 85)
```

### Example 3: Contrast and brightness adjustment

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    ApplyFunc(func(img *vips.ImageRef) error {
        // Increase brightness by 10% and saturation by 20%
        return img.Modulate(1.1, 1.2, 0)
    }).
    ApplyFunc(func(img *vips.ImageRef) error {
        // Increase contrast
        return img.Linear([]float64{1.3}, []float64{0})
    })

result, _ := processor.ToBytes(ipxpress.FormatWebP, 80)
```

### Example 4: Thumbnail with special processing

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    Resize(200, 200).
    ApplyFunc(func(img *vips.ImageRef) error {
        // Blur for the thumbnail
        if err := img.Blur(1.0); err != nil {
            return err
        }
        // Sharpen
        return img.Sharpen(1.5, 0.5, 1.0)
    })

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 75)
```

### Example 5: Using built-in predefined operations

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

// Use built-in factory functions to create operations
processor.ApplyCustom(ipxpress.GaussianBlurOperation(2.5), nil).
    ApplyCustom(ipxpress.SaturationOperation(1.2), nil)

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 85)
```

## Built-in operations

IPXpress provides factory functions for common operations:

- `GaussianBlurOperation(sigma)` - Gaussian blur
- `EdgeDetectionOperation(kernel)` - Edge detection
- `SepiaOperation()` - Sepia effect
- `BrightnessOperation(brightness)` - Brightness adjustment
- `SaturationOperation(saturation)` - Saturation adjustment
- `ContrastOperation(contrast)` - Contrast adjustment

## VipsOperationBuilder - Built-in methods

The builder provides convenient chain methods:

```go
builder := ipxpress.NewVipsOperationBuilder(processor)
err := builder.
    Blur(2.0).                    // Gaussian blur
    Sharpen(1.5, 0.5, 1.0).      // Sharpen
    Modulate(1.1, 1.2, 0).        // Brightness, saturation, hue
    Median(3).                    // Median filter
    Error()

if err != nil {
    log.Fatal(err)
}
```

## Full access to libvips

If you need an operation not included in built-in methods, use `ImageRef()` directly:

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

img := processor.ImageRef()
if img != nil {
    // Full access to all vips.ImageRef methods
    img.Blur(...)
    img.Sharpen(...)
    img.Convolve(...)
    img.Composite(...)
    // etc.
}
```

## Error handling

All methods support chaining and error handling:

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    Resize(800, 600).
    ApplyFunc(func(img *vips.ImageRef) error {
        return img.Blur(2.0)
    })

if processor.Err() != nil {
    log.Printf("Processing error: %v", processor.Err())
    return
}

result, err := processor.ToBytes(ipxpress.FormatJPEG, 85)
if err != nil {
    log.Printf("Encoding error: %v", err)
}
```

## Performance

- All operations run in memory via libvips
- Caching is handled at the Handler level
- Concurrent requests are supported (configured via `ProcessingLimit`)
- Automatic memory management on Processor close

## Resource cleanup

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    Resize(800, 600).
    ApplyFunc(func(img *vips.ImageRef) error {
        return img.Blur(1.5)
    })

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 85)
processor.Close() // Important: free resources
```

