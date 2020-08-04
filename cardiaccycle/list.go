package cardiaccycle

import (
	"container/ring"
	"math"
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

	lastPixelArea := 0.0
	for i := 0; i < l.Len(); i++ {
		thisEntry := entryValue(l.Value)

		adj := l.GetAdjacent(adjacentN)
		mapped, err := discardExtremes(adj, discardN)
		if err != nil {
			return out, err
		}

		synthEntry := synthEntry{InstanceNumber: thisEntry.InstanceNumber, Metric: median(mapped), TrueMetric: thisEntry.Metric}
		synthetic = append(synthetic, synthEntry)

		if i > 0 {
			if x := math.Abs(thisEntry.Metric - lastPixelArea); x > out.MaxOneStepShift {
				out.MaxOneStepShift = x
			}
		}
		lastPixelArea = thisEntry.Metric

		l.Ring = l.Next()
	}

	// Note that we are sorting on the TrueMetric, not the median metric. So,
	// while we do define a median metric, we don't record exactly where its max
	// and min are.
	sort.Slice(synthetic, func(i, j int) bool {
		return synthetic[i].TrueMetric < synthetic[j].TrueMetric
	})

	max := synthetic[len(synthetic)-1]
	min := synthetic[0]

	// Scale the onestepshift value to represent a fraction of the total range
	// starting at 0. Otherwise, people with absolutely higher values will look
	// like they have higher onestepshifts, but from a fractional basis they
	// might not. (e.g., 2->1 vs 5->2.5 both represent a 50% reduction, but
	// unless scaled, the 5->2.5 will look more extreme).
	out.MaxOneStepShift = out.MaxOneStepShift / (max.TrueMetric)

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
