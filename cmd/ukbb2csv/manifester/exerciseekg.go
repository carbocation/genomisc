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

const ExerciseEKGFieldIDChunk = "_6025_"

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

	fmt.Fprintln(STDOUT, strings.ReplaceAll(strings.Join(header, "\t"), ".", "_"))

	// Ensure a path separator if we are using a path
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}

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

			if !strings.Contains(file.Name(), ExerciseEKGFieldIDChunk) {
				return
			}
			// Exercise EKG

			ekg := bulkprocess.EKGExercise{}

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
				NA(ekg.ObservationType),
				NA(ekg.ObservationDateTime.Year),
				NA(ekg.ObservationDateTime.Month),
				NA(ekg.ObservationDateTime.Day),
				NA(ekg.ObservationDateTime.Hour),
				NA(testDuration),
				NA(ekg.ClinicalInfo.DeviceInfo.Desc),
				NA(ekg.ClinicalInfo.DeviceInfo.SoftwareVer),
				NA(ekg.ClinicalInfo.DeviceInfo.AnalysisVer),
				NA(ekg.PatientVisit.AssignedPatientLocation.Facility),
				NA(ekg.PatientVisit.AssignedPatientLocation.LocationNumber),
				NA(ekg.PatientInfo.PaceMaker),

				// Only relevant for exercise
				NA(ekg.Protocol.Device),
				NA(ekg.ExerciseMeasurements.ExercisePhaseTime.Minute),
				NA(ekg.ExerciseMeasurements.MaxWorkload.Text),
				NA(ekg.ExerciseMeasurements.RestingStats.RestHR),
				NA(ekg.ExerciseMeasurements.MaxHeartRate),
				NA(ekg.ExerciseMeasurements.MaxPredictedHR),
				NA(ekg.ExerciseMeasurements.PercentAchievedMaxPredicted),
				NA(ekg.TWACycleData.Resolution.Text),
				NA(ekg.TWACycleData.Resolution.Units),
				NA(ekg.TWACycleData.SampleRate.Text),
				NA(ekg.TWACycleData.SampleRate.Units),
				NA(ekg.TrendData.NumberOfEntries),
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
