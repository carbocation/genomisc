package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

const (
	SamplingHz = 500.0
	LowPassHz  = 0.4
	HighPassHz = 40.0
)

func main() {
	var filename string
	var createPNG bool
	var debug bool
	var widthPx, heightPx int

	flag.StringVar(&filename, "file", "", "XML file (CardioSoft 6.73 output)")
	flag.BoolVar(&createPNG, "createpng", false, "Create PNG representations of the EKG strips?")
	flag.IntVar(&widthPx, "width", 256, "(Optional) If creating PNGs, what pixel width?")
	flag.IntVar(&heightPx, "height", 256, "(Optional) If creating PNGs, what pixel height?")
	flag.BoolVar(&debug, "debug", false, "Print extra metadata during processing?")
	flag.Parse()

	if filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := run(filename, createPNG, widthPx, heightPx, debug); err != nil {
		log.Fatalln(err)
	}
}

func run(filename string, createPNG bool, widthPx, heightPx int, debug bool) error {
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

	if err := processFullDisclosureStrip(filepath.Base(filename), doc, createPNG, widthPx, heightPx, debug); err != nil {
		return err
	}

	if err := processSummaryStrip(filepath.Base(filename), doc, createPNG, widthPx, heightPx, debug); err != nil {
		return err
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

			if err := PlotLeadFloat(filename, "AvgLead_"+v.Lead, -1.0, 1.0, widthPx, heightPx, vf); err != nil {
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

			valsFloat, err := BandPassFilter(vf, SamplingHz, LowPassHz, HighPassHz, 1)
			if err != nil {
				return err
			}

			if err := PlotLeadFloat(filename, "Lead_"+v.Lead, -1.0, 1.0, widthPx, heightPx, valsFloat); err != nil {
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
