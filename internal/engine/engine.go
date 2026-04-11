package engine

import (
	"fmt"
	"inplainsight-desktop/internal/analyze"
	"inplainsight-desktop/internal/models"
	"inplainsight-desktop/internal/parser"
	"io"
	"path/filepath"
)

type StegoEngine struct {
	Parsers   []parser.FileParser
	Analyzers []analyze.StegoAnalyzer
}

func (e *StegoEngine) Run(filename string, input io.Reader) (*models.AnalysisResult, error) {
	ext := filepath.Ext(filename)
	var activeParser parser.FileParser

	// Find suitable parser
	for _, p := range e.Parsers {
		if p.CanHandle(ext) {
			activeParser = p
			break
		}
	}

	if activeParser == nil {
		return nil, fmt.Errorf("no parser found for extension: %s", ext)
	}

	// Parse file
	parsed, err := activeParser.Parse(input)
	if err != nil {
		return nil, err
	}

	// Run applicable analyzers
	result := &models.AnalysisResult{
		FileName: filename,
		FileType: parsed.Type,
		Metadata: parsed.Metadata,
	}

	for _, a := range e.Analyzers {
		findings, err := a.Analyze(parsed)
		if err == nil {
			result.Findings = append(result.Findings, findings...)
		}
	}

	return result, nil
}
