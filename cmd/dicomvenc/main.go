package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"strings"

	"github.com/carbocation/pfx"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

func main() {

	fmt.Fprintf(os.Stderr, "This dicomvenc binary was built at: %s\n", builddate)

	var inputPath, outputPath string
	var includeOverlay bool
	flag.StringVar(&inputPath, "file", "", "Path to the local DICOM file. If this points to a folder, all .dcm files in the folder will be converted")
	flag.StringVar(&outputPath, "out", "", "XXX Path to the local folder where the extracted PNGs will go")
	flag.BoolVar(&includeOverlay, "include-overlay", true, "XXX Print the overlay on top of the images?")

	flag.Parse()

	fmt.Fprintln(os.Stderr, strings.Join(os.Args, " "))

	if inputPath == "" {
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

	tagMap, err := bulkprocess.DicomToTagMap(f)
	if err != nil {
		return err
	}

	flowVenc, err := fetchFlowVenc(tagMap)
	if err != nil {
		return pfx.Err(err)
	}

	log.Printf("%+v\n", flowVenc)

	pixelElem, exists := tagMap[dicomtag.PixelData]
	if !exists {
		return fmt.Errorf("PixelData not found")
	}

	// Main image
	// imgPixels := make([]int, 0)

	data := pixelElem[0].(element.PixelDataInfo)

	for _, frame := range data.Frames {
		if frame.IsEncapsulated {
			return fmt.Errorf("Frame is encapsulated, which we did not expect")
		}

		for j := 0; j < len(frame.NativeData.Data); j++ {
			if j%192 == 0 {
				fmt.Println()
			}

			// imgPixels = append(imgPixels, frame.NativeData.Data[j][0])
			vel := flowVenc.PixelIntensityToVelocity(float64(frame.NativeData.Data[j][0]))
			fmt.Printf("%03.0f ", 250.0+math.Round(vel))
		}
	}

	return nil

	// return imgToPNG(img, outputPath, filepath.Base(inputPath))
}

func fetchFlowVenc(tagMap map[dicomtag.Tag][]interface{}) (*bulkprocess.VENC, error) {

	var bitsStored uint16
	var venc string

	// Bits Stored:
	bs, exists := tagMap[dicomtag.BitsStored]
	if !exists {
		return nil, fmt.Errorf("BitsStored not found")
	}

	for _, v := range bs {
		bitsStored = v.(uint16)
		break
	}

	// Flow VENC:

	// Siemens header data requires special treatment
	ve, exists := tagMap[dicomtag.Tag{Group: 0x0029, Element: 0x1010}]
	if !exists {
		return nil, fmt.Errorf("Siemens header not found")
	}

	for _, v := range ve {
		sc, err := bulkprocess.ParseSiemensHeader(v)
		if err != nil {
			return nil, err
		}

		for _, v := range sc.Slice() {
			if v.Name != "FlowVenc" {
				continue
			}

			for _, encoding := range v.SubElementData {
				venc = encoding
				break
			}
		}
	}

	out, err := bulkprocess.NewVENC(venc, bitsStored)

	return &out, err
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
