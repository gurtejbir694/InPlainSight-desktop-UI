package engine

import (
	"bytes"
	"fmt"
	"github.com/bogem/id3v2/v2"
	"inplainsight-desktop/internal/parser"
	"io"
)

// RepairFile attempts to fix structural issues or sanitize data based on the file type.
func (e *StegoEngine) RepairFile(file *parser.ParsedFile) ([]byte, error) {
	if file == nil || len(file.RawData) == 0 {
		return nil, fmt.Errorf("no data provided for repair")
	}

	switch file.Type {
	case "PNG Image":
		return e.repairPNG(file.RawData)
	case "JPEG Image":
		return e.repairJPEG(file.RawData)
	case "WAV Audio":
		return e.sanitizeWAV(file.RawData)
	case "MP3 Audio":
		return e.sanitizeMP3(file.RawData)
	default:
		// If no specific repair logic, return original data
		return file.RawData, nil
	}
}

// sanitizeWAV zeroes out the LSB of every PCM sample to destroy hidden payloads
func (e *StegoEngine) sanitizeWAV(data []byte) ([]byte, error) {
	if len(data) < 44 {
		return nil, fmt.Errorf("invalid WAV file")
	}

	sanitized := make([]byte, len(data))
	copy(sanitized, data)

	// Skip 44-byte RIFF header, wipe Bit 0 of every byte thereafter
	for i := 44; i < len(sanitized); i++ {
		sanitized[i] &= 254 // 11111110
	}
	return sanitized, nil
}

// sanitizeMP3 strips all metadata frames where out-of-band data is often hidden
func (e *StegoEngine) sanitizeMP3(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	tag, err := id3v2.ParseReader(reader, id3v2.Options{Parse: true})
	if err != nil {
		return nil, err
	}
	defer tag.Close()

	// Clear all frames (Comments, Custom Text, etc.)
	tag.DeleteAllFrames()

	var buf bytes.Buffer
	_, err = tag.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *StegoEngine) repairPNG(data []byte) ([]byte, error) {
	iendMarker := []byte("\x00\x00\x00\x00IEND\xae\x42\x60\x82")
	iendPos := bytes.Index(data, iendMarker)

	if iendPos == -1 {
		return append(data, iendMarker...), nil
	}
	endOfFile := iendPos + len(iendMarker)
	if endOfFile < len(data) {
		return data[:endOfFile], nil
	}
	return data, nil
}

func (e *StegoEngine) repairJPEG(data []byte) ([]byte, error) {
	eoiMarker := []byte{0xFF, 0xD9}
	eoiPos := bytes.LastIndex(data, eoiMarker)

	if eoiPos == -1 {
		return append(data, eoiMarker...), nil
	}
	endOfFile := eoiPos + len(eoiMarker)
	if endOfFile < len(data) {
		return data[:endOfFile], nil
	}
	return data, nil
}
