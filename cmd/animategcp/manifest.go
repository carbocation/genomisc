package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
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

	man, _, err := bulkprocess.MaybeOpenFromGoogleStorage(manifestPath, client)
	if err != nil {
		return nil, err
	}
	cr := csv.NewReader(man)
	cr.Comma = '\t'

	out := make(map[manifestKey][]manifestEntry)
	var dicom, timepoint, sampleid, instance int = -1, -1, -1, -1

	i := 0

	for ; ; i++ {

		cols, err := cr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

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

			if dicom < 0 || timepoint < 0 || sampleid < 0 || instance < 0 {
				return nil, fmt.Errorf("did not find all columns. Please check dicom_column_name")
			}

			fmt.Println()
			continue
		}

		if i%100000 == 0 {
			fmt.Printf("\rParsed %d lines from the manifest", i)
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

	fmt.Printf("\rParsed %d lines from the manifest", i)
	fmt.Println()

	return out, nil
}
