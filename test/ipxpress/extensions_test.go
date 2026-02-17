package ipxpress_test

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

// TestImageRefAccess verifies direct ImageRef access works
func TestImageRefAccess(t *testing.T) {
	proc := ipxpress.New()

	imgRef := proc.ImageRef()
	if imgRef != nil {
		t.Errorf("expected nil ImageRef before loading image, got non-nil")
	}

	// After loading an image, it should work
	// (In real test you'd need actual image data)
}

// TestApplyFunc verifies ApplyFunc works with proper error handling
func TestApplyFunc(t *testing.T) {
	proc := ipxpress.New()

	// Test with nil function
	result := proc.ApplyFunc(nil)
	if result.Err() == nil {
		t.Error("expected error when ApplyFunc receives nil function")
	}

	// Test with no image
	proc2 := ipxpress.New()
	result2 := proc2.ApplyFunc(func(img *vips.ImageRef) error {
		return nil
	})
	if result2.Err() == nil {
		t.Error("expected error when ApplyFunc called without image")
	}
}

// TestApplyCustom verifies custom operations work
func TestApplyCustom(t *testing.T) {
	proc := ipxpress.New()

	// Test with nil operation
	result := proc.ApplyCustom(nil, nil)
	if result.Err() == nil {
		t.Error("expected error when ApplyCustom receives nil operation")
	}
}

// TestVipsOperationBuilder verifies builder pattern works
func TestVipsOperationBuilder(t *testing.T) {
	proc := ipxpress.New()
	builder := ipxpress.NewVipsOperationBuilder(proc)

	// Should handle nil image gracefully
	err := builder.
		Blur(2.0).
		Error()

	if err == nil {
		t.Error("expected error when operating on nil image")
	}
}

// TestBuilderChaining verifies method chaining works
func TestBuilderChaining(t *testing.T) {
	proc := ipxpress.New()
	builder := ipxpress.NewVipsOperationBuilder(proc)

	// Should be able to chain methods
	result := builder.
		Blur(2.0).
		Sharpen(1.5, 0.5, 1.0).
		Modulate(1.0, 1.0, 0)

	// Should return builder for chaining
	if result == nil {
		t.Error("expected builder to return self for chaining")
	}
}

// TestPredefinedOperations verifies predefined operations exist
func TestPredefinedOperations(t *testing.T) {
	ops := []struct {
		name string
		op   ipxpress.CustomOperation
	}{
		{"GaussianBlur", ipxpress.GaussianBlurOperation(2.0)},
		{"Sepia", ipxpress.SepiaOperation()},
		{"Brightness", ipxpress.BrightnessOperation(1.1)},
		{"Saturation", ipxpress.SaturationOperation(1.2)},
		{"Contrast", ipxpress.ContrastOperation(1.1)},
	}

	for _, tt := range ops {
		t.Run(tt.name, func(t *testing.T) {
			if tt.op == nil {
				t.Errorf("%s operation is nil", tt.name)
			}
		})
	}
}
