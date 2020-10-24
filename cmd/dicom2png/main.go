package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

// Safe for concurrent use by multiple goroutines
var client *storage.Client

func main() {

	fmt.Fprintf(os.Stderr, "This dicom2png binary was built at: %s\n", builddate)

	var inputPath, outputPath, manifest string
	var includeOverlay bool
	flag.StringVar(&inputPath, "raw", "", "Path to the folder containing the raw zip files (if begins with gs://, will try the Google Storage URL)")
	flag.StringVar(&outputPath, "out", "", "Path to the local folder where the extracted PNGs will go")
	flag.StringVar(&manifest, "manifest", "", "Manifest file containing Zip names and Dicom names.")
	flag.BoolVar(&includeOverlay, "include-overlay", true, "Print the overlay on top of the images?")

	flag.Parse()
	if inputPath == "" || outputPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Initialize the Google Storage client only if we're pointing to Google
	// Storage paths.
	if strings.HasPrefix(inputPath, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err := run(inputPath, outputPath, manifest, includeOverlay); err != nil {
		log.Fatalln(err)
	}

}

func run(inputPath, outputPath, manifest string, includeOverlay bool) error {

	zipMap, err := getZipMap(manifest)
	if err != nil {
		return err
	}

	// Now iterate over the zips and print out the requested dicoms as images

	concurrency := runtime.NumCPU()
	sem := make(chan bool, concurrency)

	for zipFile, dicomList := range zipMap {
		sem <- true
		go func(zipFile string, dicoms []string) {

			// The main purpose of this loop is to handle a specific filesystem
			// error (input/output error) that largely happens with GCSFuse, and
			// retry a few times before giving up.
			for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

				err := ProcessOneZipFile(inputPath, outputPath, zipFile, dicoms, includeOverlay)

				if err != nil && loadAttempts == maxLoadAttempts {
					// We've exhausted our retries. Fail hard.
					log.Fatalln(err)
				} else if err != nil {
					log.Println("Sleeping 5s to recover from", err.Error(), ". Attempt #", loadAttempts)
					time.Sleep(5 * time.Second)
					continue
				}

				// If no error, we're done
				break
			}
			<-sem

		}(zipFile, dicomList)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func getZipMap(manifest string) (map[string][]string, error) {
	f, err := os.Open(manifest)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = '\t'
	entries, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	zipFileCol, dicomFileCol := -1, -1

	// First, identify whether we are extracting multiple images from any zips.
	// If so, it will be more efficient to open the zip one time and extract the
	// desired images, rather than opening/closing the zip for each image
	// (especially if over gcsfuse)
	zipMap := make(map[string][]string) // map[zip_file][]dicom_file
	for i, row := range entries {
		if i == 0 {
			for j, col := range row {
				if col == "zip_file" {
					zipFileCol = j
				} else if col == "dicom_file" {
					dicomFileCol = j
				}
			}

			continue
		} else if zipFileCol < 0 || dicomFileCol < 0 {
			return nil, fmt.Errorf("Did not identify zip_file or dicom_file in the header line of %s", manifest)
		}

		// Append to this zip file's list of individual dicom images to process
		zipMap[row[zipFileCol]] = append(zipMap[row[zipFileCol]], row[dicomFileCol])
	}

	return zipMap, nil
}

// ProcessOneZipFile prints out PNGs for all DICOM images within a zip file that
// are found in dicomList. TODO: handle errors more thoughtfully, e.g., with an
// error channel.
func ProcessOneZipFile(inputPath, outputPath, zipName string, dicomList []string, includeOverlay bool) error {

	// Note: uses global client, which is OK for concurrent use
	path := filepath.Join(inputPath, zipName)
	if strings.HasPrefix(inputPath, "gs://") {
		path = inputPath
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		path += zipName
	}

	f, nBytes, err := bulkprocess.MaybeOpenFromGoogleStorage(path, client)
	if err != nil {
		return fmt.Errorf("ProcessOneZipFile fatal error (terminating on zip %s): %v", zipName, err)
	}
	defer f.Close()

	// The zip is now open, so we don't have to reopen/reclose for every dicom
	for _, dicomName := range dicomList {

		err := func(dicom string) error {

			// TODO: Can consider optimizing further. This is an o(n^2)
			// operation - but if you printed the Dicom image to PNG within this
			// function, you could make it accept a map and then only iterate in
			// o(n) time. Not sure this is a bottleneck yet.
			img, err := bulkprocess.ExtractDicomFromReaderAt(f, nBytes, dicomName, includeOverlay)
			if err != nil {
				return err
			}

			return imgToPNG(img, outputPath, dicomName)
		}(dicomName)

		// Since we want to continue processing the zip even if one of its
		// dicoms was bad, simply log the error. However, if the underlying
		// file system is unreliable, we should then exit.
		if err != nil && strings.Contains(err.Error(), "input/output error") {

			// If we had an error due to an unreliable filesystem, we need to
			// fail the whole job or we will end up with unreliable or missing
			// data.
			return fmt.Errorf("ProcessOneZipFile fatal error (terminating on dicom %s in zip %s): %v", dicomName, zipName, err)

		} else if err != nil {

			// Non-i/o errors usually indicate that this is just a bad internal
			// file that will never be readable, in which case we should move on
			// rather than quitting.
			log.Printf("ProcessOneZipFile error (skipping dicom %s in zip %s): %v\n", dicomName, zipName, err)

		}
	}

	return nil
}

func imgToPNG(img image.Image, outputPath, dicomName string) error {

	// Will output an RGBA image since that's apparently easier for FastAI to work with
	size := img.Bounds().Size()
	rect := image.Rect(0, 0, size.X, size.Y)
	colImg := image.NewRGBA(rect)
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			pixel := img.At(x, y)

			colImg.Set(x, y, color.RGBA64Model.Convert(pixel))
		}
	}

	f, err := os.Create(outputPath + "/" + dicomName + ".png")
	if err != nil {
		return err
	}
	defer f.Close()

	// Use a buffered writer in case we end up writing to a high-latency disk
	// such as gcsfuse
	fw := bufio.NewWriter(f)

	if err := png.Encode(fw, colImg); err != nil {
		return err
	}

	fw.Flush()

	return nil
}
