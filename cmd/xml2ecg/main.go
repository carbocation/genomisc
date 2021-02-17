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

	flag.StringVar(&filename, "file", "", "XML file (CardioSoft 6.73 output)")
	flag.Parse()

	if filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := run(filename); err != nil {
		log.Fatalln(err)
	}
}

func run(filename string) error {
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

	if err := processFullDisclosureSrip(filepath.Base(filename), doc); err != nil {
		return err
	}

	if err := processSummaryStrip(filepath.Base(filename), doc); err != nil {
		return err
	}

	return nil
}

func processSummaryStrip(filename string, doc CardiologyXML) error {
	outFile, err := os.Create(strings.TrimSuffix(filename, ".xml") + "_summary.csv")
	if err != nil {
		return err
	}
	defer outFile.Close()

	var OUTFILE = bufio.NewWriter(outFile)
	defer OUTFILE.Flush()

	fmt.Println(doc.RestingECGMeasurements.MedianSamples.SampleRate)

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

		fmt.Println(v.Lead, len(vals), vals[0:20])

		if err := PlotLead(filename, "AvgLead_"+v.Lead, vals); err != nil {
			return err
		}

		fmt.Fprintln(OUTFILE, strings.Join(vals, ","))
	}

	return nil
}

func processFullDisclosureSrip(filename string, doc CardiologyXML) error {
	outFile, err := os.Create(strings.TrimSuffix(filename, ".xml") + "_full.csv")
	if err != nil {
		return err
	}
	defer outFile.Close()

	var OUTFILE = bufio.NewWriter(outFile)
	defer OUTFILE.Flush()

	fmt.Println(doc.StripData.Resolution)
	fmt.Println(doc.StripData.SampleRate)

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

		fmt.Println(v.Lead, len(vals), vals[0:20])

		if err := PlotLead(filename, "Lead_"+v.Lead, vals); err != nil {
			return err
		}

		fmt.Fprintln(OUTFILE, strings.Join(vals, ","))
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
