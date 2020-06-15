package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

func parseCovarFile(out map[string]File, covarFile, sampleID, imageID, timeID, pxHeight, pxWidth string) error {

	f, err := os.Open(covarFile)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
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

		pxHeight, err := strconv.ParseFloat(cols[colPxHeight], 64)
		if err != nil {
			return err
		}

		pxWidth, err := strconv.ParseFloat(cols[colPxWidth], 64)
		if err != nil {
			return err
		}

		entry := out[cols[colImageID]]

		entry.BadWhy = ""
		entry.SampleID = cols[colsSampleID]
		entry.TimeID = timeID
		entry.PxHeight = pxHeight
		entry.PxWidth = pxWidth

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

	if expected := 5; found != expected {
		return colSampleID, colImageID, colTimeID, colPxHeight, colPxWidth, fmt.Errorf("Expected to find %d header columns, but found %d (%v)", expected, found, foundCols)
	}

	return colSampleID, colImageID, colTimeID, colPxHeight, colPxWidth, nil
}
