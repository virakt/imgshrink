package main

import (
	"fmt"

	"github.com/virakt/imgshrink/internal/api"
	"github.com/virakt/imgshrink/internal/compressor"
)

func main() {
	// Create API instance
	imageAPI := api.NewImageAPI()

	// Get default options
	options := imageAPI.GetDefaultOptions()
	options.Quality = 80

	// Compress a single image
	result, err := imageAPI.CompressImage("img.jpg", options)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Compressed: %s -> %s\n",
		compressor.FormatBytes(result.InputSize),
		compressor.FormatBytes(result.OutputSize))
	fmt.Printf("Reduction: %.1f%%\n", result.Reduction)

	// Batch compress
	// files := []string{"image1.jpg", "image2.png", "image3.jpg"}
	// progressChan := make(chan *compressor.CompressionResult)

	// go func() {
	//     for result := range progressChan {
	//         fmt.Printf("Completed: %s\n", result.InputPath)
	//     }
	// }()

	// batchResult, _ := imageAPI.BatchCompress(files, options, progressChan)
	// fmt.Printf("Total reduction: %.1f%%\n", batchResult.TotalReduction)
}
