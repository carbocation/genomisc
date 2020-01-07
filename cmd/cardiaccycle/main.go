package main

import (
	"container/ring"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type Result struct {
	SampleID            string
	Column              string
	InstanceNumberAtMin uint16
	Min                 float64
	InstanceNumberAtMax uint16
	Max                 float64
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
	InstanceNumber
	Metric
)

func main() {
	var input string
	flag.StringVar(&input, "file", "", "Comma-delimited file with a header and 3 columns in order: sample_id, instance_number, and a metric")

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
		res.SampleID = sampleID
		res.Column = colName

		results = append(results, res)
	}

	fmt.Println(strings.Join([]string{"sample_id", "column", "instance_number_at_min", "min", "instance_number_at_max", "max", "window", "discards"}, "\t"))
	for _, v := range results {
		fmt.Printf("%s\t%s\t%d\t%f\t%d\t%f\t%d\t%d\n", v.SampleID, v.Column, v.InstanceNumberAtMin, v.Min, v.InstanceNumberAtMax, v.Max, v.Window, v.Discards)
	}

	return nil
}

func sampleMapToRing(sampleMap map[string]map[uint16]float64) (map[string]*List, error) {
	out := make(map[string]*List)

	for sampleID, counts := range sampleMap {
		cl := &List{ring.New(len(counts))}

		for i := 1; i <= cl.Len(); i++ {
			cl.Ring.Value = Entry{InstanceNumber: uint16(i), Metric: counts[uint16(i)]}
			cl.Ring = cl.Next()
		}

		out[sampleID] = cl
	}

	return out, nil
}

func fetchSampleMap(input string) (map[string]map[uint16]float64, string, error) {
	colName := ""

	f, err := os.Open(input)
	if err != nil {
		return nil, colName, err
	}

	sampleMap := make(map[string]map[uint16]float64) // map[sample_id] contains => map[instance_number]metric

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
			colName = line[2]
			continue
		}

		if len(line) != 3 {
			return nil, colName, fmt.Errorf("Expected 3 columns, got %d", len(line))
		}

		instanceMap := sampleMap[line[SampleID]]
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

		sampleMap[line[SampleID]] = instanceMap
	}

	return sampleMap, colName, nil
}
