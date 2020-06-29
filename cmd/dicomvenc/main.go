package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/carbocation/pfx"
	"gonum.org/v1/gonum/stat"

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

	var inputPath, maskPath, configPath string

	flag.StringVar(&inputPath, "file", "", "Path to the local DICOM file.")
	flag.StringVar(&maskPath, "mask", "", "Path to the local mask file, in the form of an encoded PNG.")
	flag.StringVar(&configPath, "config", "", "Path to the config.json file, to interpret the pixel mask meaning.")

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

	// Print the header
	fmt.Println(strings.Join([]string{
		"dicom",
		"label_id",
		"label_name",
		"pixels",
		"area_cm2",
		"flow_cm3_sec",
		"vti_mean_cm",
		"vti_99pct_cm",
		"vti_100pct_cm",
		"flow_abs_cm3_sec",
		"velocity_mean_cm_sec",
		"velocity_stdev_cm_sec",
		"velocity_min_cm_sec",
		"velocity_01pct_cm_sec",
		"velocity_99pct_cm_sec",
		"velocity_max_cm_sec",
		"venc_limit",
		"duration_sec",
	}, "\t"))

	// Do the work
	err = run(inputPath, maskPath, config)

	if err != nil {
		log.Fatalln(err)
	}

}

type vencPixel struct {
	PixelNumber int
	FlowVenc    float64
}

func run(inputPath, maskPath string, config overlay.JSONConfig) error {

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

	// Get the duration of time for this frame. Needed to infer VTI.
	dt, err := deltaT(tagMap)
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

		// Get a measure of dispersion of velocity estimates across the pixels
		pixelVelocities := make([]float64, 0, len(v))
		for _, px := range v {
			pixelVelocities = append(pixelVelocities, px.FlowVenc)
		}
		velStDev := stat.StdDev(pixelVelocities, nil)

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

		// Get VTI (units are cm/contraction; here, just unitless cm) This is
		// the integral of venc (cm/sec) over the unit of time (portion of a
		// second). Specifically, we want a peak velocity. To try to reduce
		// error, will take 99% value rather than the very top pixel. Since this
		// is directional, will assess the mean velocity first, and then take
		// the extremum that is directionally consistent with the bulk flow.
		sort.Float64Slice(pixelVelocities).Sort()
		lowVel := stat.Quantile(0.01, stat.LinInterp, pixelVelocities, nil)
		highVel := stat.Quantile(0.99, stat.LinInterp, pixelVelocities, nil)

		vti99pct := dt * highVel
		vtimax := dt * maxPix

		if sum < 0 {
			vti99pct = dt * lowVel
			vtimax = dt * minPix
		}

		vtimean := dt * sum / float64(len(v))

		// Convert to units of "cm^3 / sec"
		absFlow := absSum * pxHeightCM * pxWidthCM
		flow := sum * pxHeightCM * pxWidthCM

		fmt.Printf("%s\t%d\t%s\t%d\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\n",
			filepath.Base(inputPath),
			label.ID,
			strings.ReplaceAll(label.Label, " ", "_"),
			len(v),
			float64(len(v))*pxHeightCM*pxWidthCM,
			flow,
			vtimean,
			vti99pct,
			vtimax,
			absFlow,
			sum/float64(len(v)),
			velStDev,
			minPix,
			lowVel,
			highVel,
			maxPix,
			flowVenc.FlowVenc,
			dt,
		)
	}

	return nil
}

// deltaT returns the assumed amount of time (in seconds) belonging to each
// image in a cycle. It is the entire duration that it takes to acquire all
// images (N), divided by N. Here we assume it is the same for every image in a
// sequence - in all inspected images, this seems to be true.
func deltaT(tagMap map[dicomtag.Tag][]interface{}) (deltaTMS float64, err error) {

	nominalInterval, cardiacNumberOfImages := 0.0, 0.0

	// nominalInterval
	val, exists := tagMap[dicomtag.NominalInterval]
	if !exists {
		return 0, pfx.Err(fmt.Errorf("NominalInterval not found"))
	}
	for _, v := range val {
		nominalInterval, err = strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return 0, pfx.Err(err)
		}
		break
	}

	// cardiacNumberOfImages
	val, exists = tagMap[dicomtag.CardiacNumberOfImages]
	if !exists {
		return 0, pfx.Err(fmt.Errorf("CardiacNumberOfImages not found"))
	}
	for _, v := range val {
		cardiacNumberOfImages, err = strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return 0, pfx.Err(err)
		}
		break
	}

	// Convert milliseconds to seconds
	return (1 / 1000.0) * nominalInterval / cardiacNumberOfImages, nil
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

	// They're in mm -- convert to cm
	pxHeightCM *= 0.1
	pxWidthCM *= 0.1

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
