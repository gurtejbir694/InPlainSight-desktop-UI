package analyze

import (
	"bytes"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"inplainsight-desktop/internal/models"
	"inplainsight-desktop/internal/parser"
)

type ExifAnalyzer struct{}

func (a *ExifAnalyzer) Name() string { return "EXIF Metadata Investigator" }

func (a *ExifAnalyzer) Analyze(file *parser.ParsedFile) ([]models.Finding, error) {
	var findings []models.Finding

	// Attempt to decode EXIF
	x, err := exif.Decode(bytes.NewReader(file.RawData))
	if err != nil {
		// Instead of returning nothing, let's return a 'Low' confidence finding
		// so the user knows the scan happened.
		findings = append(findings, models.Finding{
			AnalyzerName: a.Name(),
			Description:  "No EXIF metadata block found in this image.",
			DataFound:    "Metadata stripped or not present.",
			Confidence:   "Low",
		})
		return findings, nil
	}

	// 1. Check for Hardware/Software Signatures
	camModel, _ := x.Get(exif.Model)
	software, _ := x.Get(exif.Software)

	if camModel != nil || software != nil {
		data := ""
		if camModel != nil {
			data += fmt.Sprintf("Device: %s ", camModel)
		}
		if software != nil {
			data += fmt.Sprintf("Software: %s", software)
		}

		findings = append(findings, models.Finding{
			AnalyzerName: a.Name(),
			Description:  "Hardware/Software fingerprints detected.",
			DataFound:    data,
			Confidence:   "Medium",
		})
	}

	// 2. Check for GPS
	lat, long, err := x.LatLong()
	if err == nil {
		findings = append(findings, models.Finding{
			AnalyzerName: a.Name(),
			Description:  "Geospatial (GPS) coordinates identified.",
			DataFound:    fmt.Sprintf("Lat: %f, Long: %f", lat, long),
			Location:     "EXIF GPS Segment",
			Confidence:   "High",
		})
	}

	return findings, nil
}
