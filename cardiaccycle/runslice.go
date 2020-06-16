package cardiaccycle

import (
	"fmt"
)

// RunFromSlices runs the cardiac cycle program using 3 slices of input: the
// sample identifiers, the timing identifiers, and a metric value (e.g., number
// of connected components, or area, etc). The length of the 3 inputs must be
// the same, because the 0th sampleID corresponds to the 0th timeID and 0th
// metric, etc. AdjacentN is the number of adjacent timepoints to look at when
// computing a smoothed average, and discardN is the number of most extreme
// values to discard when computing the smoothed average.
func RunFromSlices(samples []string, timeIDs []uint16, metric []float64, adjacentN, discardN int) ([]Result, error) {
	sampleMap, err := fetchSampleMapFromSlices(samples, timeIDs, metric)
	if err != nil {
		return nil, err
	}

	ringMap, err := sampleMapToRing(sampleMap)
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(ringMap))

	for sampleID, v := range ringMap {
		res, err := v.Extrema(adjacentN, discardN)
		if err != nil {
			return nil, err
		}
		res.Identifier = sampleID.Identifier

		results = append(results, res)
	}

	return results, nil
}

func fetchSampleMapFromSlices(samples []string, timeIDs []uint16, metric []float64) (sampleMap map[SeriesSample]map[uint16]float64, err error) {

	if (len(samples) != len(timeIDs)) || (len(metric) != len(samples)) {
		return nil, fmt.Errorf("All input slices must have the same length")
	}

	sampleMap = make(map[SeriesSample]map[uint16]float64) // map[sample_id] contains => map[instance_number]metric

	for i, sampleID := range samples {

		samp := SeriesSample{
			Identifier: sampleID,
		}

		instanceMap := sampleMap[samp]
		if instanceMap == nil {
			instanceMap = make(map[uint16]float64)
		}

		instanceMap[timeIDs[i]] = metric[i]

		sampleMap[samp] = instanceMap
	}

	return sampleMap, nil
}
