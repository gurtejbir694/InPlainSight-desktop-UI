package analyze

import (
	"fmt"
	"inplainsight-desktop/internal/models"
	"inplainsight-desktop/internal/parser"
	"math"
)

type AudioLSBAnalyzer struct{}

func (a *AudioLSBAnalyzer) Name() string {
	return "Audio LSB Bit-Plane Analyzer"
}

func (a *AudioLSBAnalyzer) Analyze(file *parser.ParsedFile) ([]models.Finding, error) {
	if file.Type != "WAV Audio" {
		return nil, nil
	}

	pcmData := file.RawData[44:]
	var bit0Count int
	for _, b := range pcmData {
		if b&1 == 1 {
			bit0Count++
		}
	}

	// Calculate probability of bit being 1
	p1 := float64(bit0Count) / float64(len(pcmData))
	p0 := 1.0 - p1

	// Calculate Entropy: H = -(p0*log2(p0) + p1*log2(p1))
	entropy := 0.0
	if p0 > 0 && p1 > 0 {
		entropy = -(p0*math.Log2(p0) + p1*math.Log2(p1))
	}

	var findings []models.Finding
	// In a perfect random bit-plane (stego), entropy will be very close to 1.0
	if entropy > 0.99 {
		findings = append(findings, models.Finding{
			AnalyzerName: a.Name(),
			Description:  "Bit-Plane 0 shows near-perfect entropy.",
			DataFound:    fmt.Sprintf("Entropy: %.4f (Theoretical Max: 1.0)", entropy),
			Confidence:   "High",
		})
	}

	return findings, nil
}
