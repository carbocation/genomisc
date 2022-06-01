package main

import (
	"bytes"
	"embed"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
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

func main() {
	var geneName string

	flag.StringVar(&geneName, "gene", "", "Name of the gene whose GRCH37 transcript's chr:pos you would like to lookup.")
	flag.Parse()

	if geneName == "" {
		flag.PrintDefaults()
		return
	}

	if err := LookupTranscripts(geneName); err != nil {
		log.Fatalln(err)
	}
}

func SymbolMaxWindow(geneName string) error {
	return nil
}

//go:embed lookups/*
var embeddedTemplates embed.FS

// Fetches all transcripts for a named symbol
func LookupTranscripts(geneName string) error {
	BioMartFilename := "ensembl.grch37.p13.genes"
	fileBytes, err := embeddedTemplates.ReadFile(fmt.Sprintf("lookups/%s", BioMartFilename))
	if err != nil {
		return err
	}

	cr := csv.NewReader(bytes.NewReader(fileBytes))
	cr.Comma = '\t'

	results := make([][]string, 0)

	header := make([]string, 0)
	var i int64
	for {
		rec, err := cr.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		i++
		if i == 1 {
			header = append(header, rec...)

			continue
		}

		if len(rec[Chromosome]) > 2 {
			// Longer chromosome names probably represent patches, which we are
			// not equipped to handle.
			continue
		}

		if rec[GeneName] != geneName {
			continue
		}

		strand := "-"
		if rec[Strand] == "1" {
			strand = "+"
		}

		results = append(results, []string{rec[GeneName], rec[Chromosome], rec[TranscriptStartOneBased], rec[TranscriptEndOneBased], rec[TranscriptLengthIncludingUTRAndCDS], strand})
	}

	if len(results) < 1 {
		return fmt.Errorf("No results were found for %s. Were you using a transcript name instead of a gene name?", geneName)
	}

	fmt.Println("Gene\tChromosome\tTranscriptStart\tTranscriptEnd\tTranscriptLength\tStrand")
	for _, result := range results {
		fmt.Println(strings.Join(result, "\t"))
	}

	return nil
}
