// sample_file_row	source	score	n_incremented

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	SampleIDCol = iota
	SourceCol
	ScoreCol
	NCol
)

type SampleSource struct {
	SampleID string
	Source   string
}

type ScoreCount struct {
	Score float64
	N     int64
}

func main() {
	var folder, sampleFile string

	flag.StringVar(&folder, "chunks", "", "Path to folder containing PRS chunks")
	flag.StringVar(&sampleFile, "sample", "", "(Optional) Path to .sample file to convert .bgen line numbers into true sample IDs")
	flag.Parse()

	if folder == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[SampleSource]ScoreCount)

	var header []string

	// Accumulate score
	for k, f := range files {
		if f.IsDir() {
			continue
		}

		fi, err := os.Open(filepath.Join(folder, f.Name()))
		if err != nil {
			log.Fatalln(err)
		}

		r := csv.NewReader(fi)
		r.Comma = '\t'

		rows, err := r.ReadAll()
		if err != nil {
			log.Fatalln(err)
		}

		for i, cols := range rows {
			if i == 0 {
				if k == 0 {
					header = cols
				}
				continue
			}

			ss := SampleSource{
				SampleID: cols[SampleIDCol],
				Source:   cols[SourceCol],
			}

			score, err := strconv.ParseFloat(cols[ScoreCol], 64)
			if err != nil {
				log.Fatalln(err)
			}

			count, err := strconv.ParseInt(cols[NCol], 10, 64)
			if err != nil {
				log.Fatalln(err)
			}

			entry := m[ss]
			entry.Score += score
			entry.N += count

			m[ss] = entry
		}
	}

	// Sort by source and then by ID
	sorted := make([]SampleSource, 0, len(m))

	for ss := range m {
		sorted = append(sorted, ss)
	}

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Source != sorted[j].Source {
			return sorted[i].Source < sorted[j].Source
		}

		return sorted[i].SampleID < sorted[j].SampleID
	})

	var samp [][]string
	if sampleFile != "" {
		samp, err = readSampleFile(sampleFile)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// If we passed a sample file, then override the sample header
	if len(samp) > 0 {
		header[SampleIDCol] = "sample_id"
	}

	// Emit header
	fmt.Println(strings.Join(header, "\t"))

	// Emit output
	for _, ss := range sorted {
		entry := m[ss]

		// If we passed a sample file, then override the sample
		sampleID := ss.SampleID
		if len(samp) > 0 {
			fileRow, err := strconv.ParseInt(ss.SampleID, 10, 64)
			if err != nil {
				log.Fatalln(err)
			}

			// .sample files have a two-line header
			sampleID = samp[fileRow+2][0]
		}

		fmt.Printf("%s\t%s\t%.5g\t%d\n", sampleID, ss.Source, entry.Score, entry.N)
	}

}

func readSampleFile(samplePath string) ([][]string, error) {
	// Load the .sample file:
	sf, err := os.Open(samplePath)
	if err != nil {
		return nil, err
	}
	sfCSV := csv.NewReader(sf)
	sfCSV.Comma = ' '
	sampleFileContents, err := sfCSV.ReadAll()
	if err != nil {
		return nil, err
	}

	return sampleFileContents, nil
}
