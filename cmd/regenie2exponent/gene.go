package main

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/carbocation/pfx"
)

var BioMartFilename string = assemblies["38"]

var assemblies = map[string]string{
	"37": "ensembl.grch37.p13.genes",
	"38": "ensembl.grch38.p12.genes",
}

//go:embed lookups/*
var embeddedTemplates embed.FS

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
	ENSG   string
	Symbol string
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

func GeneMapFromGeneSlice(in []Gene, key int) map[string]Gene {
	if key == 0 {
		key = GeneStableID
	}

	out := make(map[string]Gene, len(in))
	for _, gene := range in {
		out[gene.ENSG] = gene
	}

	return out
}

// FetchGenes pulls in all transcripts
func FetchGenes() ([]Gene, error) {
	fileBytes, err := embeddedTemplates.ReadFile(fmt.Sprintf("lookups/%s", BioMartFilename))
	if err != nil {
		return nil, pfx.Err(err)
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

		results = append(results, Gene{
			ENSG:   rec[GeneStableID],
			Symbol: rec[GeneName],
		})
	}

	return results, nil
}
