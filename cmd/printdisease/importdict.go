package main

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"

	"github.com/carbocation/genomisc/ukbb"
	"github.com/gocarina/gocsv"
)

type UKBField struct {
	Path         string
	Category     int
	FieldID      int
	Field        string
	Participants int
	Items        int
	Stability    string
	ValueType    string
	Units        string
	ItemType     string
	Strata       string
	Sexed        string
	Instances    int
	Array        int
	Coding       NonzeroCoding
	Notes        string
	Link         string
}

type NonzeroCoding struct {
	Coding int
}

func (coding NonzeroCoding) String() string {
	if coding.Coding == 0 {
		return ""
	}

	return strconv.Itoa(coding.Coding)
}

func ImportDictionary(url string) (map[int]UKBField, error) {
	log.Printf("Importing dict from %s\n", url)

	fileBytes, err := OpenFileOrURL(url)
	if err != nil {
		return nil, err
	}

	records := []*UKBField{}

	// Tell gocsv to use tab as the delimiter
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		inCloser := io.NopCloser(in)
		r := csv.NewReader(ukbb.NewCSVQuoteFixReadCloser(inCloser))
		r.Comma = ','
		r.LazyQuotes = true
		return r
	})

	if err := gocsv.UnmarshalBytes(fileBytes, &records); err != nil {
		return nil, err
	}

	out := make(map[int]UKBField)
	for _, record := range records {
		out[record.FieldID] = *record
	}

	return out, nil
}
