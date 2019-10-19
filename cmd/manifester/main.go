package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/html/charset"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
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
	var filetypes string

	flag.StringVar(&path, "path", "./", "Path where the UKBB bulk .zip files are being held.")
	flag.StringVar(&filetypes, "type", "dicomzip", "File type. Options include 'dicomzip' (default), '12leadekg', and 'exerciseekg' (for EKG data).")
	flag.Parse()

	if filetypes == "exerciseekg" {
		if err := ManifestForExerciseEKG(path); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if filetypes == "12leadekg" {
		if err := ManifestFor12LeadEKG(path); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if filetypes != "dicomzip" {
		log.Println("Requested filetype '' not recognized")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := ManifestForDicom(path); err != nil {
		log.Fatalln(err)
	}
}

func ManifestFor12LeadEKG(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return pfx.Err(err)
	}

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

	header := []string{
		"sample_id",
		"FieldID",
		"instance",
		"xml_file",
		"ObservationType",
		"ObservationDate",
		"ObservationDateTime.Hour",
		"ClinicalInfo.DeviceInfo.Desc",
		"ClinicalInfo.DeviceInfo.SoftwareVer",
		"ClinicalInfo.DeviceInfo.AnalysisVer",
		"PatientInfo.PaceMaker",

		// Only relevant for 12-lead
		"DeviceType",
		"RestingECGMeasurements.SokolovLVHIndex.Text",
		"RestingECGMeasurements.SokolovLVHIndex.Units",
		"RestingECGMeasurements.MeasurementTable.LeadOrder",
		"RestingECGMeasurements.VentricularRate.Text",
		"RestingECGMeasurements.PQInterval.Text",
		"RestingECGMeasurements.PDuration.Text",
		"RestingECGMeasurements.QRSDuration.Text",
		"RestingECGMeasurements.QTInterval.Text",
		"RestingECGMeasurements.QTCInterval.Text",
		"RestingECGMeasurements.RRInterval.Text",
		"RestingECGMeasurements.PPInterval.Text",
		"RestingECGMeasurements.PAxis.Text",
		"RestingECGMeasurements.RAxis.Text",
		"RestingECGMeasurements.TAxis.Text",
		"RestingECGMeasurements.QTDispersion.Text",
		"RestingECGMeasurements.QTDispersionBazett.Text",
		"RestingECGMeasurements.QRSNum",
	}

	fmt.Println(strings.ReplaceAll(strings.Join(header, "\t"), ".", "_"))

	for _, file := range files {

		// Will block after `concurrency` simultaneous goroutines are running
		semaphore <- struct{}{}

		go func(file os.FileInfo) {

			// Be sure to permit unblocking once we finish
			defer func() { <-semaphore }()

			if !strings.HasSuffix(file.Name(), ".xml") {
				return
			}

			if !strings.Contains(file.Name(), "_20205_") {
				return
			}
			// Exercise EKG

			ekg := bulkprocess.EKG12Lead{}

			f, err := os.Open(file.Name())
			if err != nil {
				log.Println(err)
				return
			}
			defer f.Close()

			decoder := xml.NewDecoder(f)
			decoder.CharsetReader = charset.NewReaderLabel

			if err := decoder.Decode(&ekg); err != nil {
				log.Println(err)
				return
			}

			results <- fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s-%02s-%02s\t%02s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
				strings.Split(file.Name(), "_")[0],
				strings.Split(file.Name(), "_")[1],
				strings.Split(file.Name(), "_")[2],
				file.Name(),
				ekg.ObservationType,
				ekg.ObservationDateTime.Year,
				ekg.ObservationDateTime.Month,
				ekg.ObservationDateTime.Day,
				ekg.ObservationDateTime.Hour,
				ekg.ClinicalInfo.DeviceInfo.Desc,
				ekg.ClinicalInfo.DeviceInfo.SoftwareVer,
				ekg.ClinicalInfo.DeviceInfo.AnalysisVer,
				ekg.PatientInfo.PaceMaker,

				// Only relevant for 12-lead
				ekg.DeviceType,
				ekg.RestingECGMeasurements.SokolovLVHIndex.Text,
				ekg.RestingECGMeasurements.SokolovLVHIndex.Units,
				ekg.RestingECGMeasurements.MeasurementTable.LeadOrder,
				ekg.RestingECGMeasurements.VentricularRate.Text,
				ekg.RestingECGMeasurements.PQInterval.Text,
				ekg.RestingECGMeasurements.PDuration.Text,
				ekg.RestingECGMeasurements.QRSDuration.Text,
				ekg.RestingECGMeasurements.QTInterval.Text,
				ekg.RestingECGMeasurements.QTCInterval.Text,
				ekg.RestingECGMeasurements.RRInterval.Text,
				ekg.RestingECGMeasurements.PPInterval.Text,
				ekg.RestingECGMeasurements.PAxis.Text,
				ekg.RestingECGMeasurements.RAxis.Text,
				ekg.RestingECGMeasurements.TAxis.Text,
				ekg.RestingECGMeasurements.QTDispersion.Text,
				ekg.RestingECGMeasurements.QTDispersionBazett.Text,
				ekg.RestingECGMeasurements.QRSNum,
			)
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

func ManifestForExerciseEKG(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return pfx.Err(err)
	}

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

	header := []string{
		"sample_id",
		"FieldID",
		"instance",
		"xml_file",
		"ObservationType",
		"ObservationDate",
		"ekg.ObservationDateTime.Hour",
		"testDuration",
		"ClinicalInfo.DeviceInfo.Desc",
		"ClinicalInfo.DeviceInfo.SoftwareVer",
		"ClinicalInfo.DeviceInfo.AnalysisVer",
		"PatientVisit.AssignedPatientLocation.Facility",
		"PatientVisit.AssignedPatientLocation.LocationNumber",
		"PatientInfo.PaceMaker",

		// Only relevant for exercise
		"Protocol.Device",
		"ExerciseMeasurements.ExercisePhaseTime.Minute",
		"ExerciseMeasurements.MaxWorkload.Text",
		"ExerciseMeasurements.RestingStats.RestHR",
		"ExerciseMeasurements.MaxHeartRate",
		"ExerciseMeasurements.MaxPredictedHR",
		"ExerciseMeasurements.PercentAchievedMaxPredicted",
		"TWACycleData.Resolution.Text",
		"TWACycleData.Resolution.Units",
		"TWACycleData.SampleRate.Text",
		"TWACycleData.SampleRate.Units",
		"TrendData.NumberOfEntries",
	}

	fmt.Println(strings.ReplaceAll(strings.Join(header, "\t"), ".", "_"))

	semaphore := make(chan struct{}, concurrency)

	for _, file := range files {

		// Will block after `concurrency` simultaneous goroutines are running
		semaphore <- struct{}{}

		go func(file os.FileInfo) {

			// Be sure to permit unblocking once we finish
			defer func() { <-semaphore }()

			if !strings.HasSuffix(file.Name(), ".xml") {
				return
			}

			if !strings.Contains(file.Name(), "_6025_") {
				return
			}
			// Exercise EKG

			ekg := bulkprocess.EKGExercise{}

			f, err := os.Open(file.Name())
			if err != nil {
				log.Println(err)
				return
			}
			defer f.Close()

			decoder := xml.NewDecoder(f)
			decoder.CharsetReader = charset.NewReaderLabel

			if err := decoder.Decode(&ekg); err != nil {
				log.Println(err)
				return
			}

			testDuration := ""

			start, err := time.Parse(time.RFC3339Nano, fmt.Sprintf("%s-%02s-%02sT%02s:%02s:%02s.0Z",
				ekg.ObservationDateTime.Year,
				ekg.ObservationDateTime.Month,
				ekg.ObservationDateTime.Day,
				ekg.ObservationDateTime.Hour,
				ekg.ObservationDateTime.Minute,
				ekg.ObservationDateTime.Second))
			if err != nil {
				log.Println(err)
			} else {
				end, err := time.Parse(time.RFC3339Nano, fmt.Sprintf("%s-%02s-%02sT%02s:%02s:%02s.0Z",
					ekg.ObservationEndDateTime.Year,
					ekg.ObservationEndDateTime.Month,
					ekg.ObservationEndDateTime.Day,
					ekg.ObservationEndDateTime.Hour,
					ekg.ObservationEndDateTime.Minute,
					ekg.ObservationEndDateTime.Second))
				if err != nil {
					log.Println(err)
				} else {
					duration := end.Sub(start)
					testDuration = fmt.Sprintf("%.0f", duration.Seconds())
				}
			}

			results <- fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s-%02s-%02s\t%02s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
				strings.Split(file.Name(), "_")[0],
				strings.Split(file.Name(), "_")[1],
				strings.Split(file.Name(), "_")[2],
				file.Name(),
				ekg.ObservationType,
				ekg.ObservationDateTime.Year,
				ekg.ObservationDateTime.Month,
				ekg.ObservationDateTime.Day,
				ekg.ObservationDateTime.Hour,
				testDuration,
				ekg.ClinicalInfo.DeviceInfo.Desc,
				ekg.ClinicalInfo.DeviceInfo.SoftwareVer,
				ekg.ClinicalInfo.DeviceInfo.AnalysisVer,
				ekg.PatientVisit.AssignedPatientLocation.Facility,
				ekg.PatientVisit.AssignedPatientLocation.LocationNumber,
				ekg.PatientInfo.PaceMaker,

				// Only relevant for exercise
				ekg.Protocol.Device,
				ekg.ExerciseMeasurements.ExercisePhaseTime.Minute,
				ekg.ExerciseMeasurements.MaxWorkload.Text,
				ekg.ExerciseMeasurements.RestingStats.RestHR,
				ekg.ExerciseMeasurements.MaxHeartRate,
				ekg.ExerciseMeasurements.MaxPredictedHR,
				ekg.ExerciseMeasurements.PercentAchievedMaxPredicted,
				ekg.TWACycleData.Resolution.Text,
				ekg.TWACycleData.Resolution.Units,
				ekg.TWACycleData.SampleRate.Text,
				ekg.TWACycleData.SampleRate.Units,
				ekg.TrendData.NumberOfEntries,
			)
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

func ManifestForDicom(path string) error {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return pfx.Err(err)
	}

	fmt.Fprintf(STDOUT, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
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

func PrintCSVRow(row bulkprocess.DicomOutput, results chan<- string) error {
	studyDate, err := row.ParsedDate()
	if err != nil {
		return err
	}

	overlayText := "NoOverlay"
	if row.DicomMeta.HasOverlay {
		overlayText = "HasOverlay"
	}

	results <- fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%.8f\t%d\t%d\t%d\t%d\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%s\t%s\t%s\t%s\t%s\t%v",
		row.SampleID, row.FieldID, row.Instance, row.Index, row.ZipFile,
		row.Filename, row.DicomMeta.SeriesDescription, studyDate.Format("2006-01-02"),
		row.DicomMeta.InstanceNumber, overlayText, row.DicomMeta.OverlayFraction, row.DicomMeta.OverlayRows, row.DicomMeta.OverlayCols,
		row.DicomMeta.Rows, row.DicomMeta.Cols,
		row.DicomMeta.PatientX, row.DicomMeta.PatientY, row.DicomMeta.PatientZ, row.DicomMeta.PixelHeightMM, row.DicomMeta.PixelWidthMM,
		row.DicomMeta.SliceThicknessMM,
		row.DicomMeta.SeriesNumber, row.DicomMeta.AcquisitionNumber, row.DicomMeta.DeviceSerialNumber, row.DicomMeta.StationName, row.DicomMeta.SoftwareVersions,
		row.DicomMeta.EchoTime)
	return nil
}
