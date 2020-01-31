package main

import (
	"log"

	"github.com/gocarina/gocsv"
)

type UKBCoding struct {
	Coding  string
	Value   string
	Meaning string
}

func ImportCoding(url string) (map[string]map[string]string, error) {
	log.Printf("Importing codings from %s\n", url)

	fileBytes, err := OpenFileOrURL(url)
	if err != nil {
		return nil, err
	}

	records := []*UKBCoding{}

	if err := gocsv.UnmarshalBytes(fileBytes, &records); err != nil {
		return nil, err
	}

	out := make(map[string]map[string]string) // Coding => map[Value]Meaning
	for _, record := range records {
		rec, exists := out[record.Coding]
		if !exists {
			rec = make(map[string]string)
		}
		rec[record.Value] = record.Meaning
		out[record.Coding] = rec
	}

	return out, nil
}
