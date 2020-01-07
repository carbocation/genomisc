package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

const (
	SampleID = iota
	InstanceNumber
	Metric
)

func main() {
	var input string
	flag.StringVar(&input, "file", "", "Comma-delimited file with 3 columns: sample_id, instance_number, and a metric")

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
	sampleMap, err := fetchSampleMap(input)
	if err != nil {
		return err
	}

	countMap := make(map[int]int)
	for _, v := range sampleMap {
		if len(countMap) == 0 {
			log.Println(v)
		}
		countMap[len(v)]++
	}

	fmt.Println(len(sampleMap))
	fmt.Println(countMap)

	return nil
}

func fetchSampleMap(input string) (map[string]map[uint16]float64, error) {
	f, err := os.Open(input)
	if err != nil {
		return nil, err
	}

	sampleMap := make(map[string]map[uint16]float64) // map[sample_id] contains => map[instance_number]metric

	r := csv.NewReader(f)
	for i := 0; ; i++ {
		line, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if i == 0 {
			// Skip the header
			continue
		}

		if len(line) != 3 {
			return nil, fmt.Errorf("Expected 3 columns, got %d", len(line))
		}

		instanceMap := sampleMap[line[SampleID]]
		if instanceMap == nil {
			instanceMap = make(map[uint16]float64)
		}

		inst, err := strconv.ParseUint(line[InstanceNumber], 10, 64)
		if err != nil {
			return nil, err
		}

		metric, err := strconv.ParseFloat(line[Metric], 64)
		if err != nil {
			return nil, err
		}

		instanceMap[uint16(inst)] = metric

		sampleMap[line[SampleID]] = instanceMap
	}

	return sampleMap, nil
}
