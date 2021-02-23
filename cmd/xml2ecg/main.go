package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wcharczuk/go-chart/v2"
	"golang.org/x/net/html/charset"
)

func main() {
	var filename string
	var createPNG bool
	var debug bool

	flag.StringVar(&filename, "file", "", "XML file (CardioSoft 6.73 output)")
	flag.BoolVar(&createPNG, "createpng", false, "Create PNG representations of the EKG strips?")
	flag.BoolVar(&debug, "debug", false, "Print extra metadata during processing?")
	flag.Parse()

	if filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := run(filename, createPNG, debug); err != nil {
		log.Fatalln(err)
	}
}

func run(filename string, createPNG, debug bool) error {
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

	if err := processFullDisclosureStrip(filepath.Base(filename), doc, createPNG, debug); err != nil {
		return err
	}

	if err := processSummaryStrip(filepath.Base(filename), doc, createPNG, debug); err != nil {
		return err
	}

	return nil
}

func processSummaryStrip(filename string, doc CardiologyXML, createPNG, debug bool) error {
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
			if err := PlotLead(filename, "AvgLead_"+v.Lead, vals); err != nil {
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

func processFullDisclosureStrip(filename string, doc CardiologyXML, createPNG, debug bool) error {
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
			if err := PlotLead(filename, "Lead_"+v.Lead, vals); err != nil {
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

func PlotLead(filename, lead string, vals []string) error {
	floatVals, err := stringSliceToFloatSlice(vals)
	if err != nil {
		return err
	}

	graph := chart.Chart{
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: intSeq(len(floatVals)),
				YValues: floatVals,
			},
		},
	}

	// Render to a byte buffer
	buffer := bytes.NewBuffer([]byte{})
	if err := graph.Render(chart.PNG, buffer); err != nil {
		return err
	}

	outFile, err := os.Create(strings.TrimSuffix(filename, ".xml") + "_" + lead + ".png")
	if err != nil {
		return err
	}
	if _, err := buffer.WriteTo(outFile); err != nil {
		return err
	}

	return nil
}

// This is very UK Biobank-specific, but the files are stored as
// sample_field_instance_arrayidx and so we can extract metadata from them.
func fileNameToSampleInstance(filename string) (string, string, error) {
	parts := strings.Split(filename, "_")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("Expected at least two '_' characters in the filename, but found %d", len(parts)-1)
	}

	return parts[0], parts[2], nil
}

func stringSliceToFloatSlice(in []string) ([]float64, error) {
	output := make([]float64, len(in))
	for i, fv := range in {
		// output[i] = float64(fv)
		v, err := strconv.ParseFloat(fv, 64)
		if err != nil {
			return nil, err
		}
		output[i] = v
	}

	return output, nil
}

func intSeq(N int) []float64 {
	output := make([]float64, N)
	for i := 0; i < N; i++ {
		output[i] = float64(i)
	}

	return output
}
