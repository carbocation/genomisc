package main

import (
	"encoding/csv"
	"os"

	"github.com/carbocation/pfx"
)

func ExtractPhenoFileNames(fileList string) ([]string, error) {
	f, err := os.Open(fileList)
	if err != nil {
		return nil, pfx.Err(err)
	}

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		return nil, pfx.Err(err)
	}

	output := make([]string, 0)
	for _, row := range records {
		if len(row) < 1 {
			continue
		}

		// The first column is defined to contain the file path. There is no
		// header.
		output = append(output, row[0])
	}

	return output, nil
}
