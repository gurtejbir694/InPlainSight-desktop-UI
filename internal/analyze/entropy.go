package analyze

import (
	"fmt"
	"inplainsight-desktop/internal/models"
	"inplainsight-desktop/internal/parser"
	"math"
)

// CalculateEntropy is a shared helper for the whole analyze package
func CalculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}

	counts := make(map[byte]float64)
	for _, b := range data {
		counts[b]++
	}

	var entropy float64
	fileSize := float64(len(data))
	for _, count := range counts {
		p := count / fileSize
		entropy -= p * math.Log2(p)
	}
	return entropy
}

type EntropyAnalyzer struct{}

func (a *EntropyAnalyzer) Name() string { return "Shannon Entropy Analysis" }

func (a *EntropyAnalyzer) Analyze(file *parser.ParsedFile) ([]models.Finding, error) {
	entropy := CalculateEntropy(file.RawData)

	confidence := "Low"
	if entropy > 7.5 {
		confidence = "High"
	}

	return []models.Finding{{
		AnalyzerName: a.Name(),
		Description:  "Calculates overall data randomness.",
		DataFound:    fmt.Sprintf("Global Entropy: %.4f", entropy),
		Confidence:   confidence,
	}}, nil
}
