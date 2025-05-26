package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type ImageProcessor struct {
	width         int
	height        int
	pixels        [][]Pixel
	originalPixels [][]Pixel
}

type Pixel struct {
	R, G, B, A uint8
}

type Filter func(*ImageProcessor)

func clamp(v float64) uint8 {
	if v > 255 {
		return 255
	}
	if v < 0 {
		return 0
	}
	return uint8(v)
}

func (ip *ImageProcessor) Clone() *ImageProcessor {
	newPixels := make([][]Pixel, ip.height)
	for i := range newPixels {
		newPixels[i] = make([]Pixel, ip.width)
		copy(newPixels[i], ip.pixels[i])
	}
	return &ImageProcessor{
		width:         ip.width,
		height:        ip.height,
		pixels:        newPixels,
		originalPixels: ip.originalPixels,
	}
}

func NewImageProcessor(path string) (*ImageProcessor, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var img image.Image
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	pixels := make([][]Pixel, height)
	originalPixels := make([][]Pixel, height)
	
	for y := 0; y < height; y++ {
		pixels[y] = make([]Pixel, width)
		originalPixels[y] = make([]Pixel, width)
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixel := Pixel{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
			pixels[y][x] = pixel
			originalPixels[y][x] = pixel
		}
	}

	return &ImageProcessor{
		width:         width,
		height:        height,
		pixels:        pixels,
		originalPixels: originalPixels,
	}, nil
}

func (ip *ImageProcessor) SaveImage(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, ip.width, ip.height))
	
	for y := 0; y < ip.height; y++ {
		for x := 0; x < ip.width; x++ {
			pixel := ip.pixels[y][x]
			img.Set(x, y, color.RGBA{pixel.R, pixel.G, pixel.B, pixel.A})
		}
	}

	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(output, img, &jpeg.Options{Quality: 90})
	case ".png":
		return png.Encode(output, img)
	default:
		return jpeg.Encode(output, img, &jpeg.Options{Quality: 90})
	}
}

func (ip *ImageProcessor) Grayscale() {
	for y := 0; y < ip.height; y++ {
		for x := 0; x < ip.width; x++ {
			pixel := ip.pixels[y][x]
			gray := uint8(float64(pixel.R)*0.299 + float64(pixel.G)*0.587 + float64(pixel.B)*0.114)
			ip.pixels[y][x] = Pixel{gray, gray, gray, pixel.A}
		}
	}
}

func (ip *ImageProcessor) Sepia() {
	for y := 0; y < ip.height; y++ {
		for x := 0; x < ip.width; x++ {
			pixel := ip.pixels[y][x]
			r := float64(pixel.R)
			g := float64(pixel.G)
			b := float64(pixel.B)

			newR := clamp((r * 0.393) + (g * 0.769) + (b * 0.189))
			newG := clamp((r * 0.349) + (g * 0.686) + (b * 0.168))
			newB := clamp((r * 0.272) + (g * 0.534) + (b * 0.131))

			ip.pixels[y][x] = Pixel{newR, newG, newB, pixel.A}
		}
	}
}

func (ip *ImageProcessor) Brightness(factor float64) {
	for y := 0; y < ip.height; y++ {
		for x := 0; x < ip.width; x++ {
			pixel := ip.pixels[y][x]
			ip.pixels[y][x] = Pixel{
				R: clamp(float64(pixel.R) * factor),
				G: clamp(float64(pixel.G) * factor),
				B: clamp(float64(pixel.B) * factor),
				A: pixel.A,
			}
		}
	}
}

func (ip *ImageProcessor) Contrast(factor float64) {
	factor = (100.0 + factor) / 100.0
	factor *= factor

	for y := 0; y < ip.height; y++ {
		for x := 0; x < ip.width; x++ {
			pixel := ip.pixels[y][x]
			
			r := float64(pixel.R)/255.0 - 0.5
			g := float64(pixel.G)/255.0 - 0.5
			b := float64(pixel.B)/255.0 - 0.5

			r = (r * factor) + 0.5
			g = (g * factor) + 0.5
			b = (b * factor) + 0.5

			ip.pixels[y][x] = Pixel{
				R: clamp(r * 255),
				G: clamp(g * 255),
				B: clamp(b * 255),
				A: pixel.A,
			}
		}
	}
}

