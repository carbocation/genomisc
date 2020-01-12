package main

import (
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

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

func main() {

	fmt.Fprintf(os.Stderr, "This dicom2png binary was built at: %s\n", builddate)

	var inputPath, outputPath, manifest string
	var includeOverlay bool
	flag.StringVar(&inputPath, "raw", "", "Path to the local folder containing the raw zip files")
	flag.StringVar(&outputPath, "out", "", "Path to the local folder where the extracted PNGs will go")
	flag.StringVar(&manifest, "manifest", "", "Manifest file containing Zip names and Dicom names.")
	flag.BoolVar(&includeOverlay, "include-overlay", true, "Print the overlay on top of the images?")

	flag.Parse()
	if inputPath == "" || outputPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
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
			ProcessOneZipFile(inputPath, outputPath, zipFile, dicoms, includeOverlay)
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

func ProcessOneZipFile(inputPath, outputPath, zipName string, dicomList []string, includeOverlay bool) {

	f, err := os.Open(filepath.Join(inputPath, zipName))
	if err != nil {
		log.Printf("Skipping zip %s due to err: %s", zipName, err)
		return
	}
	defer f.Close()

	// the zip reader wants to know the # of bytes in advance
	nBytes, err := f.Stat()
	if err != nil {
		log.Printf("Skipping zip %s due to err: %s", zipName, err)
		return
	}

	// The zip is now open, so we don't have to reopen/reclose for every dicom
	for _, dicomName := range dicomList {
		err := func(dicom string) error {
			// TODO: Can consider optimizing further. This is an o(n^2)
			// operation - but if you printed the Dicom image to PNG within this
			// function, you could make it accept a map and then only iterate in
			// o(n) time. Not sure this is a bottleneck yet.
			img, err := bulkprocess.ExtractDicomFromReaderAt(f, nBytes.Size(), dicomName, includeOverlay)
			if err != nil {
				return err
			}

			return imgToPNG(img, outputPath, dicomName)
		}(dicomName)

		// Since we want to continue processing the zip even if one of its
		// dicoms was bad, simply log the error.
		if err != nil {
			log.Printf("%v: skipping dicom %s in zip %s", err, dicomName, zipName)
		}
	}
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

	if err := png.Encode(f, colImg); err != nil {
		return err
	}

	return nil
}
