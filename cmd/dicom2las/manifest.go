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
	zip                   string
	series                string
	dicom                 string
	ImagePositionPatientX float64
	ImagePositionPatientY float64
	ImagePositionPatientZ float64
	PixelWidthNativeX     float64
	PixelWidthNativeY     float64
	PixelWidthNativeZ     float64

	Etc map[string]string
}

func parseManifest(manifestPath string) (map[manifestKey][]manifestEntry, error) {

	out := make(map[manifestKey][]manifestEntry)
	var dicom, sampleid, instance, zip, series, ippX, ippY, ippZ, widthX, widthY, widthZ int = -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1

	man, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	cr := csv.NewReader(man)
	cr.Comma = '\t'

	i := 0

	var header []string

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
				if col == ImagePositionPatientXColumn {
					ippX = k
				}
				if col == ImagePositionPatientYColumn {
					ippY = k
				}
				if col == ImagePositionPatientZColumn {
					ippZ = k
				}
				if col == PixelWidthNativeXColumn {
					widthX = k
				}
				if col == PixelWidthNativeYColumn {
					widthY = k
				}
				if col == PixelWidthNativeZColumn {
					widthZ = k
				}

			}

			header = cols

			if dicom < 0 || sampleid < 0 || instance < 0 || zip < 0 || series < 0 {
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

		key := manifestKey{SampleID: cols[sampleid], Instance: cols[instance]}
		value := manifestEntry{dicom: cols[dicom], series: cols[series], zip: cols[zip]}

		// If x, y, and/or z are provided, then we do expect them to be well
		// behaved floats.
		if ippX >= 0 {
			value.ImagePositionPatientX, err = strconv.ParseFloat(cols[ippX], 64)
			if err != nil {
				return nil, err
			}
		}
		if ippY >= 0 {
			value.ImagePositionPatientY, err = strconv.ParseFloat(cols[ippY], 64)
			if err != nil {
				return nil, err
			}
		}
		if ippZ >= 0 {
			value.ImagePositionPatientZ, err = strconv.ParseFloat(cols[ippZ], 64)
			if err != nil {
				return nil, err
			}
		}

		if widthX >= 0 {
			value.PixelWidthNativeX, err = strconv.ParseFloat(cols[widthX], 64)
			if err != nil {
				return nil, err
			}
		}
		if widthY >= 0 {
			value.PixelWidthNativeY, err = strconv.ParseFloat(cols[widthY], 64)
			if err != nil {
				return nil, err
			}
		}
		if widthZ >= 0 {
			value.PixelWidthNativeZ, err = strconv.ParseFloat(cols[widthZ], 64)
			if err != nil {
				return nil, err
			}
		}

		// Store all other columns as key/value pairs.
		for k, v := range cols {
			if k == dicom || k == sampleid || k == instance || k == zip || k == series || k == ippX || k == ippY || k == ippZ {
				continue
			}

			if value.Etc == nil {
				value.Etc = make(map[string]string)
			}
			value.Etc[header[k]] = v
		}

		entry := out[key]
		entry = append(entry, value)
		out[key] = entry
	}

	fmt.Printf("\rParsed %d lines from the manifest", i)
	fmt.Println()

	return out, nil
}
