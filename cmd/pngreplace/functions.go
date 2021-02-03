package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// Via https://flaviocopes.com/go-list-files/
func scanFolder(dirname string) ([]os.FileInfo, error) {

	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	return files, nil
}

func getDicomSlice(manifest string) ([]string, error) {
	f, err := os.Open(manifest)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = '\t'
	entries, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	dicomFileCol := -1

	// First, identify whether we are extracting multiple images from any zips.
	// If so, it will be more efficient to open the zip one time and extract the
	// desired images, rather than opening/closing the zip for each image
	// (especially if over gcsfuse)
	dicomSlice := make([]string, 0, len(entries)) // []dicom_filename
	for i, row := range entries {
		if i == 0 {
			for j, col := range row {
				if col == "dicom_file" {
					dicomFileCol = j
				}
			}

			continue
		} else if dicomFileCol < 0 {
			return nil, fmt.Errorf("Did not identify dicom_file in the header line of %s", manifest)
		}

		// Append to this zip file's list of individual dicom images to process
		dicomSlice = append(dicomSlice, row[dicomFileCol])
	}

	return dicomSlice, nil
}
