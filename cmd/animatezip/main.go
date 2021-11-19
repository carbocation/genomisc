package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

type seriesMap struct {
	Zip    string
	Series string
}

const (
	SampleIDColumnName = "sample_id"
	InstanceColumnName = "instance"
)

var (
	SeriesColumnName    = "series"
	TimepointColumnName = "trigger_time"
	ZipColumnName       = "zip_file"
	DicomColumnName     = "dicom_file"
)

// Safe for concurrent use by multiple goroutines so we'll make this a global
var client *storage.Client

func main() {
	var includeOverlay, doNotSort, labelDicom, batch bool
	var manifest, folder string
	var delay int
	flag.StringVar(&manifest, "manifest", "", "Path to manifest file")
	flag.StringVar(&folder, "folder", "", "Path to google storage folder that contains zip files.")
	flag.StringVar(&SeriesColumnName, "series_column_name", "series", "Name of the column in the manifest with the series.")
	flag.StringVar(&DicomColumnName, "dicom_column_name", "dicom_file", "Name of the column in the manifest with the dicoms.")
	flag.StringVar(&ZipColumnName, "zip_column_name", "zip_file", "Name of the column in the manifest with the zip file.")
	flag.StringVar(&TimepointColumnName, "sequence_column_name", "trigger_time", "Name of the column that indicates the order of the images with an increasing number.")
	flag.IntVar(&delay, "delay", 2, "Milliseconds between each frame of the gif.")
	flag.BoolVar(&includeOverlay, "overlay", false, "Print overlay information, if present?")
	flag.BoolVar(&doNotSort, "donotsort", false, "Pass this if you do not want to sort the manifest (i.e., you've already sorted it)")
	flag.BoolVar(&labelDicom, "labeldicom", false, "Pass this if you want to print the dicom name at the top of each frame of the animated gif.")
	flag.BoolVar(&batch, "batch", false, "Pass this if you want to run in batch mode instead of interactive mode; if so, all gifs will be created and then the program will exit.")
	flag.Parse()

	if manifest == "" || folder == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	folder = strings.TrimSuffix(folder, "/")

	// Initialize the Google Storage client, but only if our folder indicates
	// that we are pointing to a Google Storage path.
	if strings.HasPrefix(folder, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Println("Animated Gif maker")

	if batch {
		if err := runBatch(manifest, folder, delay, includeOverlay, doNotSort, labelDicom); err != nil {
			log.Fatalln(err)
		}
		return
	}

	if err := run(manifest, folder, delay, includeOverlay, doNotSort, labelDicom); err != nil {
		log.Fatalln(err)
	}

	log.Println("Quitting")
}

func runBatch(manifest, folder string, delay int, includeOverlay, doNotSort, labelDicom bool) error {
	man, err := parseManifest(manifest, doNotSort)
	if err != nil {
		return err
	}

	fmt.Println("We are aware of", len(man), "samples in the manifest")

	for key := range man {
		entries, exists := man[key]
		if !exists {
			fmt.Println(key, "not found in the manifest")
			continue
		}

		if err = processEntries(entries, key, folder, delay, labelDicom, includeOverlay, doNotSort, false); err != nil {
			log.Println(err)
		}
	}

	return nil
}

func run(manifest, folder string, delay int, includeOverlay, doNotSort, labelDicom bool) error {

	man, err := parseManifest(manifest, doNotSort)
	if err != nil {
		return err
	}

	rdr := bufio.NewReader(os.Stdin)
	fmt.Println("We are aware of", len(man), "samples in the manifest")
	fmt.Println("An example of the sampleid_instance format is: 1234567_2")
	fmt.Println("Enter 'rand' for a random entry")
	fmt.Println("Enter 'q' to quit")
	fmt.Println("---------------------")

	for {
		fmt.Print("[sampleid_instance]> ")
		text, err := rdr.ReadString('\n')
		if err != nil {
			return err
		}
		text = strings.ReplaceAll(text, "\n", "")

		if text == "q" {
			fmt.Println("quitting")
			break
		}

		var key manifestKey

		if text == "rand" {

			for k := range man {
				key = k
				break
			}
		} else {

			input := strings.Split(text, "_")
			if len(input) != 2 {
				fmt.Println("Expected sampleid separated from instance by an underscore (_)")
				continue
			}

			key = manifestKey{SampleID: input[0], Instance: input[1]}
		}

		entries, exists := man[key]
		if !exists {
			fmt.Println(key, "not found in the manifest")
			continue
		}

		if err = processEntries(entries, key, folder, delay, labelDicom, includeOverlay, doNotSort, true); err != nil {
			log.Println(err)
		}
	}

	return nil
}

// processEntries handles fetching the (possibly remote) zip file, ingesting the
// desired images from the DICOM, computing the palatte, and emitting the .gif
// file(s).
func processEntries(entries []manifestEntry, key manifestKey, folder string, delay int, labelDicom, includeOverlay, doNotSort, beVerbose bool) error {
	started := time.Now()
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	if !doNotSort {
		// Sort by series, then by instance
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].series != entries[j].series {
				return entries[i].series < entries[j].series
			}

			return entries[i].timepoint < entries[j].timepoint
		})
	}

	// Make zipfile and series aware:
	zipSeriesMap := make(map[seriesMap][]string)
	for _, entry := range entries {
		zipKey := seriesMap{Zip: entry.zip, Series: entry.series}
		pngs := zipSeriesMap[zipKey]
		pngs = append(pngs, entry.dicom)
		zipSeriesMap[zipKey] = pngs
	}

	// Fetch images from each zipfile once:
	zipMap := make(map[string]map[string]image.Image)
	for entry := range zipSeriesMap {
		if _, exists := zipMap[entry.Zip]; !exists {
			if beVerbose {
				fmt.Printf("Fetching images for %+v:%v\n", key, entry.Zip)
			}

			// Fetch the zip file just once per zip, even if it has many series
			imgMap, err := bulkprocess.FetchImagesFromZIP(folder+"/"+entry.Zip, includeOverlay, client)
			if err != nil {
				return err
			}

			zipMap[entry.Zip] = imgMap
		}
	}

	if beVerbose {
		fmt.Printf("Found %d zip files+series combinations. One gif will be created for each.\n", len(zipSeriesMap))
	}

	// For now, doing one zip at a time
	errchan := make(chan error)

	for zip, pngs := range zipSeriesMap {
		imgMap := zipMap[zip.Zip]

		// Write the dicomName on top of each image?
		if labelDicom {
			for _, dicomName := range pngs {
				thisImg := imgMap[dicomName]

				label := dicomName

				drawableImg := addLabelGG(thisImg, label)
				imgMap[dicomName] = drawableImg
			}
		}

		go func(zip seriesMap, imgMap map[string]image.Image, pngs []string) {
			outName := zip.Zip + "_" + zip.Series + ".gif"

			errchan <- makeOneGifFromImageMap(pngs, imgMap, outName, delay)
		}(zip, imgMap, pngs)
	}

	var err error
	completed := 0

WaitLoop:
	for {
		select {
		case err = <-errchan:
			completed++
			if err != nil {
				fmt.Println("Error making gif:", err.Error())
			}

			if completed >= len(zipSeriesMap) {
				break WaitLoop
			}
		case current := <-tick.C:
			if beVerbose {
				fmt.Printf("\rCreating %d more gifs for %+v (%s)", len(zipSeriesMap)-completed, key, current.Sub(started))
			}
		}
	}

	if err == nil && beVerbose {
		fmt.Printf("\nSuccessfully created file #%d for %s_%s\n", completed, key.SampleID, key.Instance)
	}

	return err
}
