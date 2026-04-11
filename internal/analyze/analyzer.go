package analyze

import (
	"inplainsight-desktop/internal/models"
	"inplainsight-desktop/internal/parser"
)

type StegoAnalyzer interface {
	Name() string
	Analyze(file *parser.ParsedFile) ([]models.Finding, error)
}
