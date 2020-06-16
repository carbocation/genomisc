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

func parsePixelcountFile(pixelcountFile, imageID, pixels, connectedComponents string) (map[string]File, error) {

	out := make(map[string]File)

	f, err := os.Open(pixelcountFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rc io.ReadCloser = f
	if strings.HasSuffix(pixelcountFile, ".gz") {
		rc, err = gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer rc.Close()
	}

	r := csv.NewReader(rc)
	r.Comma = '\t'

	var colImageID, colPixels, colConnectedComponents int
	for i := 0; ; i++ {
		cols, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if i == 0 {
			colImageID, colPixels, colConnectedComponents, err = readPixelHeader(cols, imageID, pixels, connectedComponents)
			if err != nil {
				return nil, err
			}
			continue
		}

		pixelCount, err := strconv.ParseFloat(cols[colPixels], 64)
		if err != nil {
			return nil, err
		}

		connectedComponentCount, err := strconv.ParseFloat(cols[colConnectedComponents], 64)
		if err != nil {
			return nil, err
		}

		out[cols[colImageID]] = File{
			BadWhy:              []string{"No_covariates"},
			Pixels:              pixelCount,
			ConnectedComponents: connectedComponentCount,
		}
	}

	return out, nil
}

func readPixelHeader(cols []string, imageID, pixels, connectedComponents string) (colImageID, colPixels, colConnectedComponents int, err error) {

	found := 0
	for col, v := range cols {
		switch v {
		case imageID:
			found++
			colImageID = col
		case pixels:
			found++
			colPixels = col
		case connectedComponents:
			found++
			colConnectedComponents = col
		}
	}

	if expected := 3; found != expected {
		return 0, 0, 0, fmt.Errorf("Expected to find %d header columns, but found %d", expected, found)
	}

	return colImageID, colPixels, colConnectedComponents, nil
}
