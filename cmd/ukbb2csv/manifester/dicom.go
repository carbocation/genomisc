package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
)

func ManifestForDicom(path, fileList string) error {

	var files []string

	if fileList != "" {
		// File list - process just the requested files. Will prefix with path +
		// "/" if path is not the empty string ("").

		fl, err := os.Open(fileList)
		if err != nil {
			return pfx.Err(err)
		}

		cr := csv.NewReader(fl)
		for {
			cols, err := cr.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return pfx.Err(err)
			}

			if len(cols) < 1 {
				continue
			}

			if len(cols[0]) < 1 {
				continue
			}

			files = append(files, cols[0])
		}

	} else {
		// No file list - process all of the items within the folder

		fileInfos, err := ioutil.ReadDir(path)
		if err != nil {
			return pfx.Err(err)
		}

		for _, f := range fileInfos {
			if f.IsDir() {
				continue
			}

			files = append(files, f.Name())
		}
	}

	fmt.Fprintf(STDOUT, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
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
		"acquisition_time",
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

	// Ensure a path separator if we are using a path
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	semaphore := make(chan struct{}, concurrency)

	for _, file := range files {

		// Will block after `concurrency` simultaneous goroutines are running
		semaphore <- struct{}{}

		go func(file string) {

			// Be sure to permit unblocking once we finish
			defer func() { <-semaphore }()

			if !strings.HasSuffix(file, ".zip") {
				return
			}

			err := bulkprocess.CardiacMRIZipIterator(path+file, func(dcm bulkprocess.DicomOutput) error {
				if err := PrintCSVRow(dcm, results); err != nil {
					log.Printf("Error parsing %+v\n", dcm)
					return err
				}

				return nil
			})
			if err != nil {
				log.Println("Error parsing", path+file)
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
