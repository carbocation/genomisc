package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/carbocation/pfx"
	"github.com/gobuffalo/packr/v2"
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
	PlusStrand              bool
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
	lookups := packr.New("Human Genome Assembly", "./lookups")

	file, err := lookups.Find(BioMartFilename)
	if err != nil {
		return nil, pfx.Err(err)
	}

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

		results = append(results, Gene{Symbol: rec[GeneName], Chromosome: rec[Chromosome], EarliestTranscriptStart: start, LatestTranscriptEnd: end})
	}

	return results, nil
}

// SimplifyTranscripts fetches all transcripts for a named symbol and simplifies
// them to the largest span of transcript start - transcript end, so there is
// just one entry per gene.
func SimplifyTranscripts(genes []Gene) ([]Gene, error) {
	geneNameMap := make(map[string]struct{})
	for _, gene := range genes {
		geneNameMap[gene.Symbol] = struct{}{}
	}

	lookups := packr.New("Human Genome Assembly", "./lookups")

	file, err := lookups.Find(BioMartFilename)
	if err != nil {
		return nil, pfx.Err(err)
	}
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

		if _, exists := geneNameMap[rec[GeneName]]; !exists {
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

		strandInt, err := strconv.Atoi(rec[Strand])
		if err != nil {
			return nil, err
		}

		results = append(results, Gene{
			Symbol:                  rec[GeneName],
			Chromosome:              rec[Chromosome],
			EarliestTranscriptStart: start,
			LatestTranscriptEnd:     end,
			PlusStrand:              strandInt > 0,
		})
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

	if len(keepers) < len(geneNameMap) {
		missing := make([]string, 0)
		for key := range geneNameMap {
			if _, exists := keepers[key]; !exists {
				missing = append(missing, key)
			}
		}
		return nil, fmt.Errorf("ERR1: Your Mendelian set contained %d genes, but we could only map %d of them. Missing: %v", len(geneNameMap), len(keepers), missing)
	}

	output := make([]Gene, 0, len(keepers))
	for _, gene := range keepers {
		output = append(output, gene)
	}

	return output, nil
}
