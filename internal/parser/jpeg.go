package parser

import (
	"io"
	"strconv"
	"strings"
)

type JPEGParser struct{}

// CanHandle matches the interface in parser.go by checking the file extension
func (p *JPEGParser) CanHandle(ext string) bool {
	cleanedExt := strings.ToLower(ext)
	return cleanedExt == ".jpg" || cleanedExt == ".jpeg"
}

func (p *JPEGParser) Parse(r io.Reader) (*ParsedFile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Returning a ParsedFile struct that matches your parser.go exactly
	return &ParsedFile{
		Type:    "JPEG Image",
		RawData: data,
		Metadata: map[string]string{
			"size_bytes": strconv.Itoa(len(data)),
		},
	}, nil
}
