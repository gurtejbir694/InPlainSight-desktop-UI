package analyze

import (
	"image"
	"image/color"
)

// InvertFilter swaps every pixel value to its opposite (255 - value)
func InvertFilter(img image.Image) image.Image {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			// Convert from 16-bit back to 8-bit (0-255)
			dest.Set(x, y, color.RGBA{
				R: uint8(255 - (r >> 8)),
				G: uint8(255 - (g >> 8)),
				B: uint8(255 - (b >> 8)),
				A: uint8(a >> 8),
			})
		}
	}
	return dest
}

// ContrastStretch identifies the darkest and brightest pixels and scales the image
// NewValue = (OldValue - Min) * (255 / (Max - Min))
func ContrastStretch(img image.Image) image.Image {
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)

	var min, max uint8 = 255, 0

	// Pass 1: Find Min/Max Luminance
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			lum := uint8(0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))
			if lum < min {
				min = lum
			}
			if lum > max {
				max = lum
			}
		}
	}

	// Avoid division by zero
	if max == min {
		return img
	}

	scale := 255.0 / float64(max-min)

	// Pass 2: Apply scaling
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			dest.Set(x, y, color.RGBA{
				R: uint8(float64(uint8(r>>8)-min) * scale),
				G: uint8(float64(uint8(g>>8)-min) * scale),
				B: uint8(float64(uint8(b>>8)-min) * scale),
				A: uint8(a >> 8),
			})
		}
	}
	return dest
}
