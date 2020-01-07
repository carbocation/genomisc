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

type List struct {
	*ring.Ring
}

func entryValue(entry interface{}) Entry {
	return entry.(Entry)
}

func (l *List) GetAdjacent(n int) []Entry {
	out := make([]Entry, 0, 1+2*n)

	out = append(out, entryValue(l.Value))

	current := l.Ring
	for i := 0; i < n; i++ {
		current = current.Prev()
		out = append(out, entryValue(current.Value))
	}

	current = l.Ring
	for i := 0; i < n; i++ {
		current = current.Next()
		out = append(out, entryValue(current.Value))
	}

	return out
}

type Entry struct {
	InstanceNumber uint16
	Metric         float64
}

func discardExtremes(entries []Entry, discardN int) ([]Entry, error) {
	if discardN >= len(entries) {
		return nil, fmt.Errorf("Tried to discard %d but only have %d", discardN, len(entries))
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Metric < entries[j].Metric
	})

	out := make([]Entry, 0, len(entries)-2*discardN)
	for i := discardN; i < len(entries)-discardN; i++ {
		out = append(out, entries[i])
	}

	return out, nil
}

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

func median(entries []Entry) float64 {
	floats := make([]float64, 0, len(entries))
	for _, v := range entries {
		floats = append(floats, v.Metric)
	}

	sort.Float64s(floats)

	mIdx := len(floats) / 2

	if len(floats)%2 == 1 {
		return floats[mIdx]
	}

	return (floats[mIdx-1] + floats[mIdx]) / 2.0

}

type synthEntry struct {
	InstanceNumber uint16
	Metric         float64
	TrueMetric     float64
}

func (l *List) Extrema(adjacentN, discardN int) (Result, error) {
	out := Result{}

	synthetic := make([]synthEntry, 0, l.Len())

	for i := 0; i < l.Len(); i++ {
		thisEntry := entryValue(l.Value)

		adj := l.GetAdjacent(2)
		mapped, err := discardExtremes(adj, 1)
		if err != nil {
			return out, err
		}

		synthEntry := synthEntry{InstanceNumber: thisEntry.InstanceNumber, Metric: median(mapped), TrueMetric: thisEntry.Metric}
		synthetic = append(synthetic, synthEntry)

		l.Ring = l.Next()
	}

	sort.Slice(synthetic, func(i, j int) bool {
		return synthetic[i].Metric < synthetic[j].Metric
	})

	max := synthetic[len(synthetic)-1]
	min := synthetic[0]

	out.InstanceNumberAtMax = max.InstanceNumber
	out.InstanceNumberAtMin = min.InstanceNumber
	out.Max = max.TrueMetric
	out.Min = min.TrueMetric
	out.Discards = discardN
	out.Window = adjacentN

	return out, nil
}

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
	sampleMap, colName, err := fetchSampleMap(input)
	if err != nil {
		return err
	}

	ringMap, err := sampleMapToRing(sampleMap)
	if err != nil {
		return err
	}

	countMap := make(map[int]int)
	for _, v := range sampleMap {
		// if len(countMap) == 0 {
		// 	log.Println(v)
		// }
		countMap[len(v)]++
	}

	// fmt.Println(len(sampleMap))
	// fmt.Println(countMap)
	// fmt.Println(len(ringMap))

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
		if len(counts) != 50 {
			continue
		}

		cl := &List{ring.New(50)}

		for i := 1; i <= 50; i++ {
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
