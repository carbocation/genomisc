package cardiaccycle

import (
	"container/ring"
	"sort"
)

// SeriesSample represents a sample-instance-series pair. Most samples have just
// one series (set of images) occurring at once instance (imaging visit).
// However, thousands of samples will have at least a second instance.
// Additionally, many samples have multiple series within one instance. This
// appears to be due to problems with image acquisition in the first series. You
// don't want to accidentally model them jointly.
type SeriesSample struct {
	Identifier string
}

func sampleMapToRing(sampleMap map[SeriesSample]map[uint16]float64) (map[SeriesSample]*List, error) {
	out := make(map[SeriesSample]*List)

	for sampleID, counts := range sampleMap {
		cl := &List{ring.New(len(counts))}

		// Note: For some samples, the instance_number doesn't increase linearly
		// from 1 : X. In one example, the instance_number started at 51 and
		// continued to 100. So, can't assume that if there are 100, they will
		// go 1-100 nicely.
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
