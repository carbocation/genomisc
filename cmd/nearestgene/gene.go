package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gobuffalo/packr"
)

const (
	GeneStableID int = iota
	TranscriptStableID
	ProteinStableID
	Chromosome
	GeneStartOneBased
	GeneEndOneBased
	Strand
	TranscriptStartOneBased
	TranscriptEndOneBased
	TranscriptLengthIncludingUTRAndCDS
	GeneName
)

type Gene struct {
	Symbol          string
	Chromosome      string
	TranscriptStart int
	TranscriptEnd   int
}

func ReadSitesFile(fileName string) ([]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cr := csv.NewReader(f)
	cr.Comment = '#'
	lines, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	sites := make([]string, 0)
	for _, cols := range lines {
		if len(cols) < 1 {
			continue
		}
		sites = append(sites, strings.TrimSpace(cols[0]))
	}

	return sites, nil
}

// FetchGenes pulls in all transcripts
func FetchGenes() ([]Gene, error) {
	lookups := packr.NewBox("./lookups")

	file := lookups.Bytes(BioMartFilename)
	buf := bytes.NewBuffer(file)
	cr := csv.NewReader(buf)
	cr.Comma = '\t'
	cr.Comment = '#'

	results := make([]Gene, 0)

	var i int64
	for {
		rec, err := cr.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		i++
		if i == 1 {
			continue
		}

		if len(rec[Chromosome]) > 2 {
			// Longer chromosome names probably represent patches, which we are
			// not equipped to handle.
			continue
		}

		if strings.TrimSpace(rec[ProteinStableID]) == "" {
			// No protein product - not relevant to this program.
			continue
		}

		start, err := strconv.Atoi(rec[TranscriptStartOneBased])
		if err != nil {
			return nil, err
		}

		end, err := strconv.Atoi(rec[TranscriptEndOneBased])
		if err != nil {
			return nil, err
		}

		// Reverse start and end for minus strand
		if rec[Strand] == "-1" {
			start, end = end, start
		}

		results = append(results, Gene{Symbol: rec[GeneName], Chromosome: rec[Chromosome], TranscriptStart: start, TranscriptEnd: end})
	}

	return results, nil
}
