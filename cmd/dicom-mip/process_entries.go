package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"sort"
	"time"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

type seriesMap struct {
	Zip    string
	Series string
}

func processEntries(entries []manifestEntry, key manifestKey, folder string, doNotSort, beVerbose bool) error {
	started := time.Now()
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	if !doNotSort {
		// Sort by series, then by ImagePatientPosition:Z
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].series != entries[j].series {
				return entries[i].series < entries[j].series
			}

			return entries[i].ImagePositionPatientZ < entries[j].ImagePositionPatientZ
		})
	}

	// Make zipfile and series aware:
	zipSeriesMap := make(map[seriesMap][]manifestEntry)
	for _, entry := range entries {
		zipKey := seriesMap{Zip: entry.zip, Series: entry.series}
		pngData := zipSeriesMap[zipKey]
		pngData = append(pngData, entry)
		zipSeriesMap[zipKey] = pngData
	}

	// Fetch images from each zipfile once:
	zipMap := make(map[string]map[string]image.Image)
	for entry := range zipSeriesMap {
		if _, exists := zipMap[entry.Zip]; !exists {
			if beVerbose {
				fmt.Printf("Fetching images for %+v:%v\n", key, entry.Zip)
			}

			// Fetch the zip file just once per zip, even if it has many series
			imgMap, err := bulkprocess.FetchImagesFromZIP(folder+"/"+entry.Zip, false, client)
			if err != nil {
				return err
			}

			zipMap[entry.Zip] = imgMap
		}
	}

	if beVerbose {
		fmt.Printf("Found %d zip files+series combinations. Summary images and movies will be created for each.\n", len(zipSeriesMap))
	}

	// For now, doing one zip at a time
	errchan := make(chan error)

	for zip, pngData := range zipSeriesMap {
		imgMap := zipMap[zip.Zip]

		go func(zip seriesMap, imgMap map[string]image.Image, pngData []manifestEntry) {
			outName := zip.Zip + "_" + zip.Series + ".coronal.png"
			im, err := canvasMakeOneCoronalMIPFromImageMapNonsquare(pngData, imgMap, AverageIntensity, 0)
			if err != nil {
				errchan <- err
				return
			}
			errchan <- savePNG(im, outName)

		}(zip, imgMap, pngData)

		go func(zip seriesMap, imgMap map[string]image.Image, pngData []manifestEntry) {
			outName := zip.Zip + "_" + zip.Series + ".sagittal.png"
			im, err := canvasMakeOneSagittalMIPFromImageMapNonsquare(pngData, imgMap, AverageIntensity, 0)
			if err != nil {
				errchan <- err
				return
			}
			errchan <- savePNG(im, outName)
		}(zip, imgMap, pngData)

		// go func() {
		// 	errchan <- nil
		// 	errchan <- nil
		// }()
		// continue

		// First create a sagittal MIP, then use its width in pixels to
		// determine the number of frames for the GIF of the coronal MIP:
		go func(zip seriesMap, imgMap map[string]image.Image, pngData []manifestEntry) {
			outName := zip.Zip + "_" + zip.Series + ".coronal.mp4"
			// outName := zip.Zip + "_" + zip.Series + ".coronal.gif"
			_, canvasDepth, _, _, _, _ := findCanvasAndOffsets(pngData, imgMap)

			log.Println("\nCreating a", int(math.Ceil(canvasDepth)), "frame GIF for", outName)
			defer func(name string) {
				log.Println("\nDone creating", name)
			}(outName)
			imgList := make([]image.Image, 0, int(math.Ceil(canvasDepth)))
			for i := 0; i < int(math.Ceil(canvasDepth)); i++ {
				im2, err := canvasMakeOneCoronalMIPFromImageMapNonsquare(pngData, imgMap, SliceIntensity, i)
				if err != nil {
					errchan <- err
					break
				}

				allBlack := true
			OuterLoop:
				for j := 0; j < im2.Bounds().Max.Y; j++ {
					for k := 0; k < im2.Bounds().Max.X; k++ {
						if col := im2.At(k, j); col != (color.RGBA{0, 0, 0, 255}) {
							allBlack = false
							break OuterLoop
						}
					}
				}

				if !allBlack {
					imgList = append(imgList, im2)
				}
			}

			errchan <- makeOneMPEG(imgList, outName, 20)
			// errchan <- makeOneGIF(imgList, outName, 10, false)

		}(zip, imgMap, pngData)

		// First create a coronal MIP, then use its width in pixels to
		// determine the number of frames for the GIF of the sagittal MIP:
		go func(zip seriesMap, imgMap map[string]image.Image, pngData []manifestEntry) {
			outName := zip.Zip + "_" + zip.Series + ".sagittal.mp4"
			// outName := zip.Zip + "_" + zip.Series + ".sagittal.gif"
			im, err := canvasMakeOneCoronalMIPFromImageMapNonsquare(pngData, imgMap, AverageIntensity, 0)
			if err != nil {
				errchan <- err
				return
			}

			imgList := make([]image.Image, 0, im.Bounds().Max.X)
			for i := 0; i < im.Bounds().Max.X; i++ {
				im2, err := canvasMakeOneSagittalMIPFromImageMapNonsquare(pngData, imgMap, SliceIntensity, i)
				if err != nil {
					errchan <- err
					break
				}

				allBlack := true
			OuterLoop:
				for j := 0; j < im2.Bounds().Max.Y; j++ {
					for k := 0; k < im2.Bounds().Max.X; k++ {
						if col := im2.At(k, j); col != (color.RGBA{0, 0, 0, 255}) {
							allBlack = false
							break OuterLoop
						}
					}
				}

				if !allBlack {
					imgList = append(imgList, im2)
				}
			}

			errchan <- makeOneMPEG(imgList, outName, 10)
			// errchan <- makeOneGIF(imgList, outName, 20, false)

		}(zip, imgMap, pngData)
	}

	completed := 0

	var err error
	imPerZip := 4
WaitLoop:
	for {
		select {
		case err = <-errchan:
			completed++
			if beVerbose && err != nil {
				fmt.Println("Error making gif:", err.Error())
			}

			// We produce N gifs per zipfile+series combination
			if completed >= (imPerZip * len(zipSeriesMap)) {
				break WaitLoop
			}
		case current := <-tick.C:
			if beVerbose {
				fmt.Printf("\rCreating %d more images for %+v (%s)", imPerZip*len(zipSeriesMap)-completed, key, current.Sub(started))
			}
		}
	}

	if beVerbose && err == nil {
		fmt.Printf("\nSuccessfully created file #%d for %s_%s\n", completed, key.SampleID, key.Instance)
	}

	return nil
}
