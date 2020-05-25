package main

import (
	"fmt"
	"sort"
)

type Entry struct {
	InstanceNumber uint16
	Metric         float64
}

func entryValue(entry interface{}) Entry {
	return entry.(Entry)
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
