package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	_ "github.com/carbocation/genomisc/compileinfoprint"
	"github.com/carbocation/genomisc/overlay"
)

// Safe for concurrent use by multiple goroutines
var client *storage.Client

func main() {
	fmt.Fprintf(os.Stderr, "%q\n", os.Args)

	var threshold int
	var path1, path2, manifest, suffix, labelsFile string

	flag.StringVar(&path1, "path1", "", "Path to folder with source images (base/truth)")
	flag.StringVar(&path2, "output", "", "Path to folder where modified images will be put")
	flag.StringVar(&manifest, "manifest", "", "(Optional) Path to manifest. If provided, will only look at files in the manifest rather than listing the entire directory's contents.")
	flag.StringVar(&suffix, "suffix", ".png.mask.png", "(Optional) Suffix after .dcm. Only used if using the -manifest option.")
	flag.StringVar(&labelsFile, "replacements", "", "(Optional) json file with replacements, formatted like demo.replacements.json")
	flag.Parse()

	if path1 == "" || path2 == "" || labelsFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the Google Storage client only if we're pointing to Google
	// Storage paths.
	if strings.HasPrefix(path1, "gs://") || strings.HasPrefix(path2, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	lf, err := os.Open(labelsFile)
	if err != nil {
		log.Fatalln(err)
	}

	newLabels, err := ParseReplacementFile(lf)
	if err != nil {
		log.Fatalln(err)
	}

	if manifest != "" {

		if err := runSlice(path1, path2, suffix, newLabels, manifest, threshold); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if err := runFolder(path1, path2, newLabels, threshold); err != nil {
		log.Fatalln(err)
	}

}

func runSlice(path1, path2, suffix string, newLabels ReplacementMap, manifest string, threshold int) error {

	dicoms, err := getDicomSlice(manifest)
	if err != nil {
		return err
	}

	concurrency := 4 * runtime.NumCPU()
	sem := make(chan bool, concurrency)

	// Process every image in the manifest
	for i, file := range dicoms {
		sem <- true
		go func(file string) {

			// The main purpose of this loop is to handle a specific filesystem
			// error (input/output error) that largely happens with GCSFuse, and
			// retry a few times before giving up.
			for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

				err := processOneImageFromPath(path1+"/"+file+suffix, path2+"/"+file+suffix, file, threshold, newLabels)

				if err != nil && loadAttempts == maxLoadAttempts {

					// We've exhausted our retries. Fail hard.
					log.Fatalln(err)

				} else if err != nil && strings.Contains(err.Error(), "input/output error") {

					// If it's an i/o error, we can retry
					log.Println("Sleeping 5s to recover from", err.Error(), ". Attempt #", loadAttempts)
					time.Sleep(5 * time.Second)
					continue

				} else if err != nil {

					// If it's an error that is not an i/o error, don't retry
					log.Println(err)
					break

				}

				// If no error, don't retry
				break
			}

			<-sem
		}(file)

		if (i+1)%1000 == 0 {
			log.Printf("Processed %d images\n", i+1)
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func runFolder(path1, path2 string, newLabels ReplacementMap, threshold int) error {

	files, err := scanFolder(path1)
	if err != nil {
		return err
	}

	// concurrency := 4 * runtime.NumCPU()
	concurrency := 1
	sem := make(chan bool, concurrency)

	// Process every image in the folder
	for i, file := range files {
		if file.IsDir() {
			continue
		}

		sem <- true
		go func(file string) {
			var err error
			if strings.HasSuffix(file, ".tar.gz") {
				if err = processOneTarGZFilepath(path1+"/"+file, path2, file, threshold, newLabels); err != nil {
					log.Println(err)
				}
			} else if strings.HasSuffix(file, ".png") ||
				strings.HasSuffix(file, ".gif") ||
				strings.HasSuffix(file, ".bmp") ||
				strings.HasSuffix(file, ".jpeg") ||
				strings.HasSuffix(file, ".jpg") {
				if err = processOneImageFromPath(path1+"/"+file, path2, file, threshold, newLabels); err != nil {
					log.Printf("%s: %s\n", file, err)
				}
			}
			<-sem
		}(file.Name())

		if (i+1)%1000 == 0 {
			log.Printf("Processed %d images\n", i+1)
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func processOneImageFromPath(filePath1, filePath2, filename string, threshold int, newLabels ReplacementMap) error {
	// Open the files
	overlay1, err := overlay.OpenImageFromLocalFileOrGoogleStorage(filePath1, client)
	if err != nil {
		return err
	}

	outImg, err := processOneImage(overlay1, filePath2, filename, threshold, newLabels)
	if err != nil {
		return err
	}

	outFile, err := os.Create(filepath.Join(filePath2, filename))
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(outFile)
	defer bw.Flush()

	// Write the PNG representation of our ID-encoded image to disk
	return png.Encode(bw, outImg)
}

func processOneImage(overlay1 image.Image, filePath2, filename string, threshold int, newLabels ReplacementMap) (image.Image, error) {

	output := overlay1.(draw.Image)
	r1 := overlay1.Bounds()

	for y := 0; y < r1.Bounds().Max.Y; y++ {
		for x := 0; x < r1.Bounds().Max.X; x++ {

			col1 := overlay1.At(x, y)

			// Replace the old ID with the new ID
			newCol1 := newLabels.Replace(col1)

			if newCol1 != col1 {
				output.Set(x, y, newCol1)
			}
		}
	}

	return output, nil
}
