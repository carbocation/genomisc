package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/carbocation/genomisc/overlay"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
	hist2 "github.com/grd/histogram"
	"github.com/suyashkumar/dicom/dicomtag"
	"github.com/suyashkumar/dicom/element"
)

func plotFromManifest(manifest, zipPath, maskFolder, maskSuffix string, config overlay.JSONConfig) error {
	zipMap, err := getZipMap(manifest)
	if err != nil {
		return err
	}

	for zipFile, dicomList := range zipMap {
		func(zipFile string, dicoms []maskMap) {

			// The main purpose of this loop is to handle a specific filesystem
			// error (input/output error) that largely happens with GCSFuse, and
			// retry a few times before giving up.
			for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {

				err := plotOneZipFile(zipPath, zipFile, dicoms, maskFolder, maskSuffix, config)

				if err != nil && loadAttempts == maxLoadAttempts {
					// We've exhausted our retries. Fail hard.
					log.Fatalln(err)
				} else if err != nil {
					log.Println("Sleeping 5s to recover from", err.Error(), ". Attempt #", loadAttempts)
					time.Sleep(5 * time.Second)
					continue
				}

				// If no error, we don't need to retry
				break
			}

		}(zipFile, dicomList)
	}

	return nil
}

func plotOneZipFile(zipPath, zipFile string, dicoms []maskMap, maskFolder, maskSuffix string, config overlay.JSONConfig) error {
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

	// [pixel class] => 30 instances of a variable number of values each
	pixelData := make(map[string][]plotStruct)

	// The zip is now open, so we don't have to reopen/reclose for every dicom
	for _, dicomNames := range dicoms {
		dicomPix, err := func(dicomPair maskMap) (map[string]plotStruct, error) {

			// TODO: Can consider optimizing further. This is an o(n^2)
			// operation - but if you printed the Dicom image to PNG within this
			// function, you could make it accept a map and then only iterate in
			// o(n) time. Not sure this is a bottleneck yet.
			dcmReadSeeker, err := extractDicomReaderFromZip(f, nBytes.Size(), dicomPair.VencDicom)
			if err != nil {
				return nil, err
			}

			// Load the mask as an image. The mask comes from the cine dicom
			// while we will apply it to the phase (VENC) dicom.
			maskPath := filepath.Join(maskFolder, dicomPair.CineDicom+maskSuffix)
			rawOverlayImg, err := overlay.OpenImageFromLocalFile(maskPath)
			if err != nil {
				return nil, err
			}

			return runPlot(dcmReadSeeker, rawOverlayImg, dicomNames.VencDicom, config)

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

		// pixelData = append(pixelData, dicomPix)
		for label, v := range dicomPix {
			pixelData[label] = append(pixelData[label], v)
		}
	}

	for lab, pixVals := range pixelData {

		frames := len(pixVals)
		heightMultiplier := 5
		widthMultiplier := 10
		nBins := 100
		outImg := image.NewNRGBA(image.Rect(0, 0, widthMultiplier*frames, heightMultiplier*nBins))
		for y := 0; y < heightMultiplier*nBins; y++ {
			for x := 0; x < widthMultiplier*30; x++ {
				outImg.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}

		// Sort the instances
		sort.Slice(pixVals, func(i, j int) bool { return pixVals[i].InstanceNumber < pixVals[j].InstanceNumber })

		// Establish range
		min := math.MaxFloat64
		max := -1.0
		for _, dicom := range pixVals {
			for _, px := range dicom.Pixels {
				if px < min {
					min = px
				}

				if px > max {
					max = px
				}
			}
		}

		for x, dicom := range pixVals {
			hg, err := hist2.NewHistogram(hist2.Range(min, uint(nBins), (max-min)/float64(nBins)))
			if err != nil {
				return err
			}

			for _, px := range dicom.Pixels {
				hg.Add(px)
			}

			maxCount := 0
			for y := 0; y < nBins; y++ {
				if v := hg.Get(y); v > maxCount {
					maxCount = v
				}
			}

			logCount := math.Log1p(float64(maxCount))

			for y := 0; y < nBins; y++ {
				logRatio := math.Log1p(float64(hg.Get(y))) / logCount

				val := uint8(math.Floor(255 * logRatio))

				for mult := 0; mult < heightMultiplier; mult++ {
					for xmul := 0; xmul < widthMultiplier; xmul++ {
						if val == 0 {
							continue
						}
						outImg.Set(x*widthMultiplier+xmul, y*heightMultiplier+mult, color.NRGBA{R: val, G: val, B: val, A: 255})
					}
				}
			}

			// Mark the zero velocity point
			binWithZero, err := hg.Find(0)
			if err != nil {
				return err
			}
			for mult := 0; mult < heightMultiplier; mult++ {
				for xmul := 0; xmul < widthMultiplier; xmul++ {
					outImg.Set(x*widthMultiplier+xmul, binWithZero*heightMultiplier+mult, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
				}
			}

			// Outline the vmax
			binWith99, err := hg.Find(dicom.V99)
			if err != nil {
				return err
			}
			for mult := 0; mult < heightMultiplier; mult++ {
				for xmul := 0; xmul < widthMultiplier; xmul++ {
					outImg.Set(x*widthMultiplier+xmul, binWith99*heightMultiplier+mult, color.NRGBA{R: 0, G: 255, B: 255, A: 255})
				}
			}

			// Outline the vmax
			sign := 1.0
			if dicom.Vmax > 0 {
				sign = -1.0
			}
			binWithMax, err := hg.Find(dicom.Vmax + 0.0001*sign)
			if err != nil {
				return err
			}
			for mult := 0; mult < heightMultiplier; mult++ {
				for xmul := 0; xmul < widthMultiplier; xmul++ {
					outImg.Set(x*widthMultiplier+xmul, binWithMax*heightMultiplier+mult, color.NRGBA{R: 255, G: 255, B: 0, A: 255})
				}
			}
		}
		f, err := os.Create(zipFile + "_" + strings.ReplaceAll(lab, " ", "_") + ".png")
		if err != nil {
			return err
		}

		if err := png.Encode(f, outImg); err != nil {
			return err
		}
	}

	return nil
}

type plotStruct struct {
	InstanceNumber int
	Vmax           float64
	V99            float64
	Pixels         []float64
}

func runPlot(f io.ReadSeeker, rawOverlayImg image.Image, dicomName string, config overlay.JSONConfig) (map[string]plotStruct, error) {

	meta, err := bulkprocess.DicomToMetadata(f)
	if err != nil {
		return nil, err
	}

	// Reset the DICOM reader
	f.Seek(0, 0)

	cols := rawOverlayImg.Bounds().Size().X
	rows := rawOverlayImg.Bounds().Size().Y

	// Will store pixels linked with each segmentation class
	segmentPixels := make(map[uint][]vencPixel)

	// Make the DICOM fields addressable as a map
	tagMap, err := bulkprocess.DicomToTagMap(f)
	if err != nil {
		return nil, err
	}

	// Load VENC data
	flowVenc, err := fetchFlowVenc(tagMap)
	if err != nil {
		return nil, pfx.Err(err)
	}

	// Load the DICOM pixel data
	pixelElem, exists := tagMap[dicomtag.PixelData]
	if !exists {
		return nil, fmt.Errorf("PixelData not found")
	}

	pxHeightCM, pxWidthCM, err := pixelHeightWidthCM(tagMap)
	if err != nil {
		return nil, err
	}

	// Get the duration of time for this frame. Needed to infer VTI.
	dt, err := deltaT(tagMap)
	if err != nil {
		return nil, err
	}

	// Need Siemens header data, if it exists
	phaseContrastN4 := "NO"
	velocityEncodingDirectionN4 := 0.0
	if elem, exists := tagMap[dicomtag.Tag{Group: 0x0029, Element: 0x1010}]; exists {
		for _, headerRow := range elem {
			sc, err := bulkprocess.ParseSiemensHeader(headerRow)
			if err != nil {
				return nil, err
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
							return nil, err
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
			return nil, fmt.Errorf("Frame is encapsulated, which we did not expect")
		}

		// Ensure that the pixels from the DICOM are in agreement with the
		// pixels from the mask.
		if x, y := len(frame.NativeData.Data), rows*cols; x != y {
			return nil, fmt.Errorf("DICOM data has %d pixels but mask data has %d pixels (%d rows and %d cols)", x, y, rows, cols)
		}

		// Iterate over the DICOM pixels
		for j := 0; j < len(frame.NativeData.Data); j++ {

			// Identify which segmentation class this pixel belongs to
			idAtPixel, err := overlay.LabeledPixelToID(rawOverlayImg.At(j%cols, j/cols))
			if err != nil {
				return nil, err
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

	out := make(map[string]plotStruct)

	// Print out summary data for each segmentation class.
	for _, label := range config.Labels.Sorted() {
		v, exists := segmentPixels[label.ID]
		if !exists {
			continue
		}

		// First, assume no aliasing
		unwrappedV := make([]vencPixel, len(v))
		copy(unwrappedV, v)

		pixdat := describeSegmentationPixels(v, dt, pxHeightCM, pxWidthCM)

		// Are we potentially aliasing, based on how close we are getting to an extremum of +/- FlowVenc?
		aliasRisk := math.Abs(pixdat.PixelVelocityMaxCMPerSec) > 0.99*flowVenc.FlowVenc ||
			math.Abs(pixdat.PixelVelocityMinCMPerSec) > 0.99*flowVenc.FlowVenc

		if aliasRisk {
			pixdat, unwrappedV, _ = deAlias(pixdat, flowVenc, dt, pxHeightCM, pxWidthCM, v)
		}

		entry := plotStruct{}

		entry.Vmax = pixdat.PixelVelocityMaxCMPerSec
		entry.V99 = pixdat.PixelVelocity99PctCMPerSec
		if pixdat.VTIMeanCM < 0 {
			entry.Vmax = pixdat.PixelVelocityMinCMPerSec
			entry.V99 = pixdat.PixelVelocity01PctCMPerSec
		}

		entry.InstanceNumber, err = strconv.Atoi(meta.InstanceNumber)
		if err != nil {
			return nil, err
		}

		outPixels := make([]float64, 0, len(unwrappedV))
		for _, pxVal := range unwrappedV {
			outPixels = append(outPixels, pxVal.FlowVenc)
		}
		entry.Pixels = outPixels

		out[label.Label] = entry
	}

	return out, nil
}
