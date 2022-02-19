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
	_ "github.com/carbocation/genomisc/compileinfoprint"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
)

const (
	SampleIDColumnName = "sample_id"
	InstanceColumnName = "instance"
)

var (
	SeriesColumnName            = "series"
	ZipColumnName               = "zip_file"
	DicomColumnName             = "dicom_file"
	PixelWidthNativeXColumn     = "px_width_mm"
	PixelWidthNativeYColumn     = "px_height_mm"
	PixelWidthNativeZColumn     = "slice_thickness_mm"
	ImagePositionPatientXColumn = "image_x"
	ImagePositionPatientYColumn = "image_y"
	ImagePositionPatientZColumn = "image_z"
)

// Safe for concurrent use by multiple goroutines so we'll make this a global
var client *storage.Client

func main() {

	var doNotSort, batch bool
	var manifest, folder string
	flag.StringVar(&manifest, "manifest", "", "Path to manifest file")
	flag.StringVar(&folder, "folder", "", "Path to google storage folder that contains zip files.")
	flag.StringVar(&DicomColumnName, "dicom_column_name", "dicom_file", "Name of the column in the manifest with the dicoms.")
	flag.StringVar(&ZipColumnName, "zip_column_name", "zip_file", "Name of the column in the manifest with the zip file.")
	flag.StringVar(&PixelWidthNativeXColumn, "pixel_width_x", "px_width_mm", "Name of the column that indicates the width of the pixels in the original images.")
	flag.StringVar(&PixelWidthNativeYColumn, "pixel_width_y", "px_height_mm", "Name of the column that indicates the height of the pixels in the original images.")
	flag.StringVar(&PixelWidthNativeZColumn, "pixel_width_z", "slice_thickness_mm", "Name of the column that indicates the depth/thickness of the pixels in the original images.")
	flag.StringVar(&ImagePositionPatientXColumn, "image_x", "image_x", "Name of the column in the manifest with the X position of the top left pixel of the images.")
	flag.StringVar(&ImagePositionPatientYColumn, "image_y", "image_y", "Name of the column in the manifest with the Y position of the top left pixel of the images.")
	flag.StringVar(&ImagePositionPatientZColumn, "image_z", "image_z", "Name of the column in the manifest with the Z position of the top left pixel of the images.")
	flag.StringVar(&SeriesColumnName, "series_column_name", "sample_id", "Name of the column that indicates the series of the images.")
	flag.BoolVar(&doNotSort, "donotsort", false, "Pass this if you do not want to sort the manifest (i.e., you've already sorted it)")
	flag.BoolVar(&batch, "batch", false, "Pass this if you want to run in batch mode.")
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

	if batch {
		if err := runBatch(manifest, folder, doNotSort); err != nil {
			log.Fatalln(err)
		}

		return
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

func runBatch(manifest, folder string, doNotSort bool) error {
	fmt.Println("MIP maker")

	man, err := parseManifest(manifest)
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

		if err = processEntries(entries, key, folder, doNotSort, false); err != nil {
			log.Println(err)
		}
	}

	return nil
}

func run(manifest, folder string, doNotSort bool) error {

	fmt.Println("MIP maker")

	man, err := parseManifest(manifest)
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

		processEntries(entries, key, folder, doNotSort, true)
	}

	return nil
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
		fmt.Printf("Found %d zip files+series combinations. One gif will be created for each.\n", len(zipSeriesMap))
	}

	// For now, doing one zip at a time
	errchan := make(chan error)

	for zip, pngData := range zipSeriesMap {
		imgMap := zipMap[zip.Zip]

		go func(zip seriesMap, imgMap map[string]image.Image, pngData []manifestEntry) {
			outName := zip.Zip + "_" + zip.Series + ".coronal.png"
			errchan <- canvasMakeOneCoronalMIPFromImageMapNonsquare(pngData, imgMap, outName)
		}(zip, imgMap, pngData)

		go func(zip seriesMap, imgMap map[string]image.Image, pngData []manifestEntry) {
			outName := zip.Zip + "_" + zip.Series + ".sagittal.png"
			errchan <- canvasMakeOneSagittalMIPFromImageMapNonsquare(pngData, imgMap, outName)
		}(zip, imgMap, pngData)
	}

	completed := 0

	var err error
WaitLoop:
	for {
		select {
		case err = <-errchan:
			completed++
			if beVerbose && err != nil {
				fmt.Println("Error making gif:", err.Error())
			}

			// We produce 2 gifs per zipfile+series combination
			if completed >= (2 * len(zipSeriesMap)) {
				break WaitLoop
			}
		case current := <-tick.C:
			if beVerbose {
				fmt.Printf("\rCreating %d more gifs for %+v (%s)", len(zipSeriesMap)-completed, key, current.Sub(started))
			}
		}
	}

	if beVerbose && err == nil {
		fmt.Printf("\nSuccessfully created file #%d for %s_%s\n", completed, key.SampleID, key.Instance)
	}

	return nil
}
