package parser

import (
	"io"
	"strconv"
	"strings"
)

type MP3Parser struct{}

func (p *MP3Parser) CanHandle(ext string) bool {
	return strings.ToLower(ext) == ".mp3"
}

func (p *MP3Parser) Parse(r io.Reader) (*ParsedFile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &ParsedFile{
		Type:    "MP3 Audio",
		RawData: data,
		Metadata: map[string]string{
			"size_bytes": strconv.Itoa(len(data)),
		},
	}, nil
}
