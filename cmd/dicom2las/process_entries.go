package main

import (
	"fmt"
	"image"
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
			outName := zip.Zip + "_" + zip.Series + ".las"
			errchan <- makeLAS(pngData, imgMap, outName)

		}(zip, imgMap, pngData)

	}

	completed := 0

	var err error
	imPerZip := 1
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
