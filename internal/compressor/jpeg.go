package compressor

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"github.com/disintegration/imaging"
)

// JPEGCompressor handles JPEG image compression
type JPEGCompressor struct{}

// NewJPEGCompressor creates a new JPEG compressor
func NewJPEGCompressor() *JPEGCompressor {
	return &JPEGCompressor{}
}

// Compress compresses a JPEG image with the given options
func (c *JPEGCompressor) Compress(inputPath string, options CompressionOptions) (*CompressionResult, error) {
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

	// Encode with compression options
	jpegOptions := &jpeg.Options{
		Quality: options.Quality,
	}

	if err := jpeg.Encode(outFile, img, jpegOptions); err != nil {
		result.Error = fmt.Errorf("failed to encode JPEG: %w", err)
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

// GetInfo returns information about a JPEG image
func (c *JPEGCompressor) GetInfo(inputPath string) (*ImageInfo, error) {
	return GetImageInfo(inputPath)
}

// EstimateSize estimates the compressed size based on options
func (c *JPEGCompressor) EstimateSize(inputPath string, options CompressionOptions) (int64, error) {
	info, err := GetImageInfo(inputPath)
	if err != nil {
		return 0, err
	}

	// Rough estimation based on quality
	// This is a simplified estimation; actual size depends on image content
	qualityFactor := float64(options.Quality) / 100.0

	// Base compression ratio for JPEG (typically 10:1 to 20:1)
	baseRatio := 0.1 + (qualityFactor * 0.4) // 10% to 50% of original

	// Adjust for resize
	resizeFactor := 1.0
	if options.ResizePercent > 0 && options.ResizePercent < 100 {
		resizeFactor = (options.ResizePercent / 100.0) * (options.ResizePercent / 100.0)
	}

	estimatedSize := float64(info.Size) * baseRatio * resizeFactor
	return int64(estimatedSize), nil
}

// applyResize resizes the image based on options
func applyResize(img image.Image, options CompressionOptions) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Resize by percentage
	if options.ResizePercent > 0 && options.ResizePercent < 100 {
		newWidth := int(float64(width) * options.ResizePercent / 100)
		newHeight := int(float64(height) * options.ResizePercent / 100)
		return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	}

	// Resize by dimensions
	if options.ResizeWidth > 0 || options.ResizeHeight > 0 {
		newWidth := options.ResizeWidth
		newHeight := options.ResizeHeight

		// Maintain aspect ratio if only one dimension is specified
		if newWidth == 0 {
			ratio := float64(newHeight) / float64(height)
			newWidth = int(float64(width) * ratio)
		} else if newHeight == 0 {
			ratio := float64(newWidth) / float64(width)
			newHeight = int(float64(height) * ratio)
		}

		return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	}

	return img
}
