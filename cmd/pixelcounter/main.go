package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/carbocation/genomisc/overlay"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()

		log.Println("Example JSONConfig file layout:")
		bts, err := json.MarshalIndent(overlay.JSONConfig{Labels: overlay.LabelMap{"Background": overlay.Label{Color: "", ID: 0}}}, "", "  ")
		if err == nil {
			log.Println(string(bts))
		}
	}
}

func main() {
	start := time.Now()
	log.Println("pixelcounter start")
	defer func() {
		log.Printf("pixelcounter end. Took %.2f seconds\n", time.Since(start).Seconds())
	}()

	var threshold int
	var overlayPath, jsonConfig, manifest, suffix, momentLabels string

	flag.StringVar(&overlayPath, "overlay", "", "Path to folder with encoded overlay images or .tar.gz files with overlay images (or both)")
	flag.StringVar(&jsonConfig, "config", "", "JSONConfig file from the github.com/carbocation/genomisc/overlay package")
	flag.StringVar(&manifest, "manifest", "", "(Optional) Path to manifest. If provided, will only look at files in the manifest rather than listing the entire directory's contents.")
	flag.StringVar(&suffix, "suffix", ".png.mask.png", "(Optional) Suffix after .dcm. Only used if using the -manifest option.")
	flag.IntVar(&threshold, "threshold", 5, "(Optional) Number of pixels below which to ignore a connected component for the thresholded subcount.")
	flag.StringVar(&momentLabels, "moment-labels", "", "(Optional) Comma-delimited list of LabelIDs for which you want to compute image moment values and bounding boxes.")
	flag.Parse()

	if overlayPath == "" || jsonConfig == "" {
		flag.Usage()
		os.Exit(1)
	}

	config, err := overlay.ParseJSONConfigFromPath(jsonConfig)
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	labelsNeedMoments, err := parseMomentLabelIDs(momentLabels)
	if err != nil {
		log.Fatalln(err)
	}
	if labelsNeedMoments != nil {
		fmt.Fprintln(os.Stderr, "Will compute moment information for the following fields:")
		for name, label := range config.Labels {
			if _, exists := labelsNeedMoments[uint8(label.ID)]; exists {
				fmt.Fprintf(os.Stderr, "%s (#%d)\n", name, label.ID)
			}
		}
	}

	if manifest != "" {

		if err := runSlice(config, overlayPath, suffix, manifest, threshold, labelsNeedMoments); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if err := runFolder(config, overlayPath, threshold, labelsNeedMoments); err != nil {
		log.Fatalln(err)
	}

}

func runSlice(config overlay.JSONConfig, overlayPath, suffix, manifest string, threshold int, labelsNeedMoments map[uint8]struct{}) error {

	dicoms, err := getDicomSlice(manifest)
	if err != nil {
		return err
	}

	printHeader(config, threshold, labelsNeedMoments)

	concurrency := 4 * runtime.NumCPU()
	sem := make(chan bool, concurrency)

	// Process every image in the manifest
	for i, file := range dicoms {
		sem <- true
		go func(file string) {
			if err := processOneImageFilepath(overlayPath+"/"+file+suffix, file, config, threshold, labelsNeedMoments); err != nil {
				log.Println(err)
			}
			<-sem
		}(file)

		if (i+1)%1000 == 0 {
			log.Printf("Processed %d images\n", i+1)
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func runFolder(config overlay.JSONConfig, overlayPath string, threshold int, labelsNeedMoments map[uint8]struct{}) error {

	files, err := os.ReadDir(overlayPath)
	if err != nil {
		return err
	}

	printHeader(config, threshold, labelsNeedMoments)

	concurrency := 4 * runtime.NumCPU()
	sem := make(chan bool, concurrency)

	// Process every image in the folder
	for i, file := range files {
		if file.IsDir() {
			continue
		}

		sem <- true
		go func(file string) {
			var err error
			if strings.HasSuffix(file, ".tar.gz") {
				if err = processOneTarGZFilepath(overlayPath+"/"+file, file, config, threshold, labelsNeedMoments); err != nil {
					log.Println(err)
				}
			} else if strings.HasSuffix(file, ".png") ||
				strings.HasSuffix(file, ".gif") ||
				strings.HasSuffix(file, ".bmp") ||
				strings.HasSuffix(file, ".jpeg") ||
				strings.HasSuffix(file, ".jpg") {
				if err = processOneImageFilepath(overlayPath+"/"+file, file, config, threshold, labelsNeedMoments); err != nil {
					log.Printf("%s: %s\n", file, err)
				}
			}
			<-sem
		}(file.Name())

		if (i+1)%1000 == 0 {
			log.Printf("Processed %d images\n", i+1)
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func printHeader(config overlay.JSONConfig, threshold int, labelsNeedMoments map[uint8]struct{}) {
	header := []string{"dicom", "width", "height", "pixels"}

	for _, v := range config.Labels.Sorted() {
		formatted := fmt.Sprintf("ID%d_%s", v.ID, strings.ReplaceAll(v.Label, " ", "_"))

		header = append(header, formatted)
		header = append(header, fmt.Sprintf("%s_%d_thresholded", formatted, threshold))
		header = append(header, fmt.Sprintf("%s_components", formatted))
		header = append(header, fmt.Sprintf("%s_%d_thresholded_components", formatted, threshold))

		if _, needsMoment := labelsNeedMoments[uint8(v.ID)]; needsMoment {
			header = append(header, fmt.Sprintf("%s_LongAxisAngle", formatted))   // LongAxisAngle
			header = append(header, fmt.Sprintf("%s_LongAxisPixels", formatted))  // LongAxisPixels
			header = append(header, fmt.Sprintf("%s_ShortAxisPixels", formatted)) // ShortAxisPixels
			header = append(header, fmt.Sprintf("%s_Eccentricity", formatted))    // Eccentricity
			header = append(header, fmt.Sprintf("%s_CentroidX", formatted))
			header = append(header, fmt.Sprintf("%s_CentroidY", formatted))
			header = append(header, fmt.Sprintf("%s_TopLeftX", formatted))
			header = append(header, fmt.Sprintf("%s_TopLeftY", formatted))
			header = append(header, fmt.Sprintf("%s_BottomRightX", formatted))
			header = append(header, fmt.Sprintf("%s_BottomRightY", formatted))
		}
	}

	header = append(header, "total_connected_components")
	header = append(header, fmt.Sprintf("total_%d_thresholded_connected_components", threshold))

	header = append(header, fmt.Sprintf("total_%d_thresholded_pixels", threshold))

	fmt.Println(strings.Join(header, "\t"))
}

func processOneTarGZFilepath(filePath, filename string, config overlay.JSONConfig, threshold int, labelsNeedMoments map[uint8]struct{}) error {
	// Open and stream/ungzip the tar.gz
	f, _, err := bulkprocess.MaybeOpenFromGoogleStorage(filePath, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)

	// Iterate over tarfile contents, processing all non-directory files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		} else if header.Typeflag == tar.TypeDir {
			continue
		}

		imageBytes, err := ioutil.ReadAll(tarReader)
		if err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}

		br := bytes.NewReader(imageBytes)
		newImg, err := bulkprocess.DecodeImageFromReader(br)
		if err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}

		if err := processOneImage(newImg, header.Name, config, threshold, labelsNeedMoments); err != nil {
			log.Printf("%s: %v\n", header.Name, err)
		}
	}

	return nil
}

func processOneImageFilepath(filePath, filename string, config overlay.JSONConfig, threshold int, labelsNeedMoments map[uint8]struct{}) error {
	rawOverlayImg, err := overlay.OpenImageFromLocalFile(filePath)
	if err != nil {
		return err
	}

	return processOneImage(rawOverlayImg, filename, config, threshold, labelsNeedMoments)
}

func processOneImage(rawOverlayImg image.Image, filename string, config overlay.JSONConfig, threshold int, labelsNeedMoments map[uint8]struct{}) error {
	// Heuristic: get dicom name
	dicom := strings.ReplaceAll(filename, ".png.mask.png", "")
	dicom = strings.ReplaceAll(dicom, ".mask.png", "")
	entry := []string{dicom}

	// Add pixel count info
	entry = append(entry, strconv.Itoa(rawOverlayImg.Bounds().Dx()))
	entry = append(entry, strconv.Itoa(rawOverlayImg.Bounds().Dy()))
	entry = append(entry, strconv.Itoa(rawOverlayImg.Bounds().Dx()*rawOverlayImg.Bounds().Dy()))

	// Count connected components
	connected, err := overlay.NewConnected(rawOverlayImg)
	if err != nil {
		return err
	}

	pixelCountMap, connectedCounts, thresholdedPixelCountMap, thresholdedConnectedCounts, err := connected.Count(config.Labels, threshold)
	if err != nil {
		return err
	}

	totalThresholdedPixels := 0
	for _, v := range config.Labels.Sorted() {
		_, needsMoment := labelsNeedMoments[uint8(v.ID)]

		if _, exists := pixelCountMap[v]; !exists {
			entry = append(entry, "0") // Pixels
			entry = append(entry, "0") // Thresholded pixels
			entry = append(entry, "0") // Connected components
			entry = append(entry, "0") // Thresholded connected components

			if needsMoment {
				entry = append(entry, "0") // LongAxisAngle
				entry = append(entry, "0") // LongAxisPixels
				entry = append(entry, "0") // ShortAxisPixels
				entry = append(entry, "0") // Eccentricity
				entry = append(entry, "0") // Centroid.X
				entry = append(entry, "0") // Centroid.Y
				entry = append(entry, "0") // TopLeft.X
				entry = append(entry, "0") // TopLeft.Y
				entry = append(entry, "0") // BottomRight.X
				entry = append(entry, "0") // BottomRight.Y
			}

			continue
		}

		entry = append(entry, strconv.Itoa(pixelCountMap[v]))              // Pixels
		entry = append(entry, strconv.Itoa(thresholdedPixelCountMap[v]))   // Thresholded pixels
		entry = append(entry, strconv.Itoa(connectedCounts[v]))            // Connected components
		entry = append(entry, strconv.Itoa(thresholdedConnectedCounts[v])) // Thresholded connected components

		if needsMoment {
			momentComponents := connected.LabeledConnectedComponents[uint8(v.ID)]

			var mergedLabeled overlay.ConnectedComponent
			initialized := false
			for _, v := range momentComponents {
				if v.PixelCount > threshold {
					if !initialized {
						mergedLabeled = v
						initialized = true
					} else {
						mergedLabeled = overlay.MergeConnectedComponentsSameLabel(mergedLabeled, v)
					}
				}
			}
			// log.Printf("%+v\n", aorticLabeled)
			moments, err := connected.ComputeMoments(mergedLabeled, overlay.MomentMethodLabel)
			if err != nil {
				return err
			}
			// log.Printf("%+v\n", x)
			entry = append(entry, strconv.FormatFloat(moments.LongAxisOrientationRadians, 'g', 4, 64)) // LongAxisAngle
			entry = append(entry, strconv.FormatFloat(moments.LongAxisPixels, 'g', 4, 64))             // LongAxisPixels
			entry = append(entry, strconv.FormatFloat(moments.ShortAxisPixels, 'g', 4, 64))            // ShortAxisPixels
			entry = append(entry, strconv.FormatFloat(moments.Eccentricity, 'g', 4, 64))               // Eccentricity
			entry = append(entry, strconv.Itoa(int(math.Floor(moments.Centroid.X))))
			entry = append(entry, strconv.Itoa(int(math.Floor(moments.Centroid.Y))))
			entry = append(entry, strconv.Itoa(moments.Bounds.TopLeft.X))
			entry = append(entry, strconv.Itoa(moments.Bounds.TopLeft.Y))
			entry = append(entry, strconv.Itoa(moments.Bounds.BottomRight.X))
			entry = append(entry, strconv.Itoa(moments.Bounds.BottomRight.Y))
		}

		// Since thresholding will change the total number of pixels, need to
		// track it
		totalThresholdedPixels += thresholdedPixelCountMap[v]
	}

	totalComponents := 0
	for _, v := range connectedCounts {
		totalComponents += v
	}
	entry = append(entry, strconv.Itoa(totalComponents))

	totalThresholdedComponents := 0
	for _, v := range thresholdedConnectedCounts {
		totalThresholdedComponents += v
	}
	entry = append(entry, strconv.Itoa(totalThresholdedComponents))
	entry = append(entry, strconv.Itoa(totalThresholdedPixels))

	fmt.Println(strings.Join(entry, "\t"))

	return nil
}
