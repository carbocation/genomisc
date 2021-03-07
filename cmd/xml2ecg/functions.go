package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/jfcg/butter"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

func init() {
	chart.DefaultColors = []drawing.Color{chart.ColorBlack}
}

func BandPassFilter(vals []float64, signalHz, highPassHz, lowPassHz, timepointsToCompress float64) ([]float64, error) {

	wcBase := 2.0 * math.Pi * timepointsToCompress / signalHz

	filt := butter.NewHighPass1(highPassHz * wcBase)
	filtL := butter.NewLowPass1(lowPassHz * wcBase)

	if filt == nil {
		return nil, fmt.Errorf("Invalid high-pass filter (attempted wc=%f for highPassHz=%f, but expect .0001 < wc && wc < 3.1415)", wcBase*highPassHz, highPassHz)
	}

	if filtL == nil {
		return nil, fmt.Errorf("Invalid low-pass filter (attempted wc=%f for lowPassHz=%f, but expect .0001 < wc && wc < 3.1415)", wcBase*lowPassHz, lowPassHz)
	}

	valsFloat := make([]float64, 0, len(vals))
	compressor := make([]float64, 0)

	for _, vf := range vals {
		if timepointsToCompress < 2 {
			valsFloat = append(valsFloat, filt.Next(filtL.Next(vf)))
			continue
		}

		compressor = append(compressor, vf)
		if float64(len(compressor)) < timepointsToCompress {
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

// PlotLeadFloat yields a grayscale image
func PlotLeadFloat(yMin, yMax float64, widthPx, heightPx int, floatVals []float64) (image.Image, error) {
	var chartRange *chart.ContinuousRange

	if yMin != yMax {
		chartRange = &chart.ContinuousRange{Min: yMin, Max: yMax}
	}

	graph := chart.Chart{
		Width:  widthPx,
		Height: heightPx,
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
		return nil, err
	}

	// Fetch the img representation of the graph
	src, _, err := image.Decode(buffer)
	if err != nil {
		return nil, err
	}

	// Create a new grayscale image
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	gray := image.NewGray(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			oldColor := src.At(x, y)
			grayColor := color.GrayModel.Convert(oldColor)
			gray.Set(x, y, grayColor)
		}
	}

	return gray, nil
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
