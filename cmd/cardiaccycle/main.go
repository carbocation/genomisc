package main

import (
	"container/ring"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Result struct {
	SampleID            string
	SeriesNumber        string
	Column              string
	InstanceNumberAtMin uint16
	Min                 float64
	SmoothedMin         float64
	InstanceNumberAtMax uint16
	Max                 float64
	SmoothedMax         float64
	Window              int
	Discards            int
}

type synthEntry struct {
	InstanceNumber uint16
	Metric         float64
	TrueMetric     float64
}

const (
	SampleID = iota
	SeriesNumber
	InstanceNumber
	Metric
)

func main() {
	var input string
	flag.StringVar(&input, "file", "", "Comma-delimited file with a header and 4 columns in order: sample_id, series_number, instance_number, and a metric")

	flag.Parse()

	if input == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(input); err != nil {
		log.Fatalln(err)
	}
}

func run(input string) error {
	sampleMap, colName, err := fetchSampleMap(input)
	if err != nil {
		return err
	}

	ringMap, err := sampleMapToRing(sampleMap)
	if err != nil {
		return err
	}

	results := make([]Result, 0, len(ringMap))

	for sampleID, v := range ringMap {
		res, err := v.Extrema(2, 1)
		if err != nil {
			return err
		}
		res.SampleID = sampleID.SampleID
		res.SeriesNumber = sampleID.SeriesNumber
		res.Column = colName

		results = append(results, res)
	}

	fmt.Println(strings.Join([]string{
		"sample_id",
		"series_number",
		"column",
		"instance_number_at_min",
		"min",
		"smoothed_min",
		"instance_number_at_max",
		"max",
		"smoothed_max",
		"window",
		"discards"},
		"\t"))

	for _, v := range results {
		fmt.Printf("%s\t%s\t%s\t%d\t%f\t%f\t%d\t%f\t%f\t%d\t%d\n",
			v.SampleID,
			v.SeriesNumber,
			v.Column,
			v.InstanceNumberAtMin,
			v.Min,
			v.SmoothedMin,
			v.InstanceNumberAtMax,
			v.Max,
			v.SmoothedMax,
			v.Window,
			v.Discards,
		)
	}

	return nil
}

func sampleMapToRing(sampleMap map[SeriesSample]map[uint16]float64) (map[SeriesSample]*List, error) {
	out := make(map[SeriesSample]*List)

	for sampleID, counts := range sampleMap {
		cl := &List{ring.New(len(counts))}

		// Note: For some samples, the instance_number doesn't increase linearly
		// from 1 : X. In one example, the instance_number started at 51 and
		// continued to 100. So, can't assume that if there ar 100, they will go
		// 1-100 nicely.
		keys := make([]uint16, 0, cl.Len())
		for key := range counts {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

		// for i := 1; i <= cl.Len(); i++ {
		for _, v := range keys {
			cl.Ring.Value = Entry{InstanceNumber: v, Metric: counts[v]}
			cl.Ring = cl.Next()
		}

		out[sampleID] = cl
	}

	return out, nil
}

func fetchSampleMap(input string) (sampleMap map[SeriesSample]map[uint16]float64, colName string, err error) {
	f, err := os.Open(input)
	if err != nil {
		return nil, colName, err
	}

	sampleMap = make(map[SeriesSample]map[uint16]float64) // map[sample_id] contains => map[instance_number]metric

	r := csv.NewReader(f)
	for i := 0; ; i++ {
		line, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, colName, err
		}

		if i == 0 {
			// Skip the header
			colName = line[Metric]
			continue
		}

		if len(line) < 4 {
			return nil, colName, fmt.Errorf("Expected >= 4 columns, got %d", len(line))
		}

		samp := SeriesSample{
			SampleID:     line[SampleID],
			SeriesNumber: line[SeriesNumber],
		}

		instanceMap := sampleMap[samp]
		if instanceMap == nil {
			instanceMap = make(map[uint16]float64)
		}

		inst, err := strconv.ParseUint(line[InstanceNumber], 10, 64)
		if err != nil {
			return nil, colName, err
		}

		metric, err := strconv.ParseFloat(line[Metric], 64)
		if err != nil {
			return nil, colName, err
		}

		instanceMap[uint16(inst)] = metric

		sampleMap[samp] = instanceMap
	}

	return sampleMap, colName, nil
}
