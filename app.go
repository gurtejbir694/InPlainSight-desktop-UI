package main

import (
	"bytes"
	"context"
	"encoding/base64"
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

	"github.com/wailsapp/wails/v2/pkg/runtime"
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
			},
			Analyzers: []analyze.StegoAnalyzer{
				&analyze.LSBAnalyzer{},
				&analyze.EntropyAnalyzer{},
				&analyze.HeaderAnalyzer{},
				&analyze.ExifAnalyzer{},
			},
		},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Image for Analysis",
		Filters: []runtime.FileFilter{
			{DisplayName: "Images (*.png;*.jpg)", Pattern: "*.png;*.jpg;*.jpeg"},
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
	// 1. Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// 2. Wrap it in a ParsedFile for the engine
	ext := filepath.Ext(path)
	fileType := "Unknown"
	if ext == ".png" {
		fileType = "PNG Image"
	} else if ext == ".jpg" || ext == ".jpeg" {
		fileType = "JPEG Image"
	}

	pf := &parser.ParsedFile{
		Type:    fileType,
		RawData: data,
	}

	// 3. Run the repair logic
	repaired, err := a.engine.RepairFile(pf)
	if err != nil {
		return "", err
	}

	// 4. Save the new file
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
