package analyze

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/madelynnblue/go-dsp/fft"
)

// GenerateSpectrogram takes raw PCM data (skipping WAV header)
// and returns a base64 encoded PNG of the frequency distribution.
func GenerateSpectrogram(pcmData []byte) (string, error) {
	// Convert 16-bit PCM bytes to float64 samples
	samples := make([]float64, len(pcmData)/2)
	for i := 0; i < len(samples); i++ {
		low := int16(pcmData[i*2])
		high := int16(pcmData[i*2+1])
		val := (high << 8) | low
		samples[i] = float64(val) / 32768.0
	}

	windowSize := 1024
	overlap := 512
	numWindows := (len(samples) - windowSize) / overlap
	if numWindows <= 0 {
		return "", nil
	}

	// X = Time, Y = Frequency (Nyquist limit is windowSize/2)
	img := image.NewRGBA(image.Rect(0, 0, numWindows, windowSize/2))

	for w := 0; w < numWindows; w++ {
		start := w * overlap
		window := samples[start : start+windowSize]

		// Perform FFT (Fast Fourier Transform)
		coeffs := fft.FFTReal(window)

		for f := 0; f < windowSize/2; f++ {
			// Calculate intensity (magnitude)
			mag := math.Sqrt(real(coeffs[f])*real(coeffs[f]) + imag(coeffs[f])*imag(coeffs[f]))
			// Logarithmic scaling for better visibility of faint signals
			intensity := uint8(math.Min(255, 20*math.Log10(mag+1)*10))

			// Forensic Heatmap (Cyan for low intensity, Red for high)
			c := color.RGBA{
				R: intensity,
				G: intensity / 2,
				B: 255 - intensity,
				A: 255,
			}
			img.Set(w, (windowSize/2)-1-f, c)
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
