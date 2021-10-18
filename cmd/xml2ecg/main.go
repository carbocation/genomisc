package main

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

const (
	// SamplingHz is the frequency, in cycles per second, at which the signal is
	// sampled. TODO: fetch from the data itself.
	SamplingHz = 500.0

	// PointsCompression represents the number of points to compress together.
	// When <= 1.0, it will allow all points to be individually represented. If
	// 2.0 or greater (N), it will average N points together.
	PointsCompression = 1.0
)

// See https://www.ahajournals.org/doi/pdf/10.1161/CIRCULATIONAHA.106.180200 for
// recommendations

var (
	// LowPassHz represents the frequency above which signals will be blocked
	// (i.e., 'pass' signals lower than this frequency, in cycles per second)
	LowPassHz float64

	// HighPassHz represents the frequency below which signals will be blocked
	// (i.e., 'pass' signals higher than this frequency, in cycles per second)
	HighPassHz float64
)

func main() {
	var filename string
	var createPNG bool
	var printDiagnoses bool
	var debug bool
	var widthPx, heightPx int

	flag.StringVar(&filename, "file", "", "XML file (CardioSoft 6.73 output)")
	flag.BoolVar(&createPNG, "createpng", false, "Create PNG representations of the EKG strips?")
	flag.BoolVar(&printDiagnoses, "diagnoses", false, "Emit the automated diagnoses to a _diagnoses.csv file?")
	flag.IntVar(&widthPx, "width", 256, "(Optional) If creating PNGs, what pixel width?")
	flag.IntVar(&heightPx, "height", 256, "(Optional) If creating PNGs, what pixel height?")
	flag.BoolVar(&debug, "debug", false, "Print extra metadata during processing?")
	flag.Float64Var(&LowPassHz, "low_pass_hz", 150.0, "Only permit frequencies below this many cycles per second")
	flag.Float64Var(&HighPassHz, "high_pass_hz", 0.05, "Only permit frequencies above this many cycles per second")
	flag.Parse()

	if filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := run(filename, createPNG, printDiagnoses, widthPx, heightPx, debug); err != nil {
		log.Fatalln(err)
	}
}

func run(filename string, createPNG, printDiagnoses bool, widthPx, heightPx int, debug bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Because the file is encoded as "ISO-8859-1" and not UTF-8, we need to use
	// the charset.NewReaderLabel
	var doc CardiologyXML
	decoder := xml.NewDecoder(f)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&doc); err != nil {
		return err
	}

	if printDiagnoses {
		if err := processDiagnoses(filepath.Base(filename), doc); err != nil {
			return err
		}
	}

	if err := processFullDisclosureStrip(filepath.Base(filename), doc, createPNG, widthPx, heightPx, debug); err != nil {
		return err
	}

	if err := processSummaryStrip(filepath.Base(filename), doc, createPNG, widthPx, heightPx, debug); err != nil {
		return err
	}

	return nil
}

func processDiagnoses(filename string, doc CardiologyXML) error {
	outFile, err := os.Create(strings.TrimSuffix(filename, ".xml") + "_diagnoses.tsv")
	if err != nil {
		return err
	}
	defer outFile.Close()

	var OUTFILE = bufio.NewWriter(outFile)
	defer OUTFILE.Flush()

	sampleID, instance, err := fileNameToSampleInstance(filename)
	if err != nil {
		return err
	}

	nDiags := len(doc.Interpretation.Diagnosis.DiagnosisText)
	aboveDash := "above"
	for i, val := range doc.Interpretation.Diagnosis.DiagnosisText {
		if val == "---" {
			aboveDash = "dash"
		} else if val == "Arrhythmia results of the full-disclosure ECG" {
			aboveDash = "arotfde"
		} else if aboveDash != "above" {
			aboveDash = "below"
		}

		fmt.Fprintln(OUTFILE, strings.Join([]string{
			sampleID,
			instance,
			doc.ClinicalInfo.DeviceInfo.SoftwareVer,
			doc.RestingECGMeasurements.DiagnosisVersion,
			"diagnosis",
			strconv.Itoa(i),
			strconv.Itoa(nDiags),
			aboveDash,
			strings.TrimSpace(val),
		}, "\t"))
	}

	nConclusions := len(doc.Interpretation.Conclusion.ConclusionText)
	aboveDash = "above"
	for i, val := range doc.Interpretation.Conclusion.ConclusionText {
		if val == "---" {
			aboveDash = "dash"
		} else if val == "Arrhythmia results of the full-disclosure ECG" {
			aboveDash = "arotfde"
		} else if aboveDash != "above" {
			aboveDash = "below"
		}

		fmt.Fprintln(OUTFILE, strings.Join([]string{
			sampleID,
			instance,
			doc.ClinicalInfo.DeviceInfo.SoftwareVer,
			doc.RestingECGMeasurements.DiagnosisVersion,
			"conclusion",
			strconv.Itoa(i),
			strconv.Itoa(nConclusions),
			aboveDash,
			strings.TrimSpace(val),
		}, "\t"))
	}

	return nil
}

