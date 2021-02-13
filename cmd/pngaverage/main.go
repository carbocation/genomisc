package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/overlay"
)

// Safe for concurrent use by multiple goroutines
var client *storage.Client

func main() {
	fmt.Fprintf(os.Stderr, "%q\n", os.Args)

	var imagePaths flagSlice
	var outputPath string

	flag.Var(&imagePaths, "input", "Paths to files with grayscale PNGs to merge. Pass once per image (e.g., -input ./img1.png -input ./img2.png).")
	flag.StringVar(&outputPath, "output", "", "Path to folder where output image will be put")
	flag.Parse()

	if len(imagePaths) < 1 || outputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the Google Storage client only if we're pointing to Google
	// Storage paths.
	for _, imagePath := range imagePaths {
		if strings.HasPrefix(imagePath, "gs://") || strings.HasPrefix(outputPath, "gs://") {
			var err error
			client, err = storage.NewClient(context.Background())
			if err != nil {
				log.Fatalln(err)
			}

			break
		}
	}

	outputName := ""
	for _, filename := range imagePaths {
		outputName += filepath.Base(filename)
	}

	if err := processOneImage(imagePaths, outputPath, outputName); err != nil {
		log.Fatalln(err)
	}

}

func processOneImage(imagePaths []string, outputPath, outputFilename string) error {
	// Open the files
	imageFiles := make([]image.Image, 0, len(imagePaths))
	for _, imagePath := range imagePaths {
		overlay1, err := overlay.OpenImageFromLocalFileOrGoogleStorage(imagePath, client)
		if err != nil {
			return err
		}

		imageFiles = append(imageFiles, overlay1)
	}

	if len(imageFiles) < 1 {
		return fmt.Errorf("No images were found")
	}

	r1 := imageFiles[0].Bounds()
	output := image.NewGray16(r1)

	var graySum float64
	n := float64(len(imageFiles))

	for y := 0; y < r1.Bounds().Max.Y; y++ {
		for x := 0; x < r1.Bounds().Max.X; x++ {

			graySum = 0.0
			for _, imgFile := range imageFiles {
				col1 := imgFile.At(x, y)
				graySum += float64(col1.(color.Gray16).Y)
			}

			// Set the output to the mean intensity
			output.Set(x, y, color.Gray16{Y: uint16(graySum / n)})
		}
	}

	outFile, err := os.Create(filepath.Join(outputPath, outputFilename))
	if err != nil {
		return err
	}

	// Write the PNG representation of our ID-encoded image to disk
	bw := bufio.NewWriter(outFile)
	defer bw.Flush()
	return png.Encode(bw, output)
}
