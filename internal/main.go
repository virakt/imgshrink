package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/virakt/imgshrink/internal/api"
	"github.com/virakt/imgshrink/internal/compressor"
	"github.com/virakt/imgshrink/internal/tui"
)

func main() {
	args := os.Args[1:]

	// If no arguments, show usage and start TUI
	if len(args) == 0 {
		if err := tui.Run(nil); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for help flag
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		printUsage()
		return
	}

	// Check for version flag
	if len(args) == 1 && (args[0] == "-v" || args[0] == "--version") {
		fmt.Println("ImgShrink v1.0.0")
		return
	}

	// Check for CLI mode flag
	cliMode := false
	var files []string
	var outputDir string
	quality := 85
	compressionLevel := 6

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--cli", "-c":
			cliMode = true
		case "--output", "-o":
			if i+1 < len(args) {
				outputDir = args[i+1]
				i++
			}
		case "--quality", "-q":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &quality)
				i++
			}
		case "--level", "-l":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &compressionLevel)
				i++
			}
		default:
			// Assume it's a file path
			files = append(files, args[i])
		}
	}

	// Expand glob patterns
	var expandedFiles []string
	for _, pattern := range files {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error expanding pattern %s: %v\n", pattern, err)
			continue
		}
		if len(matches) == 0 {
			// Not a pattern, use as-is
			expandedFiles = append(expandedFiles, pattern)
		} else {
			expandedFiles = append(expandedFiles, matches...)
		}
	}

	if cliMode {
		// Run in CLI mode (no TUI)
		runCLI(expandedFiles, outputDir, quality, compressionLevel)
	} else {
		// Start TUI with files
		if err := tui.Run(expandedFiles); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println(`ImgShrink - Image Compression Tool

Usage:
  imgshrink [options] [files...]

Options:
  -h, --help       Show this help message
  -v, --version    Show version
  -c, --cli        Run in CLI mode (no TUI)
  -o, --output     Output directory
  -q, --quality    JPEG quality (1-100, default: 85)
  -l, --level      PNG compression level (0-9, default: 6)

Examples:
  imgshrink                          # Start TUI
  imgshrink image.jpg                # Start TUI with file
  imgshrink *.png                    # Start TUI with multiple files
  imgshrink -c -q 80 image.jpg       # CLI mode with quality 80
  imgshrink -c -o ./output *.jpg     # CLI mode with output directory

Supported formats: JPEG (.jpg, .jpeg), PNG (.png)`)
}

func runCLI(files []string, outputDir string, quality, compressionLevel int) {
	if len(files) == 0 {
		fmt.Println("No files specified")
		os.Exit(1)
	}

	imageAPI := api.NewImageAPI()
	options := compressor.DefaultOptions()
	options.Quality = quality
	options.CompressionLevel = compressionLevel
	if outputDir != "" {
		options.OutputDir = outputDir
	}

	fmt.Printf("Compressing %d file(s)...\n\n", len(files))

	var totalInput, totalOutput int64
	var successCount, failCount int

	for _, file := range files {
		// Validate file
		if err := imageAPI.ValidateImage(file); err != nil {
			fmt.Printf("✗ %s: %v\n", file, err)
			failCount++
			continue
		}

		// Compress
		result, err := imageAPI.CompressImage(file, options)
		if err != nil {
			fmt.Printf("✗ %s: %v\n", file, err)
			failCount++
			continue
		}

		if result.Success {
			successCount++
			totalInput += result.InputSize
			totalOutput += result.OutputSize

			fmt.Printf("✓ %s\n", file)
			fmt.Printf("  %s → %s (%.1f%% reduction)\n",
				compressor.FormatBytes(result.InputSize),
				compressor.FormatBytes(result.OutputSize),
				result.Reduction)
			fmt.Printf("  Output: %s\n", result.OutputPath)
		} else {
			failCount++
			fmt.Printf("✗ %s: %v\n", file, result.Error)
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("Completed: %d successful, %d failed\n", successCount, failCount)
	if successCount > 0 {
		reduction := compressor.CalculateReduction(totalInput, totalOutput)
		fmt.Printf("Total: %s → %s (%.1f%% reduction)\n",
			compressor.FormatBytes(totalInput),
			compressor.FormatBytes(totalOutput),
			reduction)
		fmt.Printf("Saved: %s\n", compressor.FormatBytes(totalInput-totalOutput))
	}
}
