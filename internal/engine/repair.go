package engine

import (
	"bytes"
	"fmt"
	"inplainsight-desktop/internal/parser"
)

// RepairFile attempts to fix structural issues based on the file type.
func (e *StegoEngine) RepairFile(file *parser.ParsedFile) ([]byte, error) {
	if file == nil || len(file.RawData) == 0 {
		return nil, fmt.Errorf("no data provided for repair")
	}

	// Work on a copy to avoid mutating the original parsed data
	data := file.RawData

	if file.Type == "PNG Image" {
		// PNG IEND chunk is exactly 12 bytes:
		// Length(4) + Type(4) + CRC(4) -> \x00\x00\x00\x00IEND\xae\x42\x60\x82
		iendMarker := []byte("\x00\x00\x00\x00IEND\xae\x42\x60\x82")
		iendPos := bytes.Index(data, iendMarker)

		if iendPos == -1 {
			// CASE A: Marker is missing (File is truncated)
			// Action: Append the missing marker to make it a valid PNG again.
			return append(data, iendMarker...), nil
		} else {
			// CASE B: Marker exists, but is there data AFTER it? (Overlay/Trailing Data)
			// Action: Truncate the file to end exactly after the 12-byte IEND chunk.
			endOfFile := iendPos + len(iendMarker)
			if endOfFile < len(data) {
				return data[:endOfFile], nil
			}
		}

	} else if file.Type == "JPEG Image" {
		eoiMarker := []byte{0xFF, 0xD9}
		// Use LastIndex because JPEGs can have multiple EOI markers in thumbnails,
		// but the real end is the last one.
		eoiPos := bytes.LastIndex(data, eoiMarker)

		if eoiPos == -1 {
			// CASE A: EOI is missing
			return append(data, eoiMarker...), nil
		} else {
			// CASE B: Data exists after the last EOI marker
			endOfFile := eoiPos + len(eoiMarker)
			if endOfFile < len(data) {
				return data[:endOfFile], nil
			}
		}
	}

	// If no repairs were needed, return original data
	return data, nil
}
