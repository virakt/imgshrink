package api

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/virakt/imgshrink/internal/compressor"
)

// ImageAPI provides the main API for image compression operations
type ImageAPI struct {
	jpegCompressor *compressor.JPEGCompressor
	pngCompressor  *compressor.PNGCompressor
}

// NewImageAPI creates a new ImageAPI instance
func NewImageAPI() *ImageAPI {
	return &ImageAPI{
		jpegCompressor: compressor.NewJPEGCompressor(),
		pngCompressor:  compressor.NewPNGCompressor(),
	}
}

// CompressImage compresses a single image with the given options
func (api *ImageAPI) CompressImage(inputPath string, options compressor.CompressionOptions) (*compressor.CompressionResult, error) {
	format, err := compressor.GetImageFormat(inputPath)
	if err != nil {
		return nil, err
	}

	switch format {
	case compressor.FormatJPEG:
		return api.jpegCompressor.Compress(inputPath, options)
	case compressor.FormatPNG:
		return api.pngCompressor.Compress(inputPath, options)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GetImageInfo returns information about an image
func (api *ImageAPI) GetImageInfo(inputPath string) (*compressor.ImageInfo, error) {
	return compressor.GetImageInfo(inputPath)
}

// EstimateSize estimates the compressed size of an image
func (api *ImageAPI) EstimateSize(inputPath string, options compressor.CompressionOptions) (int64, error) {
	format, err := compressor.GetImageFormat(inputPath)
	if err != nil {
		return 0, err
	}

	switch format {
	case compressor.FormatJPEG:
		return api.jpegCompressor.EstimateSize(inputPath, options)
	case compressor.FormatPNG:
		return api.pngCompressor.EstimateSize(inputPath, options)
	default:
		return 0, fmt.Errorf("unsupported format: %s", format)
	}
}

// BatchResult contains results for batch compression
type BatchResult struct {
	Results        []*compressor.CompressionResult
	TotalInput     int64
	TotalOutput    int64
	TotalReduction float64
	SuccessCount   int
	FailCount      int
}

// BatchCompress compresses multiple images concurrently
func (api *ImageAPI) BatchCompress(inputPaths []string, options compressor.CompressionOptions, progressChan chan<- *compressor.CompressionResult) (*BatchResult, error) {
	batchResult := &BatchResult{
		Results: make([]*compressor.CompressionResult, 0, len(inputPaths)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Limit concurrency to avoid overwhelming the system
	semaphore := make(chan struct{}, 4)

	for _, path := range inputPaths {
		wg.Add(1)
		go func(inputPath string) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			result, _ := api.CompressImage(inputPath, options)

			mu.Lock()
			batchResult.Results = append(batchResult.Results, result)
			if result.Success {
				batchResult.SuccessCount++
				batchResult.TotalInput += result.InputSize
				batchResult.TotalOutput += result.OutputSize
			} else {
				batchResult.FailCount++
			}
			mu.Unlock()

			if progressChan != nil {
				progressChan <- result
			}
		}(path)
	}

	wg.Wait()

	if batchResult.TotalInput > 0 {
		batchResult.TotalReduction = compressor.CalculateReduction(batchResult.TotalInput, batchResult.TotalOutput)
	}

	return batchResult, nil
}

// ScanDirectory scans a directory for supported images
func (api *ImageAPI) ScanDirectory(dirPath string, recursive bool) ([]string, error) {
	var images []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a supported image format
		_, err = compressor.GetImageFormat(path)
		if err == nil {
			images = append(images, path)
		}

		return nil
	}

	if err := filepath.Walk(dirPath, walkFn); err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return images, nil
}

// ValidateImage checks if a file is a valid supported image
func (api *ImageAPI) ValidateImage(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	// Check format
	format, err := compressor.GetImageFormat(path)
	if err != nil {
		return err
	}

	// Try to get image info to validate it's a valid image
	_, err = compressor.GetImageInfo(path)
	if err != nil {
		return fmt.Errorf("invalid %s image: %w", format, err)
	}

	return nil
}

// GetDefaultOptions returns the default compression options
func (api *ImageAPI) GetDefaultOptions() compressor.CompressionOptions {
	return compressor.DefaultOptions()
}

// PreviewCompression returns a preview of what the compression would do
type CompressionPreview struct {
	InputPath          string
	InputSize          int64
	EstimatedSize      int64
	EstimatedReduction float64
	Format             compressor.ImageFormat
	Width              int
	Height             int
	NewWidth           int
	NewHeight          int
}

// PreviewCompression generates a preview of the compression result
func (api *ImageAPI) PreviewCompression(inputPath string, options compressor.CompressionOptions) (*CompressionPreview, error) {
	info, err := api.GetImageInfo(inputPath)
	if err != nil {
		return nil, err
	}

	estimatedSize, err := api.EstimateSize(inputPath, options)
	if err != nil {
		return nil, err
	}

	preview := &CompressionPreview{
		InputPath:          inputPath,
		InputSize:          info.Size,
		EstimatedSize:      estimatedSize,
		EstimatedReduction: compressor.CalculateReduction(info.Size, estimatedSize),
		Format:             info.Format,
		Width:              info.Width,
		Height:             info.Height,
		NewWidth:           info.Width,
		NewHeight:          info.Height,
	}

	// Calculate new dimensions if resizing
	if options.ResizePercent > 0 && options.ResizePercent < 100 {
		preview.NewWidth = int(float64(info.Width) * options.ResizePercent / 100)
		preview.NewHeight = int(float64(info.Height) * options.ResizePercent / 100)
	} else if options.ResizeWidth > 0 || options.ResizeHeight > 0 {
		if options.ResizeWidth > 0 && options.ResizeHeight > 0 {
			preview.NewWidth = options.ResizeWidth
			preview.NewHeight = options.ResizeHeight
		} else if options.ResizeWidth > 0 {
			ratio := float64(options.ResizeWidth) / float64(info.Width)
			preview.NewWidth = options.ResizeWidth
			preview.NewHeight = int(float64(info.Height) * ratio)
		} else {
			ratio := float64(options.ResizeHeight) / float64(info.Height)
			preview.NewHeight = options.ResizeHeight
			preview.NewWidth = int(float64(info.Width) * ratio)
		}
	}

	return preview, nil
}
