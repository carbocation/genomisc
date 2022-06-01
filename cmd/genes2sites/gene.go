package main

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
	Symbol                  string
	Chromosome              string
	EarliestTranscriptStart int
	LatestTranscriptEnd     int
}

func ReadMendelianGeneFile(fileName string) (map[string]struct{}, error) {
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

	genes := make(map[string]struct{})
	for _, cols := range lines {
		if len(cols) < 1 {
			continue
		}
		genes[strings.TrimSpace(cols[0])] = struct{}{}
	}

	return genes, nil
}

//go:embed lookups/*
var embeddedTemplates embed.FS

// Fetches all transcripts for a named symbol and simplifies them to the largest
// span of transcript start - transcript end, so there is just one entry per
// gene.
func SimplifyTranscripts(geneNames map[string]struct{}) (map[string]Gene, error) {
	fileBytes, err := embeddedTemplates.ReadFile(fmt.Sprintf("lookups/%s", BioMartFilename))
	if err != nil {
		return nil, err
	}

	cr := csv.NewReader(bytes.NewReader(fileBytes))
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

		if _, exists := geneNames[rec[GeneName]]; !exists {
			continue
		}

		start, err := strconv.Atoi(rec[GeneStartOneBased])
		if err != nil {
			return nil, err
		}

		end, err := strconv.Atoi(rec[GeneEndOneBased])
		if err != nil {
			return nil, err
		}

		results = append(results, Gene{Symbol: rec[GeneName], Chromosome: rec[Chromosome], EarliestTranscriptStart: start, LatestTranscriptEnd: end})
	}

	keepers := make(map[string]Gene)
	for _, transcript := range results {
		keeper, exists := keepers[transcript.Symbol]
		if !exists {
			keepers[transcript.Symbol] = transcript
			continue
		}

		if transcript.EarliestTranscriptStart < keeper.EarliestTranscriptStart {
			temp := keepers[transcript.Symbol]
			temp.EarliestTranscriptStart = transcript.EarliestTranscriptStart
			keepers[transcript.Symbol] = temp
		}

		if transcript.LatestTranscriptEnd > keeper.LatestTranscriptEnd {
			temp := keepers[transcript.Symbol]
			temp.LatestTranscriptEnd = transcript.LatestTranscriptEnd
			keepers[transcript.Symbol] = temp
		}
	}

	if len(keepers) < len(geneNames) {
		missing := make([]string, 0)
		for key := range geneNames {
			if _, exists := keepers[key]; !exists {
				missing = append(missing, key)
			}
		}
		return keepers, fmt.Errorf("ERR1: Your Mendelian set contained %d genes, but we could only map %d of them. Missing: %v", len(geneNames), len(keepers), missing)
	}

	return keepers, nil
}
