package compressor

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
)

// ImageFormat represents supported image formats
type ImageFormat string

const (
	FormatJPEG ImageFormat = "jpeg"
	FormatPNG  ImageFormat = "png"
)

// CompressionOptions holds all compression settings
type CompressionOptions struct {
	// Common options
	Quality       int     // 1-100 for JPEG, ignored for PNG
	ResizePercent float64 // 0-100, 0 means no resize
	ResizeWidth   int     // Target width, 0 means auto
	ResizeHeight  int     // Target height, 0 means auto
	StripMetadata bool    // Remove EXIF and other metadata
	OutputDir     string  // Output directory, empty means same as input
	OutputSuffix  string  // Suffix to add to filename (e.g., "_compressed")

	// JPEG specific
	Progressive     bool   // Progressive JPEG encoding
	ChromaSubsample string // "4:4:4", "4:2:2", "4:2:0"

	// PNG specific
	CompressionLevel int  // 0-9, higher = more compression
	Interlaced       bool // Adam7 interlacing
}

// DefaultOptions returns sensible default compression options
func DefaultOptions() CompressionOptions {
	return CompressionOptions{
		Quality:          85,
		ResizePercent:    0,
		ResizeWidth:      0,
		ResizeHeight:     0,
		StripMetadata:    true,
		OutputDir:        "",
		OutputSuffix:     "_compressed",
		Progressive:      true,
		ChromaSubsample:  "4:2:0",
		CompressionLevel: 6,
		Interlaced:       false,
	}
}

// ImageInfo contains information about an image
type ImageInfo struct {
	Path      string
	Format    ImageFormat
	Width     int
	Height    int
	Size      int64
	ColorMode string
}

// CompressionResult contains the result of a compression operation
type CompressionResult struct {
	InputPath  string
	OutputPath string
	InputSize  int64
	OutputSize int64
	Reduction  float64 // Percentage reduction
	Width      int
	Height     int
	Success    bool
	Error      error
}

// Compressor interface defines the compression operations
type Compressor interface {
	Compress(inputPath string, options CompressionOptions) (*CompressionResult, error)
	GetInfo(inputPath string) (*ImageInfo, error)
	EstimateSize(inputPath string, options CompressionOptions) (int64, error)
}

// GetImageFormat detects the image format from file extension
func GetImageFormat(path string) (ImageFormat, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return FormatJPEG, nil
	case ".png":
		return FormatPNG, nil
	default:
		return "", fmt.Errorf("unsupported image format: %s", ext)
	}
}

// GetImageInfo returns information about an image file
func GetImageInfo(path string) (*ImageInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Decode image config (doesn't load full image)
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	imgFormat, err := GetImageFormat(path)
	if err != nil {
		// Try to use detected format
		switch format {
		case "jpeg":
			imgFormat = FormatJPEG
		case "png":
			imgFormat = FormatPNG
		default:
			return nil, fmt.Errorf("unsupported format: %s", format)
		}
	}

	return &ImageInfo{
		Path:      path,
		Format:    imgFormat,
		Width:     config.Width,
		Height:    config.Height,
		Size:      stat.Size(),
		ColorMode: format,
	}, nil
}

// GenerateOutputPath creates the output path based on options
func GenerateOutputPath(inputPath string, options CompressionOptions) string {
	dir := filepath.Dir(inputPath)
	if options.OutputDir != "" {
		dir = options.OutputDir
	}

	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), ext)

	return filepath.Join(dir, base+options.OutputSuffix+ext)
}

// FormatBytes formats bytes into human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CalculateReduction calculates the percentage reduction
func CalculateReduction(original, compressed int64) float64 {
	if original == 0 {
		return 0
	}
	return (1 - float64(compressed)/float64(original)) * 100
}
