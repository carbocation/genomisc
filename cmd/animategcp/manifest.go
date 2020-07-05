package main

import (
	"encoding/csv"
	"os"
	"strconv"
)

type manifestKey struct {
	SampleID string
	Instance string
}

type manifestEntry struct {
	dicom     string
	timepoint float64
}

func parseManifest(manifestPath string) (map[manifestKey][]manifestEntry, error) {
	man, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	cr := csv.NewReader(man)
	cr.Comma = '\t'
	manifest, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	out := make(map[manifestKey][]manifestEntry)
	var dicom, timepoint, sampleid, instance int
	for i, cols := range manifest {
		if i == 0 {
			for k, col := range cols {
				switch col {
				case DicomColumnName:
					dicom = k
				case TimepointColumnName:
					timepoint = k
				case SampleIDColumnName:
					sampleid = k
				case InstanceColumnName:
					instance = k
				}
			}
			continue
		}

		tp, err := strconv.ParseFloat(cols[timepoint], 64)
		if err != nil {
			return nil, err
		}

		key := manifestKey{SampleID: cols[sampleid], Instance: cols[instance]}
		value := manifestEntry{dicom: cols[dicom], timepoint: tp}

		entry := out[key]
		entry = append(entry, value)
		out[key] = entry
	}

	return out, nil
}
