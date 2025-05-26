# GoLens

[![GoLens CI](https://github.com/CaydenDev/golens/actions/workflows/test.yml/badge.svg)](https://github.com/USER/golens/actions/workflows/test.yml)

## Installation

### From Source
```bash
git clone https://github.com/CaydenDev/golens.git
cd golens
go build
```

### Using Go Install
```bash
go install github.com/CaydenDev/golens@latest
```

## Usage

### Basic Usage
```bash
golens -input input.jpg -output output.jpg -brightness 1.2 -contrast 10
```

### Batch Processing
```bash
golens -input ./input_dir -output ./output_dir -grayscale -blur 2
```

### Available Options
```
-input     : Input file or directory
-output    : Output file or directory
-brightness: Brightness factor (0.0-2.0)
-contrast  : Contrast adjustment (-100 to 100)
-blur      : Blur radius (0-10)
-sharpen   : Sharpen amount (0.0-1.0)
-grayscale : Convert to grayscale
-sepia     : Apply sepia effect
-edge      : Apply edge detection
-quality   : JPEG output quality (0-100)
-resize    : Resize image (e.g., 800x600)
```

## Examples

### Convert to Grayscale
```bash
golens -input color.jpg -output gray.jpg -grayscale
```

### Resize Image
```bash
golens -input large.jpg -output small.jpg -resize 800x600
```

### Apply Multiple Effects
```bash
golens -input input.jpg -output output.jpg -brightness 1.1 -contrast 15 -blur 2
```

### Batch Process Directory
```bash
golens -input ./photos -output ./processed -grayscale -sepia
```

## Development

### Running Tests
```bash
go test -v ./...
```

### Building from Source
```bash
go build -o golens
```