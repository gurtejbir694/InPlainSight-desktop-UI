package analyze

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png" // Required for PNG decoding
	"inplainsight-desktop/internal/models"
	"inplainsight-desktop/internal/parser"
	"os"
	"unicode"
)

type LSBAnalyzer struct{}

func (a *LSBAnalyzer) Name() string {
	return "Least Significant Bit (LSB) Detector"
}

func (a *LSBAnalyzer) Analyze(file *parser.ParsedFile) ([]models.Finding, error) {
	img, _, err := image.Decode(bytes.NewReader(file.RawData))
	if err != nil {
		return nil, nil // Not a valid image for this analyzer
	}

	bounds := img.Bounds()
	var findings []models.Finding

	// We check Bit Planes 0 through 3.
	// Bit 0 is the true LSB, Bit 7 is the Most Significant Bit (MSB).
	for plane := uint(0); plane < 4; plane++ {
		var extractedBytes []byte
		var currentByte byte
		var bitCount int

		// Extracting from RGB channels
	AnalysisLoop:
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()

				for _, val := range []uint32{r, g, b} {
					// Shift and mask to get the specific bit plane
					bit := byte((uint8(val>>8) >> plane) & 1)
					currentByte = (currentByte << 1) | bit
					bitCount++

					if bitCount == 8 {
						extractedBytes = append(extractedBytes, currentByte)
						currentByte = 0
						bitCount = 0
					}
				}
				// Limit sample size per plane for performance
				if len(extractedBytes) >= 20000 {
					break AnalysisLoop
				}
			}
		}

		if len(extractedBytes) == 0 {
			continue
		}

		// Perform analysis on the data from this specific bit plane
		entropy := CalculateEntropy(extractedBytes)
		fileType := detectFileType(extractedBytes)

		// Determine if this specific plane is "Suspicious"
		if entropy > 5.0 || fileType != "Unknown/Binary" {
			confidence, description := getConfidence(entropy, fileType, plane)

			dataPreview := fmt.Sprintf("Entropy: %.4f | Format: %s", entropy, fileType)
			if fileType == "Plaintext" {
				limit := 30
				if len(extractedBytes) < 30 {
					limit = len(extractedBytes)
				}
				dataPreview = fmt.Sprintf("Text Preview: %s...", string(extractedBytes[:limit]))
			}

			findings = append(findings, models.Finding{
				AnalyzerName: a.Name(),
				Description:  description,
				Location:     fmt.Sprintf("Bit Plane: %d", plane),
				DataFound:    dataPreview,
				Confidence:   confidence,
			})
		}
	}

	// If no planes were suspicious, return a single low-confidence baseline finding
	if len(findings) == 0 {
		return []models.Finding{{
			AnalyzerName: a.Name(),
			Description:  "LSB scan complete. No high-entropy patterns found in any bit plane.",
			Confidence:   "Low",
		}}, nil
	}

	return findings, nil
}

// getConfidence calculates the threat level based on data findings
func getConfidence(entropy float64, fType string, plane uint) (string, string) {
	if fType != "Unknown/Binary" && fType != "Plaintext" {
		return "Critical", fmt.Sprintf("Detected %s file header in Bit Plane %d.", fType, plane)
	}
	if entropy > 7.5 {
		return "High", fmt.Sprintf("Extremely high randomness in Bit Plane %d (Likely encrypted).", plane)
	}
	if entropy > 5.0 {
		return "High", fmt.Sprintf("Significant noise cluster detected in Bit Plane %d.", plane)
	}
	if fType == "Plaintext" {
		return "Medium", fmt.Sprintf("Potential plaintext message found in Bit Plane %d.", plane)
	}
	return "Medium", fmt.Sprintf("Unusual entropy levels in Bit Plane %d.", plane)
}

func detectFileType(data []byte) string {
	if len(data) < 4 {
		return "Unknown/Binary"
	}
	switch {
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		return "ZIP/Archive"
	case bytes.HasPrefix(data, []byte("\xFF\xD8\xFF")):
		return "JPEG"
	case bytes.HasPrefix(data, []byte("\x89PNG")):
		return "PNG"
	case bytes.HasPrefix(data, []byte("%PDF")):
		return "PDF"
	case isPrintable(data):
		return "Plaintext"
	default:
		return "Unknown/Binary"
	}
}

func isPrintable(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	printable := 0
	for _, b := range data {
		if unicode.IsPrint(rune(b)) || unicode.IsSpace(rune(b)) {
			printable++
		}
	}
	return float64(printable)/float64(len(data)) > 0.9
}

// Add this to your LSBAnalyzer in internal/analyze/lsb.go
func (a *LSBAnalyzer) dumpExtractedData(filename string, data []byte) error {
	outName := fmt.Sprintf("extracted_%s.bin", filename)
	return os.WriteFile(outName, data, 0644)
}
