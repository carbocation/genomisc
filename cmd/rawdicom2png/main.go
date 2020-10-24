package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

func main() {

	fmt.Fprintf(os.Stderr, "This rawdicom2png binary was built at: %s\n", builddate)

	var inputPath, outputPath string
	var includeOverlay bool
	flag.StringVar(&inputPath, "file", "", "Path to the local DICOM file. If this points to a folder, all .dcm files in the folder will be converted")
	flag.StringVar(&outputPath, "out", "", "Path to the local folder where the extracted PNGs will go")
	flag.BoolVar(&includeOverlay, "include-overlay", true, "Print the overlay on top of the images?")

	flag.Parse()

	fmt.Fprintln(os.Stderr, strings.Join(os.Args, " "))

	if inputPath == "" || outputPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		log.Fatalln(err)
	}

	if fileInfo.IsDir() {
		err = runDir(inputPath, outputPath, includeOverlay)
	} else {
		err = run(inputPath, outputPath, includeOverlay)
	}

	if err != nil {
		log.Fatalln(err)
	}

}

func run(inputPath, outputPath string, includeOverlay bool) error {

	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	size, err := f.Stat()
	if err != nil {
		return err
	}

	img, err := bulkprocess.ExtractDicomFromReader(f, size.Size(), includeOverlay)
	if err != nil {
		return err
	}

	return imgToPNG(img, outputPath, filepath.Base(inputPath))
}

func runDir(inputPath, outputPath string, includeOverlay bool) error {
	dir, err := ioutil.ReadDir(inputPath)
	if err != nil {
		return err
	}

	concurrency := runtime.NumCPU()
	sem := make(chan bool, concurrency)

	for _, file := range dir {
		if !strings.HasSuffix(file.Name(), ".dcm") {
			continue
		}

		sem <- true
		go func(filename string) {
			run(filename, outputPath, includeOverlay)
			<-sem
		}(file.Name())
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
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

	if err := png.Encode(f, colImg); err != nil {
		return err
	}

	return nil
}
