package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jfcg/butter"
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
			voltageCorrection := EstimateVoltageCorrection(doc.RestingECGMeasurements.MedianSamples.Resolution.Text, doc.RestingECGMeasurements.MedianSamples.Resolution.Units)

			vf, err := stringSliceToFloatSlice(vals)
			if err != nil {
				return err
			}

			for i := range vf {
				// Voltage correction
				vf[i] *= voltageCorrection
			}

			if err := PlotLeadFloat(filename, "AvgLead_"+v.Lead, -1.0, 1.0, vf); err != nil {
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

			voltageCorrection := EstimateVoltageCorrection(doc.StripData.Resolution.Text, doc.StripData.Resolution.Units)

			vf, err := stringSliceToFloatSlice(vals)
			if err != nil {
				return err
			}

			for i := range vf {
				// Voltage correction
				vf[i] *= voltageCorrection
			}

			valsFloat, err := BandPassFilter(vf, 500, 0.4, 40, 1)
			if err != nil {
				return err
			}

			if err := PlotLeadFloat(filename, "Lead_"+v.Lead, -1.0, 1.0, valsFloat); err != nil {
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

func BandPassFilter(vals []float64, signalHz, highPassHz, lowPassHz, timepointsToCompress float64) ([]float64, error) {

	wcBase := 2.0 * math.Pi * timepointsToCompress / signalHz

	filt := butter.NewHighPass1(highPassHz * wcBase)
	filtL := butter.NewLowPass1(lowPassHz * wcBase)

	if filt == nil {
		return nil, fmt.Errorf("Invalid high-pass filter (attempted wc=%f, but expect .0001 < wc && wc < 3.1415)", wcBase*highPassHz)
	}

	if filtL == nil {
		return nil, fmt.Errorf("Invalid low-pass filter (attempted wc=%f, but expect .0001 < wc && wc < 3.1415)", wcBase*lowPassHz)
	}

	valsFloat := make([]float64, 0, len(vals))
	compressor := make([]float64, 0)

	for _, vf := range vals {
		if float64(len(compressor)) < timepointsToCompress {
			compressor = append(compressor, vf)
			continue
		}

		vf = FloatAvg(compressor)
		compressor = nil

		valsFloat = append(valsFloat, filt.Next(filtL.Next(vf)))
	}

	return valsFloat, nil
}

func EstimateVoltageCorrection(value, units string) float64 {
	voltageCorrection := 1.0

	if units == "uVperLsb" {
		vc, err := strconv.ParseFloat(value, 64)
		if err == nil {
			voltageCorrection = 0.001 * vc
		}
	}

	return voltageCorrection
}

func FloatAvg(f []float64) float64 {
	out := 0.0
	for _, v := range f {
		out += v
	}

	if x := len(f); x > 0 {
		return out / float64(x)
	}

	return out
}

func PlotLead(filename, lead string, yMin, yMax float64, vals []string) error {
	floatVals, err := stringSliceToFloatSlice(vals)
	if err != nil {
		return err
	}

	return PlotLeadFloat(filename, lead, yMin, yMax, floatVals)
}

func PlotLeadFloat(filename, lead string, yMin, yMax float64, floatVals []float64) error {
	var chartRange *chart.ContinuousRange

	if yMin != yMax {
		chartRange = &chart.ContinuousRange{Min: yMin, Max: yMax}
	}

	graph := chart.Chart{
		Width:  512,
		Height: 256,
		XAxis: chart.XAxis{
			Style: chart.Hidden(),
		},
		YAxis: chart.YAxis{
			Style: chart.Hidden(),
			Range: chartRange,
		},
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
