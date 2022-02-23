// dicom-mip produces maximum intensity projections, as well as average
// intensity projections and .mp4 videos of the projections through the 3D block
// of voxels from sagittal and coronal projections. Note that this program's
// video production requires that ffmpeg be installed. (See
// https://github.com/unixpickle/ffmpego#installation)
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	_ "github.com/carbocation/genomisc/compileinfoprint"
)

const (
	SampleIDColumnName = "sample_id"
	InstanceColumnName = "instance"
)

var (
	SeriesColumnName            = "series"
	SeriesNumberColumName       = "series_number"
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
	// go tool pprof --http=localhost:6060 ~/go/bin/dicom-mip PPROF_OUTPUT_FILE
	// defer profile.Start(profile.CPUProfile).Stop()

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
	flag.StringVar(&SeriesNumberColumName, "series_number_column_name", "series_number", "(Optional) Name of the column that indicates the series number of the images, useful for colorizing different acquisitions.")
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
