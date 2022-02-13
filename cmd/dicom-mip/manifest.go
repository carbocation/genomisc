package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type manifestKey struct {
	SampleID string
	Instance string
}

type manifestEntry struct {
	zip       string
	series    string
	dicom     string
	timepoint float64
	X         float64
	Y         float64
	Z         float64
}

func parseManifest(manifestPath string, doNotSort bool) (map[manifestKey][]manifestEntry, error) {

	out := make(map[manifestKey][]manifestEntry)
	var dicom, timepoint, sampleid, instance, zip, series, x, y, z int = -1, -1, -1, -1, -1, -1, -1, -1, -1

	man, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	cr := csv.NewReader(man)
	cr.Comma = '\t'

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
				// These are not mutually exclusive and so this should not be a
				// series of if/elses or a switch statement.
				if col == DicomColumnName {
					dicom = k
				}
				if col == TimepointColumnName {
					timepoint = k
				}
				if col == SampleIDColumnName {
					sampleid = k
				}
				if col == InstanceColumnName {
					instance = k
				}
				if col == SeriesColumnName {
					series = k
				}
				if col == ZipColumnName {
					zip = k
				}
				if col == NativeXColumnName {
					x = k
				}
				if col == NativeYColumnName {
					y = k
				}
				if col == NativeZColumnName {
					z = k
				}

			}

			if dicom < 0 || (timepoint < 0 && !doNotSort) || sampleid < 0 || instance < 0 || zip < 0 || series < 0 {
				return nil, fmt.Errorf("did not find all columns. Please check dicom_column_name")
			}

			// Note that we are not checking for the existence of the X, Y, and
			// Z columns. If they are not present, we simply will assume each
			// pixel is 1x1x1.

			fmt.Println()
			continue
		}

		if i%100000 == 0 {
			fmt.Printf("\rParsed %d lines from the manifest", i)
		}

		tp := 0.0
		if !doNotSort {
			tp, err = strconv.ParseFloat(cols[timepoint], 64)
			if err != nil {
				return nil, err
			}
		}

		key := manifestKey{SampleID: cols[sampleid], Instance: cols[instance]}
		value := manifestEntry{dicom: cols[dicom], timepoint: tp, series: cols[series], zip: cols[zip]}

		// If x, y, and/or z are provided, then we do expect them to be well
		// behaved floats.
		if x >= 0 {
			value.X, err = strconv.ParseFloat(cols[x], 64)
			if err != nil {
				return nil, err
			}
		}
		if y >= 0 {
			value.Y, err = strconv.ParseFloat(cols[y], 64)
			if err != nil {
				return nil, err
			}
		}
		if z >= 0 {
			value.Z, err = strconv.ParseFloat(cols[z], 64)
			if err != nil {
				return nil, err
			}
		}

		entry := out[key]
		entry = append(entry, value)
		out[key] = entry
	}

	fmt.Printf("\rParsed %d lines from the manifest", i)
	fmt.Println()

	return out, nil
}
