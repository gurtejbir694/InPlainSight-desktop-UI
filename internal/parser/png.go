package parser

import (
	"io"
	"strconv" // Used to convert integers to strings correctly
	"strings"
)

type PNGParser struct{}

func (p *PNGParser) CanHandle(ext string) bool {
	return strings.ToLower(ext) == ".png"
}

func (p *PNGParser) Parse(reader io.Reader) (*ParsedFile, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Simple metadata extraction
	metadata := make(map[string]string)
	// Correctly convert file size to string
	metadata["size_bytes"] = strconv.Itoa(len(data))

	return &ParsedFile{
		Type:     "PNG Image",
		RawData:  data,
		Metadata: metadata,
	}, nil
}
