package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/carbocation/pfx"

	"github.com/carbocation/genomisc/overlay"
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

	var inputPath, maskPath, configPath, outputPath string

	flag.StringVar(&inputPath, "file", "", "Path to the local DICOM file. If this points to a folder, all .dcm files in the folder will be converted")
	flag.StringVar(&maskPath, "mask", "", "Path to the local mask file, in the form of an encoded PNG.")
	flag.StringVar(&configPath, "config", "", "Path to the config.json file, to interpret the pixel mask meaning.")
	flag.StringVar(&outputPath, "out", "", "XXX Path to the local folder where the extracted PNGs will go")

	flag.Parse()

	fmt.Fprintln(os.Stderr, strings.Join(os.Args, " "))

	if inputPath == "" || maskPath == "" || configPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	config, err := overlay.ParseJSONConfigFromPath(configPath)
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		log.Fatalln(err)
	}

	if fileInfo.IsDir() {
		err = runDir(inputPath, outputPath, false)
	} else {
		err = run(inputPath, maskPath, outputPath, config)
	}

	if err != nil {
		log.Fatalln(err)
	}

}

type vencPixel struct {
	PixelNumber int
	FlowVenc    float64
}

func run(inputPath, maskPath, outputPath string, config overlay.JSONConfig) error {

	// Load the overlay mask
	rawOverlayImg, err := overlay.OpenImageFromLocalFile(maskPath)
	if err != nil {
		return err
	}
	cols := rawOverlayImg.Bounds().Size().X
	rows := rawOverlayImg.Bounds().Size().Y

	// Will store pixels linked with each segmentation class
	segmentPixels := make(map[uint][]vencPixel)

	// Load the DICOM
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Make the DICOM fields addressable as a map
	tagMap, err := bulkprocess.DicomToTagMap(f)
	if err != nil {
		return err
	}

	// Load VENC data
	flowVenc, err := fetchFlowVenc(tagMap)
	if err != nil {
		return pfx.Err(err)
	}

	// Load the DICOM pixel data
	pixelElem, exists := tagMap[dicomtag.PixelData]
	if !exists {
		return fmt.Errorf("PixelData not found")
	}

	pxHeightCM, pxWidthCM, err := pixelHeightWidthCM(tagMap)
	if err != nil {
		return err
	}

	// Iterate over the DICOM and find all pixels for each class and their VENC
	// values
	data := pixelElem[0].(element.PixelDataInfo)
	for _, frame := range data.Frames {
		if frame.IsEncapsulated {
			return fmt.Errorf("Frame is encapsulated, which we did not expect")
		}

		// Ensure that the pixels from the DICOM are in agreement with the
		// pixels from the mask.
		if x, y := len(frame.NativeData.Data), rows*cols; x != y {
			return fmt.Errorf("DICOM data has %d pixels but mask data has %d pixels (%d rows and %d cols)", x, y, rows, cols)
		}

		// Iterate over the DICOM pixels
		for j := 0; j < len(frame.NativeData.Data); j++ {

			// Identify which segmentation class this pixel belongs to
			idAtPixel, err := overlay.LabeledPixelToID(rawOverlayImg.At(j%cols, j/cols))
			if err != nil {
				return err
			}

			// Extract the pixel's VENC
			seg := vencPixel{}
			seg.PixelNumber = j
			seg.FlowVenc = flowVenc.PixelIntensityToVelocity(float64(frame.NativeData.Data[j][0]))

			// Save the pixel to the class's pixel map
			segmentPixels[uint(idAtPixel)] = append(segmentPixels[uint(idAtPixel)], seg)
		}
	}

	// Print out summary data for each segmentation class.
	for _, label := range config.Labels.Sorted() {
		v, exists := segmentPixels[label.ID]
		if !exists {
			continue
		}

		sum := 0.0
		absSum := 0.0
		minPix, maxPix := 0.0, 0.0
		for _, px := range v {
			sum += px.FlowVenc
			absSum += math.Abs(px.FlowVenc)
			if px.FlowVenc < minPix {
				minPix = px.FlowVenc
			}
			if px.FlowVenc > maxPix {
				maxPix = px.FlowVenc
			}
		}

		// Convert to units of "cc / sec"
		absSum *= pxHeightCM * pxWidthCM
		sum *= pxHeightCM * pxWidthCM

		log.Printf("%s | Pixels %d | AbsSum VENC %.3g | Sum VENC %.3g (ratio %.3g) | Mean VENC %.3g | Min, Max VENC %.3g, %.3g\n",
			label.Label,
			len(v),
			absSum,
			sum,
			sum/absSum,
			sum/float64(len(v)),
			minPix,
			maxPix)
	}

	return nil
}

func pixelHeightWidthCM(tagMap map[dicomtag.Tag][]interface{}) (pxHeightCM, pxWidthCM float64, err error) {
	val, exists := tagMap[dicomtag.PixelSpacing]
	if !exists {
		return 0, 0, pfx.Err(fmt.Errorf("PixelSpacing not found"))
	}

	for k, v := range val {
		if k == 0 {
			pxHeightCM, err = strconv.ParseFloat(v.(string), 32)
			if err != nil {
				continue
			}
		} else if k == 1 {
			pxWidthCM, err = strconv.ParseFloat(v.(string), 32)
			if err != nil {
				continue
			}
		}
	}

	return
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
			// run(filename, outputPath, includeOverlay)
			<-sem
		}(file.Name())
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}
