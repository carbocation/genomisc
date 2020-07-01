package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aybabtme/uniplot/histogram"
	"github.com/carbocation/genomisc/overlay"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

// Special value that is to be set using ldflags
// E.g.: go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"
// Consider aliasing in .profile: alias gobuild='go build -ldflags "-X main.builddate=`date -u +%Y-%m-%d:%H:%M:%S%Z`"'
var builddate string

func main() {

	fmt.Fprintf(os.Stderr, "This dicomvenc binary was built at: %s\n", builddate)

	var inputPath, maskPath, configPath, manifest, zipPath, maskFolder, maskSuffix string

	flag.StringVar(&manifest, "vencmanifest", "", "(Optional) VENC-style mapped manifest file containing Zip names and Dicom names. If provided, --zips and --out are required and --file and --mask will be ignored.")
	flag.StringVar(&zipPath, "zips", "", "(Required if --manifest is set) Path to the local folder containing the raw UK Biobank zip files")
	flag.StringVar(&maskFolder, "maskfolder", "", "(Required if --manifest is set) Path to the local folder containing the matching mask files")
	flag.StringVar(&maskSuffix, "masksuffix", ".png.mask.png", "(Required if --manifest is set) Suffix the place after the masks's .dcm name")
	flag.StringVar(&inputPath, "file", "", "Path to the local DICOM file.")
	flag.StringVar(&maskPath, "mask", "", "Path to the local mask file, in the form of an encoded PNG.")
	flag.StringVar(&configPath, "config", "", "Path to the config.json file, to interpret the pixel mask meaning.")

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

	if err != nil {
		log.Fatalln(err)
	}

	if manifest != "" {
		// Parse from zip files
		err = runFromManifest(manifest, zipPath, maskFolder, maskSuffix, config)
	} else {
		// Run one file
		err = runFromFiles(inputPath, maskPath, config)
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func runFromManifest(manifest, zipPath, maskFolder, maskSuffix string, config overlay.JSONConfig) error {
	zipMap, err := getZipMap(manifest)
	if err != nil {
		return err
	}

	// Now iterate over the zips and print out the requested dicoms as images

	concurrency := 1 //runtime.NumCPU()
	sem := make(chan bool, concurrency)

	for zipFile, dicomList := range zipMap {
		sem <- true
		go func(zipFile string, dicoms []maskMap) {

			// The main purpose of this loop is to handle a specific filesystem
			// error (input/output error) that largely happens with GCSFuse, and
			// retry a few times before giving up.
			for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

				err := processOneZipFile(zipPath, zipFile, dicoms, maskFolder, maskSuffix, config)

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

func processOneZipFile(zipPath, zipFile string, dicoms []maskMap, maskFolder, maskSuffix string, config overlay.JSONConfig) error {
	f, err := os.Open(filepath.Join(zipPath, zipFile))
	if err != nil {
		return fmt.Errorf("ProcessOneZipFile fatal error (terminating on zip %s): %v", zipFile, err)
	}
	defer f.Close()

	// the zip reader wants to know the # of bytes in advance
	nBytes, err := f.Stat()
	if err != nil {
		return fmt.Errorf("ProcessOneZipFile fatal error (terminating on zip %s): %v", zipFile, err)
	}

	// The zip is now open, so we don't have to reopen/reclose for every dicom
	for _, dicomNames := range dicoms {
		err := func(dicomPair maskMap) error {

			// TODO: Can consider optimizing further. This is an o(n^2)
			// operation - but if you printed the Dicom image to PNG within this
			// function, you could make it accept a map and then only iterate in
			// o(n) time. Not sure this is a bottleneck yet.
			dcmReadSeeker, err := extractDicomReaderFromZip(f, nBytes.Size(), dicomPair.VencDicom)
			if err != nil {
				return err
			}

			// Load the mask as an image. The mask comes from the cine dicom
			// while we will apply it to the phase (VENC) dicom.
			maskPath := filepath.Join(maskFolder, dicomPair.CineDicom+maskSuffix)
			rawOverlayImg, err := overlay.OpenImageFromLocalFile(maskPath)
			if err != nil {
				return err
			}

			return run(dcmReadSeeker, rawOverlayImg, dicomNames.VencDicom, config)
		}(dicomNames)

		// Since we want to continue processing the zip even if one of its
		// dicoms was bad, simply log the error. However, if the underlying
		// file system is unreliable, we should then exit.
		if err != nil && strings.Contains(err.Error(), "input/output error") {

			// If we had an error due to an unreliable filesystem, we need to
			// fail the whole job or we will end up with unreliable or missing
			// data.
			return fmt.Errorf("ProcessOneZipFile fatal error (terminating on dicom %s in zip %s): %v", dicomNames.VencDicom, zipFile, err)

		} else if err != nil {

			// Non-i/o errors usually indicate that this is just a bad internal
			// file that will never be readable, in which case we should move on
			// rather than quitting.
			log.Printf("ProcessOneZipFile error (skipping dicom %s in zip %s): %v\n", dicomNames.VencDicom, zipFile, err)

		}
	}

	return nil
}

func extractDicomReaderFromZip(zipReaderAt io.ReaderAt, zipNBytes int64, dicomName string) (*bytes.Reader, error) {
	var err error

	rc, err := zip.NewReader(zipReaderAt, zipNBytes)
	if err != nil {
		return nil, err
	}

	for _, v := range rc.File {
		// Iterate over all of the dicoms in the zip til we find the one with
		// the desired name
		if v.Name != dicomName {
			continue
		}

		dcmReadCloser, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer dcmReadCloser.Close()

		// Convert our readCloser to a readSeeker
		dcmBytes, err := ioutil.ReadAll(dcmReadCloser)
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(dcmBytes), nil

	}

	return nil, fmt.Errorf("Did not find the requested Dicom %s", dicomName)
}

type maskMap struct {
	VencDicom string
	CineDicom string
}

func getZipMap(manifest string) (map[string][]maskMap, error) {
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

	zipFileCol, vencDicomCol, cineDicomCol := -1, -1, -1

	// First, identify whether we are extracting multiple images from any zips.
	// If so, it will be more efficient to open the zip one time and extract the
	// desired images, rather than opening/closing the zip for each image
	// (especially if over gcsfuse)
	zipMap := make(map[string][]maskMap) // map[zip_file][]dicom_file
	for i, row := range entries {
		if i == 0 {
			for j, col := range row {
				if col == "zip_file" {
					zipFileCol = j
				} else if col == "dicom_cine" {
					cineDicomCol = j
				} else if col == "dicom_venc" {
					vencDicomCol = j
				}
			}

			continue
		} else if zipFileCol < 0 || vencDicomCol < 0 || cineDicomCol < 0 {
			return nil, fmt.Errorf("Did not identify zip_file, dicom_cine, or dicom_venc in the header line of %s", manifest)
		}

		// Append to this zip file's list of individual dicom images to process
		zipMap[row[zipFileCol]] = append(zipMap[row[zipFileCol]], maskMap{CineDicom: row[cineDicomCol], VencDicom: row[vencDicomCol]})
	}

	return zipMap, nil
}

func runFromFiles(inputPath, maskPath string, config overlay.JSONConfig) error {
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
	return run(f, rawOverlayImg, filepath.Base(inputPath), config)
}

type vencPixel struct {
	PixelNumber int
	FlowVenc    float64
}

func run(f io.ReadSeeker, rawOverlayImg image.Image, dicomName string, config overlay.JSONConfig) error {

	cols := rawOverlayImg.Bounds().Size().X
	rows := rawOverlayImg.Bounds().Size().Y

	// Will store pixels linked with each segmentation class
	segmentPixels := make(map[uint][]vencPixel)

	meta, err := bulkprocess.DicomToMetadata(f)
	if err != nil {
		return err
	}

	// Reset the DICOM reader
	f.Seek(0, 0)

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

	// Need Siemens header data, if it exists
	phaseContrastN4 := "NO"
	velocityEncodingDirectionN4 := 0.0
	if elem, exists := tagMap[dicomtag.Tag{Group: 0x0029, Element: 0x1010}]; exists {
		for _, headerRow := range elem {
			sc, err := bulkprocess.ParseSiemensHeader(headerRow)
			if err != nil {
				return err
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
							return err
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

			// Apply the "sign" immediately at the pixel level
			seg.FlowVenc = float64(sign) * flowVenc.PixelIntensityToVelocity(float64(frame.NativeData.Data[j][0]))

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

		pixdat := describeSegmentationPixels(v, dt, pxHeightCM, pxWidthCM)

		// Are we potentially aliasing, based on how close we are getting to an extremum of +/- FlowVenc?
		aliasRisk := math.Abs(pixdat.PixelVelocityMaxCMPerSec) > 0.99*flowVenc.FlowVenc ||
			math.Abs(pixdat.PixelVelocityMinCMPerSec) > 0.99*flowVenc.FlowVenc

		unwrapped := false

		if aliasRisk {
			// Generate a histogram. The number of buckets is arbitrary. TODO:
			// find a rational bucket count.
			hist := histogram.Hist(25, pixdat.PixelVelocities)

			if pixdat.PixelVelocitySumCMPerSec > 0 {
				// If the bulk flow is positive, start at the most negative
				// value:

				zeroBucket := false
				wrapPoint := math.Inf(-1)
				for _, bucket := range hist.Buckets {
					wrapPoint = bucket.Max
					if bucket.Count == 0 {
						zeroBucket = true
						break
					}
				}

				// If we never saw a zero bucket, there might not actually be
				// aliasing - the full range might be used
				if zeroBucket {
					unwrappedV := make([]vencPixel, len(v))
					copy(unwrappedV, v)
					for k, vp := range unwrappedV {
						if unwrappedV[k].FlowVenc < wrapPoint {
							unwrappedV[k].FlowVenc = 2.0*flowVenc.FlowVenc + vp.FlowVenc
						}
					}

					// And replace our stats
					pixdat = describeSegmentationPixels(unwrappedV, dt, pxHeightCM, pxWidthCM)
					unwrapped = true
				}
			} else if pixdat.PixelVelocitySumCMPerSec < 0 {
				// If the flow is negative, reverse the sort, so we can start at
				// the highest values and go downward.

				zeroBucket := false
				wrapPoint := math.Inf(1)
				sort.Slice(hist.Buckets, func(i, j int) bool {
					return hist.Buckets[i].Min < hist.Buckets[j].Min
				})

				for _, bucket := range hist.Buckets {
					wrapPoint = bucket.Min
					if bucket.Count == 0 {
						zeroBucket = true
						break
					}
				}

				// If we never saw a zero bucket, there might not actually be
				// aliasing - the full range might be used
				if zeroBucket {
					unwrappedV := make([]vencPixel, len(v))
					copy(unwrappedV, v)
					for k, vp := range unwrappedV {
						if unwrappedV[k].FlowVenc > wrapPoint {
							unwrappedV[k].FlowVenc = -2.0*flowVenc.FlowVenc + vp.FlowVenc
						}
					}

					// And replace our stats
					pixdat = describeSegmentationPixels(unwrappedV, dt, pxHeightCM, pxWidthCM)
					unwrapped = true
				}
			}

			// fmt.Printf("%+v\n", hist.Buckets)
			// if err := histogram.Fprint(os.Stdout, hist, histogram.Linear(5)); err != nil {
			// 	panic(err)
			// }
		}

		fmt.Printf("%s\t%d\t%s\t%d\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%.5g\t%s\t%s\t%.5g\t%t\t%t\t%t\n",
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
		)
	}

	return nil
}
