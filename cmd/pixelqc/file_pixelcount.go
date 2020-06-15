package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

func parsePixelcountFile(pixelcountFile, imageID string) (map[string]File, error) {

	out := make(map[string]File)

	f, err := os.Open(pixelcountFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'

	var colImageID int
	for i := 0; ; i++ {
		cols, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if i == 0 {
			colImageID, err = readPixelHeader(cols, imageID)
			if err != nil {
				return nil, err
			}
			continue
		}

		out[cols[colImageID]] = File{
			BadWhy: "No_covariates",
		}
	}

	return out, nil
}

func readPixelHeader(cols []string, imageID string) (colImageID int, err error) {

	found := 0
	for col, v := range cols {
		switch v {
		case imageID:
			found++
			colImageID = col
		}
	}

	if expected := 1; found != expected {
		return 0, fmt.Errorf("Expected to find %d header columns, but found %d", expected, found)
	}

	return colImageID, nil
}
