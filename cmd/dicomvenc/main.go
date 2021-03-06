package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/carbocation/genomisc/overlay"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

type vencPixel struct {
	PixelNumber int
	FlowVenc    float64
}

var (
	BufferSize = 4096 * 8
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

func main() {
	defer STDOUT.Flush()

	fmt.Fprintf(os.Stderr, "This dicomvenc binary was built at: %s\n", builddate)

	var inputPath, maskPath, configPath, manifest, zipPath, maskFolder, maskSuffix, plot string

	flag.StringVar(&manifest, "vencmanifest", "", "(Optional) VENC-style mapped manifest file containing Zip names and Dicom names. If provided, --zips and --out are required and --file and --mask will be ignored.")
	flag.StringVar(&zipPath, "zips", "", "(Required if --manifest is set) Path to the local folder containing the raw UK Biobank zip files")
	flag.StringVar(&maskFolder, "maskfolder", "", "(Required if --manifest is set) Path to the local folder containing the matching mask files")
	flag.StringVar(&maskSuffix, "masksuffix", ".png.mask.png", "(Required if --manifest is set) Suffix the place after the masks's .dcm name")
	flag.StringVar(&inputPath, "file", "", "Path to the local DICOM file.")
	flag.StringVar(&maskPath, "mask", "", "Path to the local mask file, in the form of an encoded PNG.")
	flag.StringVar(&configPath, "config", "", "Path to the config.json file, to interpret the pixel mask meaning.")
	flag.StringVar(&plot, "plot", "", "(Optional) Produce a plot? Accepts 'vti' or ''.")

	flag.Parse()

	fmt.Fprintln(os.Stderr, strings.Join(os.Args, " "))

	// Must pass a manifest or the raw dicom and mask path
	if manifest == "" && (inputPath == "" || maskPath == "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// If a manifest is passed, we need to know where the zips are, where the
	// masks are, and where the output should go
	if manifest != "" && (zipPath == "" || maskFolder == "" || maskSuffix == "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Always need a config file
	if configPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	config, err := overlay.ParseJSONConfigFromPath(configPath)
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	// Print the header
	if plot == "" {
		fmt.Println(strings.Join([]string{
			"dicom",
			"label_id",
			"label_name",
			"pixels",
			"area_cm2",
			"flow_cm3_sec",
			"volume_cm3",
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
			"instance_number",
			"phase_contrast_n4",
			"velocity_encoding_direction_n4",
			"venc_z_axis_sign_flipped",
			"aliasing_risk",
			"was_unwrapped",
		}, "\t"))
	}

	if err != nil {
		log.Fatalln(err)
	}

	if plot != "" {
		err = plotFromManifest(manifest, zipPath, maskFolder, maskSuffix, config)
	} else if manifest != "" {
		// Parse from zip files
		err = runFromManifest(manifest, zipPath, maskFolder, maskSuffix, config)
	} else {
		// Run one file
		err = runFromFile(inputPath, maskPath, config)
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func runFromFile(inputPath, maskPath string, config overlay.JSONConfig) error {
	// Load the dicom as a reader
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Load the mask as an image
	rawOverlayImg, err := overlay.OpenImageFromLocalFile(maskPath)
	if err != nil {
		return err
	}

	// Do the work
	str, err := run(f, rawOverlayImg, filepath.Base(inputPath), config)
	if err != nil {
		return err
	}
	fmt.Print(str)
	return nil
}

func run(f io.ReadSeeker, rawOverlayImg image.Image, dicomName string, config overlay.JSONConfig) (string, error) {

	cols := rawOverlayImg.Bounds().Size().X
	rows := rawOverlayImg.Bounds().Size().Y

	// Will store pixels linked with each segmentation class
	segmentPixels := make(map[uint][]vencPixel)

	meta, err := bulkprocess.DicomToMetadata(f)
	if err != nil {
		return "", err
	}

	// Reset the DICOM reader
	f.Seek(0, 0)

	// Make the DICOM fields addressable as a map
	tagMap, err := bulkprocess.DicomToTagMap(f)
	if err != nil {
		return "", err
	}

	// Load VENC data
	flowVenc, err := fetchFlowVenc(tagMap)
	if err != nil {
		return "", pfx.Err(err)
	}

	// Load the DICOM pixel data
	pixelElem, exists := tagMap[dicomtag.PixelData]
	if !exists {
		return "", fmt.Errorf("PixelData not found")
	}

	pxHeightCM, pxWidthCM, err := pixelHeightWidthCM(tagMap)
	if err != nil {
		return "", err
	}

	// Get the duration of time for this frame. Needed to infer VTI.
	dt, err := deltaT(tagMap)
	if err != nil {
		return "", err
	}

	// Need Siemens header data, if it exists
	phaseContrastN4 := "NO"
	velocityEncodingDirectionN4 := 0.0
	if elem, exists := tagMap[dicomtag.Tag{Group: 0x0029, Element: 0x1010}]; exists {
		for _, headerRow := range elem {
			sc, err := bulkprocess.ParseSiemensHeader(headerRow)
			if err != nil {
				return "", err
			}
			for _, v := range sc.Slice() {
				if v.Name == "PhaseContrastN4" {
					for _, subE := range v.SubElementData {
						phaseContrastN4 = subE
					}
				}
				if v.Name == "VelocityEncodingDirectionN4" && phaseContrastN4 == "YES" {
					for axis, value := range v.SubElementData {
						if axis != 2 {
							continue
						}
						velocityEncodingDirectionN4, err = strconv.ParseFloat(value, 64)
						if err != nil {
							return "", err
						}
					}
				}
			}
		}
	}

	// Not 100% clear to me how to interpret the N4 fields, but they aren't
	// always there, and they clearly define a +/- directional orientation. When
	// the Z value (1-based 3rd value) from VelocityEncodingDirectionN4 is
	// positive, it seems to indicate that the normal Z direction is reversed.
	// In the normal orientation, positive X is toward the participant's left,
	// positive Y is toward the participant's posterior, and positive Z is
	// toward the participant's head. When VelocityEncodingDirectionN4 is
	// positive, then a positive venc appears to be toward the feet rather than
	// toward the head.
	sign := 1
	if phaseContrastN4 == "YES" && velocityEncodingDirectionN4 > 0 {
		sign = -1
	}

	// Iterate over the DICOM and find all pixels for each class and their VENC
	// values
	data := pixelElem[0].(element.PixelDataInfo)
	for _, frame := range data.Frames {
		if frame.IsEncapsulated() {
			return "", fmt.Errorf("Frame is encapsulated, which we did not expect")
		}

		// Ensure that the pixels from the DICOM are in agreement with the
		// pixels from the mask.
		if x, y := len(frame.NativeData.Data), rows*cols; x != y {
			return "", fmt.Errorf("DICOM data has %d pixels but mask data has %d pixels (%d rows and %d cols)", x, y, rows, cols)
		}

		// Iterate over the DICOM pixels
		for j := 0; j < len(frame.NativeData.Data); j++ {

			// Identify which segmentation class this pixel belongs to
			idAtPixel, err := overlay.LabeledPixelToID(rawOverlayImg.At(j%cols, j/cols))
			if err != nil {
				return "", err
			}

			// Extract the pixel's VENC
			seg := vencPixel{}
			seg.PixelNumber = j

			// Apply the "sign" immediately at the pixel level
			seg.FlowVenc = float64(sign) * flowVenc.PixelIntensityToVelocity(float64(frame.NativeData.Data[j][0]))

			// Save the pixel to the class's pixel map
			segmentPixels[uint(idAtPixel)] = append(segmentPixels[uint(idAtPixel)], seg)
		}
	}

	// Print out summary data for each segmentation class.
	sb := strings.Builder{}
	for _, label := range config.Labels.Sorted() {
		v, exists := segmentPixels[label.ID]
		if !exists {
			continue
		}

		pixdat := describeSegmentationPixels(v, dt, pxHeightCM, pxWidthCM)

		// Are we potentially aliasing, based on how close we are getting to an
		// extremum of +/- FlowVenc?
		aliasRisk := math.Abs(pixdat.PixelVelocityMaxCMPerSec) > 0.99*flowVenc.FlowVenc ||
			math.Abs(pixdat.PixelVelocityMinCMPerSec) > 0.99*flowVenc.FlowVenc

		unwrapped := false
		if aliasRisk {
			pixdat, _, unwrapped = deAlias(pixdat, flowVenc, dt, pxHeightCM, pxWidthCM, v)
		}

		sb.WriteString(fmt.Sprintf("%s\t%d\t%s\t%d\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%s\t%s\t%.5g\t%t\t%t\t%t\n",
			dicomName,
			label.ID,
			strings.ReplaceAll(label.Label, " ", "_"),
			len(v),
			float64(len(v))*pxHeightCM*pxWidthCM,
			pixdat.FlowCM3PerSec,
			pixdat.FlowCM3PerSec*dt,
			pixdat.VTIMeanCM,
			pixdat.VTI99pctCM,
			pixdat.VTIMaxCM,
			pixdat.AbsFlowCM3PerSec,
			pixdat.PixelVelocitySumCMPerSec/float64(len(pixdat.PixelVelocities)),
			pixdat.PixelVelocityVelocityStandardDeviationCMPerSec,
			pixdat.PixelVelocityMinCMPerSec,
			pixdat.PixelVelocity01PctCMPerSec,
			pixdat.PixelVelocity99PctCMPerSec,
			pixdat.PixelVelocityMaxCMPerSec,
			flowVenc.FlowVenc,
			dt,
			meta.InstanceNumber,
			phaseContrastN4,
			velocityEncodingDirectionN4,
			sign < 0,
			aliasRisk,
			unwrapped,
		))
	}

	return sb.String(), nil
}
