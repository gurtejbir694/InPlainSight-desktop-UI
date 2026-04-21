package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"inplainsight-desktop/internal/analyze"
	"inplainsight-desktop/internal/engine"
	"inplainsight-desktop/internal/parser"
	"os"
	"path/filepath"
	"strings"
)

type App struct {
	ctx    context.Context
	engine *engine.StegoEngine
}

type FilterResult struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewApp() *App {
	return &App{
		engine: &engine.StegoEngine{
			Parsers: []parser.FileParser{
				&parser.PNGParser{},
				&parser.JPEGParser{},
				&parser.MP3Parser{},
				&parser.WAVParser{},
			},
			Analyzers: []analyze.StegoAnalyzer{
				&analyze.LSBAnalyzer{},
				&analyze.EntropyAnalyzer{},
				&analyze.HeaderAnalyzer{},
				&analyze.ExifAnalyzer{},
				&analyze.AudioLSBAnalyzer{},
				&analyze.ID3Analyzer{},
			},
		},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Image/Audio file for Analysis",
		Filters: []runtime.FileFilter{
			{DisplayName: "All file supported", Pattern: "*.png;*.jpg;*.jpeg;*.mp3;*.wav"},
			{DisplayName: "Images (*.png;*.jpg)", Pattern: "*.png;*.jpg;*.jpeg"},
			{DisplayName: "Audio (*.mp3;*.wav)", Pattern: "*.mp3;*.wav"},
		},
	})
}

func (a *App) AnalyzeFile(filepath string) (interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return a.engine.Run(filepath, file)
}

// GetBitPlaneImages generates 4 grayscale maps of the image's bit planes
func (a *App) GetBitPlaneImages(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Limit size for the visualizer to keep it fast
	if width > 500 {
		width = 500
	}
	if height > 500 {
		height = 500
	}

	var maps []string

	for plane := uint(0); plane < 4; plane++ {
		resImg := image.NewGray(image.Rect(0, 0, width, height))

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				r, _, _, _ := img.At(x, y).RGBA()
				// Extract bit from Red channel (standard for visual noise mapping)
				bit := uint8((uint8(r>>8) >> plane) & 1)

				if bit == 1 {
					resImg.SetGray(x, y, color.Gray{Y: 255})
				} else {
					resImg.SetGray(x, y, color.Gray{Y: 0})
				}
			}
		}

		var buf bytes.Buffer
		png.Encode(&buf, resImg)
		encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
		maps = append(maps, "data:image/png;base64,"+encoded)
	}

	return maps, nil
}

// RepairAndSave takes a file path, repairs the file, and saves it with a _repaired suffix.
func (a *App) RepairAndSave(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(path))
	fileType := "Unknown"

	// Map extensions to internal types for the Repair logic
	switch ext {
	case ".png":
		fileType = "PNG Image"
	case ".jpg", ".jpeg":
		fileType = "JPEG Image"
	case ".mp3":
		fileType = "MP3 Audio"
	case ".wav":
		fileType = "WAV Audio"
	}

	pf := &parser.ParsedFile{
		Type:    fileType,
		RawData: data,
	}

	repaired, err := a.engine.RepairFile(pf)
	if err != nil {
		return "", err
	}

	newPath := strings.TrimSuffix(path, ext) + "_repaired" + ext
	err = os.WriteFile(newPath, repaired, 0644)
	if err != nil {
		return "", err
	}

	return newPath, nil
}

// imageToBase64 is a private helper that converts an image.Image to a data URI string.
func (a *App) imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", err
	}

	imgBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + imgBase64, nil
}

func (a *App) GetForensicFilters(path string) ([]FilterResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	var results []FilterResult

	// Apply Invert
	inverted := analyze.InvertFilter(img)
	invBase64, _ := a.imageToBase64(inverted)
	results = append(results, FilterResult{Name: "Color Inversion", Data: invBase64})

	// Apply Contrast Stretch
	stretched := analyze.ContrastStretch(img)
	strBase64, _ := a.imageToBase64(stretched)
	results = append(results, FilterResult{Name: "Contrast Stretch (Histogram Equalization)", Data: strBase64})

	return results, nil
}

// GetAudioSpectrogram generates a frequency heatmap for WAV files
// GetAudioSpectrogram generates a frequency heatmap for audio files (WAV or MP3)
func (a *App) GetAudioSpectrogram(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(path))
	var pcmData []byte

	// 1. Handle different audio formats to get raw PCM
	if ext == ".wav" {
		if len(data) > 44 {
			// Skip the 44-byte WAV header for raw PCM samples
			pcmData = data[44:]
		} else {
			return "", fmt.Errorf("invalid WAV file")
		}
	} else if ext == ".mp3" {
		// Decode compressed MP3 into raw PCM using our new analyzer logic
		pcmData, err = analyze.MP3ToPCM(data)
		if err != nil {
			return "", fmt.Errorf("failed to decode MP3: %v", err)
		}
	} else {
		return "", fmt.Errorf("unsupported audio format for spectrogram")
	}

	// 2. Generate the visual spectrogram from the raw PCM
	specBase64, err := analyze.GenerateSpectrogram(pcmData)
	if err != nil {
		return "", err
	}

	// 3. Return as a Data URI for the Svelte <img> tag
	return "data:image/png;base64," + specBase64, nil
}
