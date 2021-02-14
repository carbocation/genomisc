package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/pfx"
	"golang.org/x/net/html/charset"
)

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