func processSummaryStrip(filename string, doc CardiologyXML, createPNG bool, widthPx, heightPx int, debug bool) error {
	outFile, err := os.Create(strings.TrimSuffix(filename, ".xml") + "_summary.csv")
	if err != nil {
		return err
	}
	defer outFile.Close()

	var OUTFILE = bufio.NewWriter(outFile)
	defer OUTFILE.Flush()

	if debug {
		fmt.Println(doc.RestingECGMeasurements.MedianSamples.SampleRate)
	}

	var zw *zip.Writer

	if createPNG {
		outFile, err := os.Create(strings.TrimSuffix(filename, ".xml") + "_summary.zip")
		if err != nil {
			return err
		}
		defer outFile.Close()

		zw = zip.NewWriter(outFile)
		defer zw.Close()
	}

	// Grouped by lead
	for _, v := range doc.RestingECGMeasurements.MedianSamples.WaveformData {

		// The raw data contains whitespace, newlines, and tabs which we remove
		// here.
		txt := strings.TrimSpace(v.Text)
		txt = strings.ReplaceAll(txt, "\n", "")
		txt = strings.ReplaceAll(txt, "\t", "")

		vals := strings.Split(txt, ",")

		// Fail if we haven't sucessfully stripped out all non-numeric characters
		for j, measurement := range vals {
			if _, err := strconv.Atoi(measurement); err != nil {
				return fmt.Errorf("Measurement %d is not numeric and is instead [%s]", j, measurement)
			}
		}

		if debug {
			fmt.Println(v.Lead, len(vals), vals[0:20])
		}

		if createPNG {
			voltageCorrection := EstimateVoltageCorrection(doc.RestingECGMeasurements.MedianSamples.Resolution.Text, doc.RestingECGMeasurements.MedianSamples.Resolution.Units)

			vf, err := stringSliceToFloatSlice(vals)
			if err != nil {
				return err
			}

			for i := range vf {
				// Voltage correction
				vf[i] *= voltageCorrection
			}

			gray, err := PlotLeadFloat(-1.0, 1.0, widthPx, heightPx, vf)
			if err != nil {
				return err
			}
			imgW, err := zw.Create(strings.TrimSuffix(filename, ".xml") + "_summary_" + v.Lead + ".png")
			if err != nil {
				return err
			}

			if err := png.Encode(imgW, gray); err != nil {
				return err
			}
		}

		sampleID, instance, err := fileNameToSampleInstance(filename)
		if err != nil {
			return err
		}
		fmt.Fprintln(OUTFILE, strings.Join(append([]string{sampleID, instance, v.Lead, doc.RestingECGMeasurements.MedianSamples.Resolution.Text, doc.RestingECGMeasurements.MedianSamples.SampleRate.Text}, vals...), ","))
	}

	return nil
}

func processFullDisclosureStrip(filename string, doc CardiologyXML, createPNG bool, widthPx, heightPx int, debug bool) error {
	outFile, err := os.Create(strings.TrimSuffix(filename, ".xml") + "_full.csv")
	if err != nil {
		return err
	}
	defer outFile.Close()

	var OUTFILE = bufio.NewWriter(outFile)
	defer OUTFILE.Flush()

	if debug {
		fmt.Println(doc.StripData.Resolution)
		fmt.Println(doc.StripData.SampleRate)
	}

	var zw *zip.Writer

	if createPNG {
		outFile, err := os.Create(strings.TrimSuffix(filename, ".xml") + "_full.zip")
		if err != nil {
			return err
		}
		defer outFile.Close()

		zw = zip.NewWriter(outFile)
		defer zw.Close()
	}

	// Grouped by lead
	for _, v := range doc.StripData.WaveformData {

		// The raw data contains whitespace, newlines, and tabs which we remove
		// here.
		txt := strings.TrimSpace(v.Text)
		txt = strings.ReplaceAll(txt, "\n", "")
		txt = strings.ReplaceAll(txt, "\t", "")

		vals := strings.Split(txt, ",")

		// Fail if we haven't sucessfully stripped out all non-numeric characters
		for j, measurement := range vals {
			if _, err := strconv.Atoi(measurement); err != nil {
				return fmt.Errorf("Measurement %d is not numeric and is instead [%s]", j, measurement)
			}
		}

		if debug {
			fmt.Println(v.Lead, len(vals), vals[0:20])
		}

		if createPNG {

			voltageCorrection := EstimateVoltageCorrection(doc.StripData.Resolution.Text, doc.StripData.Resolution.Units)

			vf, err := stringSliceToFloatSlice(vals)
			if err != nil {
				return err
			}

			for i := range vf {
				// Voltage correction
				vf[i] *= voltageCorrection
			}

			valsFloat, err := BandPassFilter(vf, SamplingHz, HighPassHz, LowPassHz, PointsCompression)
			if err != nil {
				return err
			}

			gray, err := PlotLeadFloat(-1.0, 1.0, widthPx, heightPx, valsFloat)
			if err != nil {
				return err
			}
			imgW, err := zw.Create(strings.TrimSuffix(filename, ".xml") + "_full_" + v.Lead + ".png")
			if err != nil {
				return err
			}

			if err := png.Encode(imgW, gray); err != nil {
				return err
			}

		}

		sampleID, instance, err := fileNameToSampleInstance(filename)
		if err != nil {
			return err
		}

		fmt.Fprintln(OUTFILE, strings.Join(append([]string{sampleID, instance, v.Lead, doc.StripData.Resolution.Text, doc.StripData.SampleRate.Text}, vals...), ","))
	}

	return nil
}
