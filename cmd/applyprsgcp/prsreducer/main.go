// sample_file_row	source	score	n_incremented

package main

import (
	"context"
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

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
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

var (
	client *storage.Client
)

func main() {

	fmt.Fprintf(os.Stderr, "%q\n", os.Args)

	var folder, sampleFile string

	flag.StringVar(&folder, "chunks", "", "Path to folder containing PRS chunks")
	flag.StringVar(&sampleFile, "sample", "", "(Optional) Path to .sample file to convert .bgen line numbers into true sample IDs, if not already done at a previous step.")
	flag.Parse()

	if folder == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Connect to Google Storage if requested
	if strings.HasPrefix(sampleFile, "gs://") {
		var err error
		client, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[SampleSource]ScoreCount)

	var header []string

	// Accumulate score
	firstFile := true
	for _, f := range files {
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
			score, err := strconv.ParseFloat(cols[ScoreCol], 64)

			if i == 0 {
				// Within each file, the first row is the header. However, if
				// the header is not set, then the ScoreCol will be able to be
				// parsed as a number.
				if err != nil {
					// There is a header
					if firstFile {
						// If this is the first file, then save the header
						header = cols
						firstFile = false
					}
					continue
				}
			}

			// For any other line besides the first line, the score not being
			// numeric is an error.
			if err != nil {
				log.Fatalln(err)
			}

			ss := SampleSource{
				SampleID: cols[SampleIDCol],
				Source:   cols[SourceCol],
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

		fi.Close()
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

	// If a header was set, then print it.
	if header != nil {
		// If we passed a sample file, then override the sample header
		if len(samp) > 0 {
			header[SampleIDCol] = "sample_id"
		}

		// Emit header
		fmt.Println(strings.Join(header, "\t"))
	}

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

		fmt.Printf("%s\t%s\t%.8g\t%d\n", sampleID, ss.Source, entry.Score, entry.N)
	}

}

func readSampleFile(samplePath string) ([][]string, error) {
	// Load the .sample file:
	sf, _, err := bulkprocess.MaybeOpenFromGoogleStorage(samplePath, client)
	if err != nil {
		return nil, err
	}
	defer sf.Close()

	sfCSV := csv.NewReader(sf)
	sfCSV.Comma = ' '
	sampleFileContents, err := sfCSV.ReadAll()
	if err != nil {
		return nil, err
	}

	return sampleFileContents, nil
}
