package models

type AnalysisResult struct {
	FileName string            `json:"file_name"`
	FileType string            `json:"file_type"`
	RawData  []byte            `json:"-"`
	Metadata map[string]string `json:"metadata"`
	Findings []Finding         `json:"findings"`
}

type Finding struct {
	AnalyzerName string `json:"analyzer_name"`
	Description  string `json:"description"`
	DataFound    string `json:"data_found,omitempty"`
	Location     string `json:"location,omitempty"`
	Confidence   string `json:"confidence"` // e.g., "High", "Medium"
}
