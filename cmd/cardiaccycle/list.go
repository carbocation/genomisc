package main

import (
	"container/ring"
	"sort"
)

type List struct {
	*ring.Ring
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
	out.SmoothedMax = max.Metric
	out.Min = min.TrueMetric
	out.SmoothedMin = min.Metric
	out.Discards = discardN
	out.Window = adjacentN

	return out, nil
}