func (ip *ImageProcessor) Sharpen(amount float64) {
	kernel := [][]float64{
		{0, -1, 0},
		{-1, 5, -1},
		{0, -1, 0},
	}
	ip.applyKernel(kernel, amount)
}

func (ip *ImageProcessor) EdgeDetection() {
	ip.Grayscale()
	kernel := [][]float64{
		{-1, -1, -1},
		{-1, 8, -1},
		{-1, -1, -1},
	}
	ip.applyKernel(kernel, 1.0)
}

func (ip *ImageProcessor) applyKernel(kernel [][]float64, amount float64) {
	result := make([][]Pixel, ip.height)
	for i := range result {
		result[i] = make([]Pixel, ip.width)
		copy(result[i], ip.pixels[i])
	}

	kernelSize := len(kernel)
	offset := kernelSize / 2

	for y := offset; y < ip.height-offset; y++ {
		for x := offset; x < ip.width-offset; x++ {
			var sumR, sumG, sumB float64

			for ky := 0; ky < kernelSize; ky++ {
				for kx := 0; kx < kernelSize; kx++ {
					pixel := ip.pixels[y-offset+ky][x-offset+kx]
					weight := kernel[ky][kx]
					sumR += float64(pixel.R) * weight
					sumG += float64(pixel.G) * weight
					sumB += float64(pixel.B) * weight
				}
			}

			originalPixel := ip.pixels[y][x]
			result[y][x] = Pixel{
				R: clamp(float64(originalPixel.R)*(1-amount) + sumR*amount),
				G: clamp(float64(originalPixel.G)*(1-amount) + sumG*amount),
				B: clamp(float64(originalPixel.B)*(1-amount) + sumB*amount),
				A: originalPixel.A,
			}
		}
	}

	ip.pixels = result
}

func (ip *ImageProcessor) Blur(radius int) {
	if radius < 1 {
		return
	}

	result := make([][]Pixel, ip.height)
	for i := range result {
		result[i] = make([]Pixel, ip.width)
	}

	kernel := make([][]float64, radius*2+1)
	sigma := float64(radius) / 2
	sum := 0.0

	for y := -radius; y <= radius; y++ {
		kernel[y+radius] = make([]float64, radius*2+1)
		for x := -radius; x <= radius; x++ {
			g := math.Exp(-(float64(x*x+y*y) / (2 * sigma * sigma)))
			kernel[y+radius][x+radius] = g
			sum += g
		}
	}

	for y := range kernel {
		for x := range kernel[y] {
			kernel[y][x] /= sum
		}
	}

	ip.applyKernel(kernel, 1.0)
}

func (ip *ImageProcessor) Reset() {
	for y := 0; y < ip.height; y++ {
		copy(ip.pixels[y], ip.originalPixels[y])
	}
}

type ProcessingOptions struct {
	Brightness    float64
	Contrast      float64
	Blur          int
	Sharpen       float64
	Grayscale     bool
	Sepia         bool
	EdgeDetection bool
	Quality       int
	Resize        string
}

func (ip *ImageProcessor) Resize(newWidth, newHeight int) {
	if newWidth <= 0 || newHeight <= 0 {
		return
	}

	result := make([][]Pixel, newHeight)
	for i := range result {
		result[i] = make([]Pixel, newWidth)
	}

	xRatio := float64(ip.width) / float64(newWidth)
	yRatio := float64(ip.height) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			px := int(float64(x) * xRatio)
			py := int(float64(y) * yRatio)
			result[y][x] = ip.pixels[py][px]
		}
	}

	ip.width = newWidth
	ip.height = newHeight
	ip.pixels = result
}

