package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/carbocation/genomisc/compileinfoprint"
	"github.com/carbocation/genomisc/ukbb"
)

const (
	ExpectedRows = 17
)

func main() {
	var (
		dictPath string
	)

	flag.StringVar(&dictPath, "dict", "https://biobank.ndph.ox.ac.uk/~bbdatan/Data_Dictionary_Showcase.csv", "URL to TSV file with the UKBB data dictionary")
	flag.Parse()

	if dictPath == "" {
		flag.PrintDefaults()
		log.Fatalln()
	}

	if err := ImportDictionary(dictPath); err != nil {
		log.Fatalln(err)
	}
}

func ImportDictionary(url string) error {
	log.Printf("Importing from %s\n", url)

	var br io.ReadCloser
	var err error
	if strings.HasPrefix(url, "http") {
		br, err = bodyReader(url)
	} else {
		br, err = os.Open(url)
	}
	if err != nil {
		return err
	}

	reader := csv.NewReader(ukbb.NewCSVQuoteFixReadCloser(br))
	reader.Comma = ','
	reader.LazyQuotes = true

	header := make([]string, 0)
	j := 0
	for ; ; j++ {
		row, err := reader.Read()
		if err != nil && err == io.EOF {
			log.Println("Exiting on line", j)
			br.Close()
			break
		} else if err != nil {
			buf := bytes.NewBuffer(nil)
			io.Copy(buf, br)
			if strings.Contains(buf.String(), "internal error") {
				log.Println("Dictionary File is not permitted to be downloaded from the UKBB")
				continue
			}
		}

		// Handle the header
		if j == 0 {
			log.Printf("Header (%d elements): %+v\n", len(row), row)
			header = append(header, row...)
			for k, v := range header {
				if v == "Coding" {
					header[k] = "coding_file_id"
					break
				}
			}

			if nCols := len(header); nCols != ExpectedRows {
				return fmt.Errorf("Expected a CSV with %d columns; got one with %d", ExpectedRows, nCols)
			}

			fmt.Println(strings.Join(header, "\t"))

			continue
		}

		// Handle the entries
		if x := len(row); x == ExpectedRows {
			fmt.Println(strings.Join(row, "\t"))
		} else {
			log.Println("Row", j, "had", x, "rows, expecte", ExpectedRows)
		}
	}

	log.Println("Created dictionary file with", j, "entries")

	return nil
}

func bodyReader(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
