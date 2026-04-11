package analyze

import (
	"bytes"
	"fmt"
	"inplainsight-desktop/internal/models"
	"inplainsight-desktop/internal/parser"
)

type HeaderAnalyzer struct{}

func (a *HeaderAnalyzer) Name() string { return "Header & Structure Analyzer" }

func (a *HeaderAnalyzer) Analyze(file *parser.ParsedFile) ([]models.Finding, error) {
	var findings []models.Finding

	// --- 1. FORMAT-SPECIFIC OVERLAY DETECTION ---
	if file.Type == "PNG Image" {
		// PNG IEND chunk logic
		iendMarker := []byte("\x00\x00\x00\x00IEND\xae\x42\x60\x82")
		iendPos := bytes.Index(file.RawData, iendMarker)

		if iendPos != -1 {
			offset := iendPos + len(iendMarker)
			if offset < len(file.RawData) {
				trailingDataSize := len(file.RawData) - offset
				findings = append(findings, models.Finding{
					AnalyzerName: a.Name(),
					Description:  "Trailing data detected after PNG IEND marker (Overlay).",
					DataFound:    fmt.Sprintf("%d bytes of hidden trailing data", trailingDataSize),
					Location:     fmt.Sprintf("Byte Offset: %d", offset),
					Confidence:   "Critical",
				})
			}
		} else {
			findings = append(findings, models.Finding{
				AnalyzerName: a.Name(),
				Description:  "Standard PNG IEND marker not found. File may be truncated or non-standard.",
				Confidence:   "Medium",
			})
		}

		// Metadata Chunk Check (PNG Specific)
		suspiciousChunks := []string{"tEXt", "zTXt", "iTXt"}
		chunkCount := 0
		for _, chunkName := range suspiciousChunks {
			chunkCount += bytes.Count(file.RawData, []byte(chunkName))
		}
		if chunkCount > 5 {
			findings = append(findings, models.Finding{
				AnalyzerName: a.Name(),
				Description:  "High volume of PNG metadata chunks detected.",
				DataFound:    fmt.Sprintf("Found %d ancillary text chunks.", chunkCount),
				Confidence:   "Medium",
			})
		}

	} else if file.Type == "JPEG Image" {
		// JPEG End of Image (EOI) marker is FF D9
		eoiMarker := []byte{0xFF, 0xD9}
		// Use LastIndex because JPEGs can contain multiple FF D9 pairs in thumbnails,
		// but the real end is usually the last one.
		eoiPos := bytes.LastIndex(file.RawData, eoiMarker)

		if eoiPos != -1 {
			offset := eoiPos + len(eoiMarker)
			if offset < len(file.RawData) {
				trailingDataSize := len(file.RawData) - offset
				findings = append(findings, models.Finding{
					AnalyzerName: a.Name(),
					Description:  "Trailing data detected after JPEG EOI marker (Overlay).",
					DataFound:    fmt.Sprintf("%d bytes of hidden trailing data", trailingDataSize),
					Location:     fmt.Sprintf("Byte Offset: %d", offset),
					Confidence:   "Critical",
				})
			}
		} else {
			findings = append(findings, models.Finding{
				AnalyzerName: a.Name(),
				Description:  "Standard JPEG EOI marker (FF D9) not found. File may be corrupted.",
				Confidence:   "Medium",
			})
		}

		// JPEG Comment Segment (COM) check
		comCount := bytes.Count(file.RawData, []byte{0xFF, 0xFE})
		if comCount > 1 {
			findings = append(findings, models.Finding{
				AnalyzerName: a.Name(),
				Description:  "Multiple JPEG COM (Comment) segments detected.",
				DataFound:    fmt.Sprintf("Found %d comment markers.", comCount),
				Confidence:   "Medium",
			})
		}
	}

	// --- 2. BASELINE REPORT ---
	if len(findings) == 0 {
		findings = append(findings, models.Finding{
			AnalyzerName: a.Name(),
			Description:  fmt.Sprintf("%s structure appears valid. No overlays or suspicious signatures detected.", file.Type),
			DataFound:    "Header structure verified.",
			Confidence:   "Low",
		})
	}

	return findings, nil
}
