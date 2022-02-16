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
	"github.com/pkg/profile"
)

const (
	SampleIDColumnName = "sample_id"
	InstanceColumnName = "instance"
)

var (
	SeriesColumnName    = "series"
	TimepointColumnName = "trigger_time"
	ZipColumnName       = "zip_file"
	DicomColumnName     = "dicom_file"
	NativeXColumnName   = "px_width_mm"
	NativeYColumnName   = "px_height_mm"
	NativeZColumnName   = "slice_thickness_mm"
)

// Safe for concurrent use by multiple goroutines so we'll make this a global
var client *storage.Client

func main() {
	defer profile.Start(profile.CPUProfile, profile.NoShutdownHook).Stop()

	var doNotSort bool
	var manifest, folder string
	flag.StringVar(&manifest, "manifest", "", "Path to manifest file")
	flag.StringVar(&folder, "folder", "", "Path to google storage folder that contains zip files.")
	flag.StringVar(&DicomColumnName, "dicom_column_name", "dicom_file", "Name of the column in the manifest with the dicoms.")
	flag.StringVar(&ZipColumnName, "zip_column_name", "zip_file", "Name of the column in the manifest with the zip file.")
	flag.StringVar(&TimepointColumnName, "sequence_column_name", "trigger_time", "Name of the column that indicates the order of the images with an increasing number.")
	flag.StringVar(&NativeXColumnName, "native_x_column_name", "px_width_mm", "Name of the column that indicates the width of the images.")
	flag.StringVar(&NativeYColumnName, "native_y_column_name", "px_height_mm", "Name of the column that indicates the height of the images.")
	flag.StringVar(&NativeZColumnName, "native_z_column_name", "slice_thickness_mm", "Name of the column that indicates the depth/thickness of the images.")
	flag.StringVar(&SeriesColumnName, "series_column_name", "series", "Name of the column that indicates the series of the images.")
	flag.BoolVar(&doNotSort, "donotsort", false, "Pass this if you do not want to sort the manifest (i.e., you've already sorted it)")
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

	if err := run(manifest, folder, doNotSort); err != nil {
		log.Fatalln(err)
	}

	log.Println("Quitting")
}

type seriesMap struct {
	Zip    string
	Series string
}

func run(manifest, folder string, doNotSort bool) error {

	fmt.Println("MIP maker")

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

	tick := time.NewTicker(1 * time.Second)

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
				fmt.Printf("Fetching images for %+v:%v\n", key, entry.Zip)

				// Fetch the zip file just once per zip, even if it has many series
				imgMap, err := bulkprocess.FetchImagesFromZIP(folder+"/"+entry.Zip, false, client)
				if err != nil {
					return err
				}

				zipMap[entry.Zip] = imgMap
			}
		}

		fmt.Printf("Found %d zip files+series combinations. One gif will be created for each.\n", len(zipSeriesMap))

		// For now, doing one zip at a time
		errchan := make(chan error)
		started := time.Now()

		for zip, pngData := range zipSeriesMap {
			imgMap := zipMap[zip.Zip]

			go func(zip seriesMap, imgMap map[string]image.Image, pngData []manifestEntry) {
				outName := zip.Zip + "_" + zip.Series + ".coronal.png"
				// errchan <- makeOneCoronalMIPFromImageMap(pngData, imgMap, outName)
				// errchan <- makeOneCoronalMIPFromImageMapNonsquare(pngData, imgMap, outName)
				errchan <- canvasMakeOneCoronalMIPFromImageMapNonsquare(pngData, imgMap, outName)

				outName = zip.Zip + "_" + zip.Series + ".sagittal.png"
				// errchan <- makeOneSagittalMIPFromImageMap(pngData, imgMap, outName)
				// errchan <- makeOneSagittalMIPFromImageMapNonsquare(pngData, imgMap, outName)
				// errchan <- vectorMakeOneSagittalMIPFromImageMapNonsquare(pngData, imgMap, outName)
				errchan <- canvasMakeOneSagittalMIPFromImageMapNonsquare(pngData, imgMap, outName)
			}(zip, imgMap, pngData)
		}

		completed := 0

	WaitLoop:
		for {
			select {
			case err = <-errchan:
				completed++
				if err != nil {
					fmt.Println("Error making gif:", err.Error())
				} else {
					log.Println("Completed a gif")
				}

				// We produce 2 gifs per zipfile+series combination
				if completed >= (2 * len(zipSeriesMap)) {
					break WaitLoop
				}
			case current := <-tick.C:
				fmt.Printf("\rCreating %d more gifs for %+v (%s)", len(zipSeriesMap)-completed, key, current.Sub(started))
			}
		}

		if err == nil {
			fmt.Printf("\nSuccessfully created file #%d for %s_%s\n", completed, key.SampleID, key.Instance)
		}
	}

	return nil
}