func (ip *ImageProcessor) ProcessImage(opts ProcessingOptions) {
	if opts.Brightness != 1.0 {
		ip.Brightness(opts.Brightness)
	}
	if opts.Contrast != 0 {
		ip.Contrast(opts.Contrast)
	}
	if opts.Grayscale {
		ip.Grayscale()
	}
	if opts.Sepia {
		ip.Sepia()
	}
	if opts.EdgeDetection {
		ip.EdgeDetection()
	}
	if opts.Sharpen != 0 {
		ip.Sharpen(opts.Sharpen)
	}
	if opts.Blur > 0 {
		ip.Blur(opts.Blur)
	}
	if opts.Resize != "" {
		parts := strings.Split(opts.Resize, "x")
		if len(parts) == 2 {
			width := 0
			height := 0
			fmt.Sscanf(parts[0], "%d", &width)
			fmt.Sscanf(parts[1], "%d", &height)
			if width > 0 && height > 0 {
				ip.Resize(width, height)
			}
		}
	}
}

func processFile(inputPath, outputPath string, opts ProcessingOptions) error {
	processor, err := NewImageProcessor(inputPath)
	if err != nil {
		return fmt.Errorf("error loading image: %v", err)
	}

	processor.ProcessImage(opts)

	return processor.SaveImage(outputPath)
}

func processBatch(inputDir, outputDir string, opts ProcessingOptions) error {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}

		inputPath := filepath.Join(inputDir, file.Name())
		outputPath := filepath.Join(outputDir, file.Name())

		if err := processFile(inputPath, outputPath, opts); err != nil {
			log.Printf("Error processing %s: %v", file.Name(), err)
			continue
		}

		fmt.Printf("Processed: %s -> %s\n", file.Name(), outputPath)
	}
	return nil
}

func main() {
	var opts ProcessingOptions

	input := flag.String("input", "", "Input file or directory")
	output := flag.String("output", "", "Output file or directory")
	flag.Float64Var(&opts.Brightness, "brightness", 1.0, "Brightness factor (0.0-2.0)")
	flag.Float64Var(&opts.Contrast, "contrast", 0, "Contrast adjustment (-100 to 100)")
	flag.IntVar(&opts.Blur, "blur", 0, "Blur radius (0-10)")
	flag.Float64Var(&opts.Sharpen, "sharpen", 0, "Sharpen amount (0.0-1.0)")
	flag.BoolVar(&opts.Grayscale, "grayscale", false, "Convert to grayscale")
	flag.BoolVar(&opts.Sepia, "sepia", false, "Apply sepia effect")
	flag.BoolVar(&opts.EdgeDetection, "edge", false, "Apply edge detection")
	flag.IntVar(&opts.Quality, "quality", 90, "JPEG output quality (0-100)")
	flag.StringVar(&opts.Resize, "resize", "", "Resize image (e.g., 800x600)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "GoLens - Image Processing Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  Single file:    golens -input input.jpg -output output.jpg -brightness 1.2 -contrast 10\n")
		fmt.Fprintf(os.Stderr, "  Batch process:   golens -input ./input_dir -output ./output_dir -grayscale -blur 2\n")
	}

	flag.Parse()

	if *input == "" || *output == "" {
		flag.Usage()
		os.Exit(1)
	}

	inputInfo, err := os.Stat(*input)
	if err != nil {
		log.Fatalf("Error accessing input: %v", err)
	}

	if inputInfo.IsDir() {
		if err := os.MkdirAll(*output, 0755); err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}
		fmt.Println("Starting batch processing...")
		if err := processBatch(*input, *output, opts); err != nil {
			log.Fatalf("Batch processing failed: %v", err)
		}
		fmt.Println("Batch processing completed successfully!")
	} else {
		fmt.Println("Processing single file...")
		if err := processFile(*input, *output, opts); err != nil {
			log.Fatalf("Processing failed: %v", err)
		}
		fmt.Println("File processed successfully!")
	}
}
