package cardiaccycle

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func RunFromFile(input string) error {
	sampleMap, colName, err := fetchSampleMapFromFile(input)
	if err != nil {
		return err
	}

	ringMap, err := sampleMapToRing(sampleMap)
	if err != nil {
		return err
	}

	results := make([]Result, 0, len(ringMap))

	for sampleID, v := range ringMap {
		res, err := v.Extrema(2, 0)
		if err != nil {
			return err
		}
		res.Identifier = sampleID.Identifier
		res.Column = colName

		results = append(results, res)
	}

	fmt.Println(strings.Join([]string{
		"identifier",
		"column",
		"max_one_step_shift",
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
		fmt.Printf("%s\t%s\t%f\t%d\t%f\t%f\t%d\t%f\t%f\t%d\t%d\n",
			v.Identifier,
			v.Column,
			v.MaxOneStepShift,
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

func fetchSampleMapFromFile(input string) (sampleMap map[SeriesSample]map[uint16]float64, colName string, err error) {
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

		if len(line) < 3 {
			return nil, colName, fmt.Errorf("Expected >= 3 columns, got %d", len(line))
		}

		samp := SeriesSample{
			Identifier: line[Identifier],
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
