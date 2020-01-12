package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
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
	var overlayPath, jsonConfig string

	flag.StringVar(&overlayPath, "overlay", "", "Path to encoded overlay image")
	flag.StringVar(&jsonConfig, "config", "", "JSONConfig file from the github.com/carbocation/genomisc/overlay package")
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

	if err := run(config, overlayPath); err != nil {
		log.Fatalln(err)
	}

}

func run(config overlay.JSONConfig, overlayPath string) error {

	files, err := scanFolder(overlayPath)
	if err != nil {
		return err
	}

	printHeader(config)

	// Process every image in the folder
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if err := processOneImage(overlayPath+"/"+file.Name(), file.Name(), config); err != nil {
			return err
		}
	}

	return nil
}

func printHeader(config overlay.JSONConfig) {
	header := []string{"dicom", "width", "height", "pixels"}
	for _, v := range config.Labels.Sorted() {
		formatted := fmt.Sprintf("ID%d_%s", v.ID, strings.ReplaceAll(v.Label, " ", "_"))

		header = append(header, formatted)
		header = append(header, fmt.Sprintf("%s_components", formatted))
	}

	fmt.Println(strings.Join(header, "\t"))
}

func processOneImage(filePath, filename string, config overlay.JSONConfig) error {
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

	countMap, err := config.Labels.CountEncodedPixels(rawOverlayImg)
	if err != nil {
		return err
	}

	// Count connected components
	connected, err := overlay.NewConnected(rawOverlayImg)
	if err != nil {
		return err
	}

	connectedCounts, err := connected.Count(config.Labels)
	if err != nil {
		return err
	}

	for _, v := range config.Labels.Sorted() {

		if _, exists := countMap[v]; !exists {
			entry = append(entry, "0")
			continue
		}

		entry = append(entry, strconv.Itoa(countMap[v]))
		entry = append(entry, strconv.Itoa(connectedCounts[v]))
	}

	fmt.Println(strings.Join(entry, "\t"))

	return nil
}
