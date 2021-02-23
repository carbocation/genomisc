package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jfcg/butter"
	"github.com/wcharczuk/go-chart/v2"
)

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
