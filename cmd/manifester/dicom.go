package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
)

func ManifestForDicom(path string) error {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return pfx.Err(err)
	}

	fmt.Fprintf(STDOUT, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		"sample_id",
		"field_id",
		"instance",
		"index",
		"zip_file",
		"dicom_file",
		"series",
		"date",
		"instance_number",
		"overlay_text",
		"overlay_fraction",
		"overlay_rows",
		"overlay_cols",
		"rows",
		"cols",
		"image_x",
		"image_y",
		"image_z",
		"px_height_mm",
		"px_width_mm",
		"slice_thickness_mm",
		"series_number",
		"acquisition_number",
		"device_serial_number",
		"station_name",
		"software_versions",
		"echo_time",
		"nominal_interval",
		"slice_location",
		"trigger_time",
	)

	concurrency := 4 * runtime.NumCPU()

	results := make(chan string, concurrency)
	doneListening := make(chan struct{})
	go func() {
		defer func() { doneListening <- struct{}{} }()
		// Serialize results so you don't dump text haphazardly into os.Stdout
		// (which is not goroutine safe).
		for {
			select {
			case res, ok := <-results:
				if !ok {
					return
				}

				fmt.Fprintln(STDOUT, res)
			}
		}

	}()

	semaphore := make(chan struct{}, concurrency)

	for _, file := range files {

		// Will block after `concurrency` simultaneous goroutines are running
		semaphore <- struct{}{}

		go func(file os.FileInfo) {

			// Be sure to permit unblocking once we finish
			defer func() { <-semaphore }()

			if !strings.HasSuffix(file.Name(), ".zip") {
				return
			}

			err := bulkprocess.CardiacMRIZipIterator(path+file.Name(), func(dcm bulkprocess.DicomOutput) error {
				if err := PrintCSVRow(dcm, results); err != nil {
					log.Printf("Error parsing %+v\n", dcm)
					return err
				}

				return nil
			})
			if err != nil {
				log.Println("Error parsing", path+file.Name())
				log.Println(err)
			}
		}(file)
	}

	// Make sure we finish all the reads before we exit, otherwise we'll lose
	// the last `concurrency` lines.
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	// Close the results channel and make sure we are done listening
	close(results)
	<-doneListening

	return nil
}
