package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func parseCovarFile(out map[string]File, covarFile, sampleID, imageID, timeID, pxHeight, pxWidth string) error {

	f, err := os.Open(covarFile)
	if err != nil {
		return err
	}
	defer f.Close()

	var rc io.ReadCloser = f
	if strings.HasSuffix(covarFile, ".gz") {
		rc, err = gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer rc.Close()
	}

	r := csv.NewReader(rc)
	r.Comma = ','

	var colsSampleID, colImageID, colTimeID, colPxHeight, colPxWidth int
	for i := 0; ; i++ {
		cols, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if i == 0 {
			colsSampleID, colImageID, colTimeID, colPxHeight, colPxWidth, err = readCovarHeader(cols, sampleID, imageID, timeID, pxHeight, pxWidth)
			if err != nil {
				return err
			}
			continue
		}

		timeID, err := strconv.ParseFloat(cols[colTimeID], 64)
		if err != nil {
			return err
		}

		pxHeightF := 1.0
		if pxHeight != "" {
			pxHeightF, err = strconv.ParseFloat(cols[colPxHeight], 64)
			if err != nil {
				return err
			}
		}

		pxWidthF := 1.0
		if pxWidth != "" {
			pxWidthF, err = strconv.ParseFloat(cols[colPxWidth], 64)
			if err != nil {
				return err
			}
		}

		entry, exists := out[cols[colImageID]]

		if !exists {
			// If the covariate entry doesn't have a corresponding pixel count
			// entry, then we will ignore it. Note that an alternative
			// perspective is that this blank entry should in fact be added,
			// because it will show that there is missing pixel data which will
			// appropriately be flagged in downstream analysis. The benefit in
			// excluding it is that we can create one single large covariate
			// entry and use it across all images. So, there are tradeoffs.
			continue
		}

		entry.BadWhy = []string{}
		entry.SampleID = cols[colsSampleID]
		entry.TimeID = timeID
		entry.PxHeight = pxHeightF
		entry.PxWidth = pxWidthF

		out[cols[colImageID]] = entry

	}

	return nil
}

func readCovarHeader(cols []string, sampleID, imageID, timeID, pxHeight, pxWidth string) (colSampleID, colImageID, colTimeID, colPxHeight, colPxWidth int, err error) {

	found := 0
	foundCols := make([]string, 0)
	for col, v := range cols {
		switch v {
		case imageID:
			found++
			colImageID = col
			foundCols = append(foundCols, "imageID")
		case timeID:
			found++
			colTimeID = col
			foundCols = append(foundCols, "colTimeID")
		case pxHeight:
			found++
			colPxHeight = col
			foundCols = append(foundCols, "colPxHeight")
		case pxWidth:
			found++
			colPxWidth = col
			foundCols = append(foundCols, "colPxWidth")
		case sampleID:
			found++
			colSampleID = col
			foundCols = append(foundCols, "colSampleID")
		}
	}

	expected := 5

	if pxWidth == "" {
		expected--
	}

	if pxHeight == "" {
		expected--
	}

	if found != expected {
		return colSampleID, colImageID, colTimeID, colPxHeight, colPxWidth, fmt.Errorf("Expected to find %d header columns, but found %d (%v)", expected, found, foundCols)
	}

	return colSampleID, colImageID, colTimeID, colPxHeight, colPxWidth, nil
}
