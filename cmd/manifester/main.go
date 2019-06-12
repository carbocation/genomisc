package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"broad/ghgwas/ukbb/bulkprocess"
)

var (
	BufferSize = 4096
	STDOUT     = bufio.NewWriterSize(os.Stdout, BufferSize)
)

// Makes one big combined manifest
// Emits to stdout
func main() {
	defer STDOUT.Flush()

	var path string

	flag.StringVar(&path, "path", "./", "Path where the UKBB bulk .zip files are being held.")

	flag.Parse()

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintf(STDOUT, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
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
				log.Fatalln(err)
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
}

func PrintCSVRow(row bulkprocess.DicomOutput, results chan<- string) error {
	studyDate, err := row.ParsedDate()
	if err != nil {
		return err
	}

	overlayText := "NoOverlay"
	if row.DicomMeta.HasOverlay {
		overlayText = "HasOverlay"
	}

	results <- fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%.8f\t%d\t%d\t%d\t%d\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%s\t%s\t%s\t%s\t%s",
		row.SampleID, row.FieldID, row.Instance, row.Index, row.ZipFile,
		row.Filename, row.DicomMeta.SeriesDescription, studyDate.Format("2006-01-02"),
		row.DicomMeta.InstanceNumber, overlayText, row.DicomMeta.OverlayFraction, row.DicomMeta.OverlayRows, row.DicomMeta.OverlayCols,
		row.DicomMeta.Rows, row.DicomMeta.Cols,
		row.DicomMeta.PatientX, row.DicomMeta.PatientY, row.DicomMeta.PatientZ, row.DicomMeta.PixelHeightMM, row.DicomMeta.PixelWidthMM,
		row.DicomMeta.SliceThicknessMM,
		row.DicomMeta.SeriesNumber, row.DicomMeta.AcquisitionNumber, row.DicomMeta.DeviceSerialNumber, row.DicomMeta.StationName, row.DicomMeta.SoftwareVersions)
	return nil
}
