package analyze

import (
	"bytes"
	"fmt"
	"inplainsight-desktop/internal/models"
	"inplainsight-desktop/internal/parser"
	"strings"

	"github.com/bogem/id3v2/v2"
)

type ID3Analyzer struct{}

// This is the missing piece!
func (a *ID3Analyzer) Name() string {
	return "ID3 Tag Investigator"
}

func (a *ID3Analyzer) Analyze(file *parser.ParsedFile) ([]models.Finding, error) {
	if file.Type != "MP3 Audio" {
		return nil, nil
	}

	var findings []models.Finding

	// We use a reader because the library needs an io.Reader
	tag, err := id3v2.ParseReader(bytes.NewReader(file.RawData), id3v2.Options{Parse: true})
	if err != nil {
		// If the file is an MP3 but has no ID3 tags, it's not an error, just no findings
		return nil, nil
	}
	defer tag.Close()

	// 1. Check for suspicious Comments (COMM frames)
	comments := tag.GetFrames(tag.CommonID("Comments"))
	for _, f := range comments {
		cf, ok := f.(id3v2.CommentFrame)
		if ok && len(cf.Text) > 100 {
			findings = append(findings, models.Finding{
				AnalyzerName: a.Name(), // Use the method here
				Description:  "Suspiciously long comment frame detected.",
				DataFound:    fmt.Sprintf("Length: %d | Start: %s", len(cf.Text), cf.Text[:30]),
				Confidence:   "Medium",
			})
		}
	}

	// 2. Scan for User Defined Text (TXXX frames)
	for id, frames := range tag.AllFrames() {
		if strings.HasPrefix(id, "T") {
			for _, f := range frames {
				findings = append(findings, models.Finding{
					AnalyzerName: a.Name(),
					Description:  "Metadata Frame: " + id,
					DataFound:    fmt.Sprintf("%v", f),
					Confidence:   "Low",
				})
			}
		}
	}

	return findings, nil
}
