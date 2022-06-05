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

	_ "github.com/carbocation/genomisc/compileinfoprint"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

func main() {

	var inputPath, outputPath, scaling, imageType string
	var includeOverlay bool
	flag.StringVar(&inputPath, "file", "", "Path to the local DICOM file. If this points to a folder, all .dcm files in the folder will be converted")
	flag.StringVar(&outputPath, "out", "", "Path to the local folder where the extracted PNGs will go")
	flag.StringVar(&scaling, "scaling", "official", "Pixel intensity scaling to use. Can be official (DICOM standard), pythonic (scaled to the max observed pixel value), or raw")
	flag.StringVar(&imageType, "imagetype", "PNGRGBA", "Options include PNGRGBA (default) or PNGGray16")
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

	// FuncOpts
	opts := make([]func(*bulkprocess.ExtractDicomOptions), 0)
	if includeOverlay {
		opts = append(opts, bulkprocess.OptIncludeOverlay())
	}
	switch scaling {
	case "pythonic":
		opts = append(opts, bulkprocess.OptWindowScalingPythonic())
	case "raw":
		opts = append(opts, bulkprocess.OptWindowScalingRaw())
	default:
		opts = append(opts, bulkprocess.OptWindowScalingOfficial())
	}

	if fileInfo.IsDir() {
		err = runDir(inputPath, outputPath, imageType, opts)
	} else {
		err = run(inputPath, outputPath, imageType, opts)
	}

	if err != nil {
		log.Fatalln(err)
	}

}

func run(inputPath, outputPath, imageType string, opts []func(*bulkprocess.ExtractDicomOptions)) error {

	f, err := os.Open(inputPath)
	if err != nil {
		return pfx.Err(err)
	}
	defer f.Close()

	size, err := f.Stat()
	if err != nil {
		return err
	}

	img, err := bulkprocess.ExtractDicomFromReaderFuncOp(f, size.Size(), opts...)
	if err != nil {
		return err
	}

	switch imageType {
	case "PNGGray16":
		// this is the native output format from the DICOM extraction library
		return imgToPNG(img, outputPath, filepath.Base(inputPath))
	default:
		// PNGRGBA
		return imgToPNGRGBA(img, outputPath, filepath.Base(inputPath))
	}
}

func runDir(inputPath, outputPath, imageType string, opts []func(*bulkprocess.ExtractDicomOptions)) error {
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
			if err := run(inputPath+"/"+filename, outputPath, imageType, opts); err != nil {
				log.Println(err)
			}
			<-sem
		}(file.Name())
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func imgToPNGRGBA(img image.Image, outputPath, dicomName string) error {

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

	return imgToPNG(colImg, outputPath, dicomName)
}

func imgToPNG(img image.Image, outputPath, dicomName string) error {
	f, err := os.Create(outputPath + "/" + dicomName + ".png")
	if err != nil {
		return err
	}

	if err := png.Encode(f, img); err != nil {
		return err
	}

	return nil
}
