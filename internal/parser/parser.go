package parser

import "io"

type ParsedFile struct {
	Type     string
	RawData  []byte
	Metadata map[string]string
}

type FileParser interface {
	CanHandle(ext string) bool
	Parse(reader io.Reader) (*ParsedFile, error)
}
