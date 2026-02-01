package compressor

import (
	"fmt"
	"image/png"
	"os"

	"github.com/disintegration/imaging"
)

// PNGCompressor handles PNG image compression
type PNGCompressor struct{}

// NewPNGCompressor creates a new PNG compressor
func NewPNGCompressor() *PNGCompressor {
	return &PNGCompressor{}
}

// Compress compresses a PNG image with the given options
func (c *PNGCompressor) Compress(inputPath string, options CompressionOptions) (*CompressionResult, error) {
	result := &CompressionResult{
		InputPath: inputPath,
		Success:   false,
	}

	// Get input file info
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to get input file info: %w", err)
		return result, result.Error
	}
	result.InputSize = inputInfo.Size()

	// Open and decode the image
	img, err := imaging.Open(inputPath, imaging.AutoOrientation(true))
	if err != nil {
		result.Error = fmt.Errorf("failed to open image: %w", err)
		return result, result.Error
	}

	// Apply resize if specified
	img = applyResize(img, options)

	// Get dimensions
	bounds := img.Bounds()
	result.Width = bounds.Dx()
	result.Height = bounds.Dy()

	// Generate output path
	outputPath := GenerateOutputPath(inputPath, options)
	result.OutputPath = outputPath

	// Ensure output directory exists
	if options.OutputDir != "" {
		if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
			result.Error = fmt.Errorf("failed to create output directory: %w", err)
			return result, result.Error
		}
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to create output file: %w", err)
		return result, result.Error
	}
	defer outFile.Close()

	// Map compression level (0-9) to PNG compression level
	var compressionLevel png.CompressionLevel
	switch {
	case options.CompressionLevel <= 0:
		compressionLevel = png.NoCompression
	case options.CompressionLevel <= 3:
		compressionLevel = png.BestSpeed
	case options.CompressionLevel <= 6:
		compressionLevel = png.DefaultCompression
	default:
		compressionLevel = png.BestCompression
	}

	// Create PNG encoder with options
	encoder := &png.Encoder{
		CompressionLevel: compressionLevel,
	}

	if err := encoder.Encode(outFile, img); err != nil {
		result.Error = fmt.Errorf("failed to encode PNG: %w", err)
		return result, result.Error
	}

	// Get output file size
	outInfo, err := os.Stat(outputPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to get output file info: %w", err)
		return result, result.Error
	}
	result.OutputSize = outInfo.Size()
	result.Reduction = CalculateReduction(result.InputSize, result.OutputSize)
	result.Success = true

	return result, nil
}

// GetInfo returns information about a PNG image
func (c *PNGCompressor) GetInfo(inputPath string) (*ImageInfo, error) {
	return GetImageInfo(inputPath)
}

// EstimateSize estimates the compressed size based on options
func (c *PNGCompressor) EstimateSize(inputPath string, options CompressionOptions) (int64, error) {
	info, err := GetImageInfo(inputPath)
	if err != nil {
		return 0, err
	}

	// PNG compression is lossless, so estimation is based on compression level
	// Higher compression = smaller file but slower
	compressionFactor := 1.0 - (float64(options.CompressionLevel) * 0.05) // 5% per level

	// Adjust for resize
	resizeFactor := 1.0
	if options.ResizePercent > 0 && options.ResizePercent < 100 {
		resizeFactor = (options.ResizePercent / 100.0) * (options.ResizePercent / 100.0)
	}

	estimatedSize := float64(info.Size) * compressionFactor * resizeFactor
	return int64(estimatedSize), nil
}
