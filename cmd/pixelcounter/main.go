package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/carbocation/genomisc/overlay"
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
	var threshold int
	var overlayPath, jsonConfig, manifest, suffix string

	flag.StringVar(&overlayPath, "overlay", "", "Path to folder with encoded overlay images")
	flag.StringVar(&jsonConfig, "config", "", "JSONConfig file from the github.com/carbocation/genomisc/overlay package")
	flag.StringVar(&manifest, "manifest", "", "(Optional) Path to manifest. If provided, will only look at files in the manifest rather than listing the entire directory's contents.")
	flag.StringVar(&suffix, "suffix", ".png.mask.png", "(Optional) Suffix after .dcm. Only used if using the -manifest option.")
	flag.IntVar(&threshold, "threshold", 5, "(Optional) Number of pixels below which to ignore a connected component for the thresholded subcount.")
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

	if manifest != "" {

		if err := runSlice(config, overlayPath, suffix, manifest, threshold); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if err := runFolder(config, overlayPath, threshold); err != nil {
		log.Fatalln(err)
	}

}

func runSlice(config overlay.JSONConfig, overlayPath, suffix, manifest string, threshold int) error {

	dicoms, err := getDicomSlice(manifest)
	if err != nil {
		return err
	}

	printHeader(config, threshold)

	concurrency := 4 * runtime.NumCPU()
	sem := make(chan bool, concurrency)

	// Process every image in the manifest
	for i, file := range dicoms {
		sem <- true
		go func(file string) {
			if err := processOneImage(overlayPath+"/"+file+suffix, file, config, threshold); err != nil {
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

func runFolder(config overlay.JSONConfig, overlayPath string, threshold int) error {

	files, err := scanFolder(overlayPath)
	if err != nil {
		return err
	}

	printHeader(config, threshold)

	concurrency := 4 * runtime.NumCPU()
	sem := make(chan bool, concurrency)

	// Process every image in the folder
	for i, file := range files {
		if file.IsDir() {
			continue
		}

		sem <- true
		go func(file string) {
			if err := processOneImage(overlayPath+"/"+file, file, config, threshold); err != nil {
				log.Println(err)
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

func printHeader(config overlay.JSONConfig, threshold int) {
	header := []string{"dicom", "width", "height", "pixels"}
	for _, v := range config.Labels.Sorted() {
		formatted := fmt.Sprintf("ID%d_%s", v.ID, strings.ReplaceAll(v.Label, " ", "_"))

		header = append(header, formatted)
		header = append(header, fmt.Sprintf("%s_%d_thresholded", formatted, threshold))
		header = append(header, fmt.Sprintf("%s_components", formatted))
		header = append(header, fmt.Sprintf("%s_%d_thresholded_components", formatted, threshold))
	}

	header = append(header, "total_connected_components")
	header = append(header, fmt.Sprintf("total_%d_thresholded_connected_components", threshold))

	header = append(header, fmt.Sprintf("total_%d_thresholded_pixels", threshold))

	fmt.Println(strings.Join(header, "\t"))
}

func processOneImage(filePath, filename string, config overlay.JSONConfig, threshold int) error {
	// Heuristic: get dicom name
	dicom := strings.ReplaceAll(filename, ".png.mask.png", "")
	dicom = strings.ReplaceAll(dicom, ".mask.png", "")

	entry := []string{dicom}

	rawOverlayImg, err := overlay.OpenImageFromLocalFile(filePath)
	if err != nil {
		return err
	}

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

		if _, exists := pixelCountMap[v]; !exists {
			entry = append(entry, "0") // Pixels
			entry = append(entry, "0") // Thresholded pixels
			entry = append(entry, "0") // Connected components
			entry = append(entry, "0") // Thresholded connected components
			continue
		}

		entry = append(entry, strconv.Itoa(pixelCountMap[v]))              // Pixels
		entry = append(entry, strconv.Itoa(thresholdedPixelCountMap[v]))   // Thresholded pixels
		entry = append(entry, strconv.Itoa(connectedCounts[v]))            // Connected components
		entry = append(entry, strconv.Itoa(thresholdedConnectedCounts[v])) // Thresholded connected components

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
