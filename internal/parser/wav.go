package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type WAVParser struct{}

func (p *WAVParser) CanHandle(ext string) bool {
	return strings.ToLower(ext) == ".wav"
}

func (p *WAVParser) Parse(r io.Reader) (*ParsedFile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) < 44 {
		return nil, fmt.Errorf("file too small to be a valid WAV")
	}

	// Basic RIFF/WAVE Header check
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a valid RIFF/WAVE container")
	}

	// Extract Sample Rate and Channels (for forensics)
	channels := binary.LittleEndian.Uint16(data[22:24])
	sampleRate := binary.LittleEndian.Uint32(data[24:28])
	bitsPerSample := binary.LittleEndian.Uint16(data[34:36])

	return &ParsedFile{
		Type:    "WAV Audio",
		RawData: data,
		Metadata: map[string]string{
			"size_bytes":      strconv.Itoa(len(data)),
			"channels":        strconv.Itoa(int(channels)),
			"sample_rate":     strconv.Itoa(int(sampleRate)),
			"bits_per_sample": strconv.Itoa(int(bitsPerSample)),
		},
	}, nil
}
