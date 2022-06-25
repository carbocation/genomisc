package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
	"golang.org/x/net/html/charset"
)

const EKGFieldIDChunk = "_20205_"

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

	fmt.Fprintln(STDOUT, strings.ReplaceAll(strings.Join(header, "\t"), ".", "_"))

	// Ensure a path separator if we are using a path
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	for _, file := range files {

		// Will block after `concurrency` simultaneous goroutines are running
		semaphore <- struct{}{}

		go func(file os.FileInfo) {

			// Be sure to permit unblocking once we finish
			defer func() { <-semaphore }()

			if !strings.HasSuffix(file.Name(), ".xml") {
				return
			}

			if !strings.Contains(file.Name(), EKGFieldIDChunk) {
				return
			}
			// Exercise EKG

			ekg := bulkprocess.EKG12Lead{}

			f, err := os.Open(path + file.Name())
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
				NA(ekg.ObservationType),
				NA(ekg.ObservationDateTime.Year),
				NA(ekg.ObservationDateTime.Month),
				NA(ekg.ObservationDateTime.Day),
				NA(ekg.ObservationDateTime.Hour),
				NA(ekg.ClinicalInfo.DeviceInfo.Desc),
				NA(ekg.ClinicalInfo.DeviceInfo.SoftwareVer),
				NA(ekg.ClinicalInfo.DeviceInfo.AnalysisVer),
				NA(ekg.PatientInfo.PaceMaker),

				// Only relevant for 12-lead
				NA(ekg.DeviceType),
				NA(ekg.RestingECGMeasurements.SokolovLVHIndex.Text),
				NA(ekg.RestingECGMeasurements.SokolovLVHIndex.Units),
				NA(ekg.RestingECGMeasurements.MeasurementTable.LeadOrder),
				NA(ekg.RestingECGMeasurements.VentricularRate.Text),
				NA(ekg.RestingECGMeasurements.PQInterval.Text),
				NA(ekg.RestingECGMeasurements.PDuration.Text),
				NA(ekg.RestingECGMeasurements.QRSDuration.Text),
				NA(ekg.RestingECGMeasurements.QTInterval.Text),
				NA(ekg.RestingECGMeasurements.QTCInterval.Text),
				NA(ekg.RestingECGMeasurements.RRInterval.Text),
				NA(ekg.RestingECGMeasurements.PPInterval.Text),
				NA(ekg.RestingECGMeasurements.PAxis.Text),
				NA(ekg.RestingECGMeasurements.RAxis.Text),
				NA(ekg.RestingECGMeasurements.TAxis.Text),
				NA(ekg.RestingECGMeasurements.QTDispersion.Text),
				NA(ekg.RestingECGMeasurements.QTDispersionBazett.Text),
				NA(ekg.RestingECGMeasurements.QRSNum),
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
