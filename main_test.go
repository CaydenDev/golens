package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageProcessor(t *testing.T) {
	testImg := createTestImage(t)
	defer os.Remove(testImg)

	tests := []struct {
		name string
		opts ProcessingOptions
	}{
		{
			name: "Grayscale",
			opts: ProcessingOptions{Grayscale: true},
		},
		{
			name: "Brightness",
			opts: ProcessingOptions{Brightness: 1.2},
		},
		{
			name: "Blur",
			opts: ProcessingOptions{Blur: 2},
		},
		{
			name: "Multiple Effects",
			opts: ProcessingOptions{
				Brightness: 1.1,
				Contrast:   10,
				Blur:       1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(t.TempDir(), "output.jpg")
			err := processFile(testImg, outputPath, tt.opts)
			if err != nil {
				t.Errorf("processFile() error = %v", err)
				return
			}

			info, err := os.Stat(outputPath)
			if err != nil {
				t.Errorf("Failed to stat output file: %v", err)
				return
			}
			if info.Size() == 0 {
				t.Error("Output file is empty")
			}
		})
	}
}

func createTestImage(t *testing.T) string {
	t.Helper()
	
	tmpfile, err := os.CreateTemp(t.TempDir(), "test*.jpg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpfile.Close()

	processor := &ImageProcessor{
		width:  100,
		height: 100,
		pixels: make([][]Pixel, 100),
	}

	for y := 0; y < processor.height; y++ {
		processor.pixels[y] = make([]Pixel, processor.width)
		for x := 0; x < processor.width; x++ {
			processor.pixels[y][x] = Pixel{
				R: uint8(x % 255),
				G: uint8(y % 255),
				B: uint8((x + y) % 255),
				A: 255,
			}
		}
	}

	if err := processor.SaveImage(tmpfile.Name()); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	return tmpfile.Name()
}
