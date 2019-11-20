package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/carbocation/genomisc"
	"github.com/carbocation/genomisc/prsparser"
)

var currentVariantScoreLookup map[ChrPos]prsparser.PRS

type ChrPos struct {
	Chromosome string
	Position   uint32
}

// LoadPRS is ***not*** safe for concurrent access from multiple goroutines
func LoadPRS(prsPath, layout string) error {
	parser, err := prsparser.New(layout)
	if err != nil {
		return fmt.Errorf("CreatePRSParserErr: %s", err.Error())
	}

	// Open PRS file
	f, err := os.Open(prsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fd, err := genomisc.MaybeDecompressReadCloserFromFile(f)
	if err != nil {
		return err
	}
	defer fd.Close()

	reader := csv.NewReader(fd)
	reader.Comma = parser.CSVReaderSettings.Comma
	reader.Comment = parser.CSVReaderSettings.Comment
	reader.TrimLeadingSpace = parser.CSVReaderSettings.TrimLeadingSpace

	currentVariantScoreLookup = nil
	currentVariantScoreLookup = make(map[ChrPos]prsparser.PRS)
	for i := 0; ; i++ {
		row, err := reader.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		val, err := parser.ParseRow(row)
		if err != nil && i == 0 {
			// Permit a header and skip it
			continue
		} else if err != nil {
			return err
		}

		p := prsparser.PRS{
			Chromosome:   val.Chromosome,
			Position:     val.Position,
			EffectAllele: val.EffectAllele,
			Allele1:      val.Allele1,
			Allele2:      val.Allele2,
			Score:        val.Score,
		}

		currentVariantScoreLookup[ChrPos{p.Chromosome, uint32(p.Position)}] = p
	}

	return nil
}

func LookupPRS(chromosome string, position uint32) *prsparser.PRS {
	if prs, exists := currentVariantScoreLookup[ChrPos{chromosome, position}]; exists {
		return &prs
	} else {
		// See if we have missed a leading zero
		chrInt, err := strconv.Atoi(chromosome)
		if err != nil {
			// If it cannot be parsed as an integer, then it was a sex
			// chromosome and it truly didn't match.
			return nil
		}

		// We parsed as an integer. Now recheck without the leading zero to see
		// if we can match.
		if prs, exists := currentVariantScoreLookup[ChrPos{strconv.Itoa(chrInt), position}]; exists {
			return &prs
		}
	}

	return nil
}

func ChromosomalPRS(currentVariantScoreLookup map[ChrPos]prsparser.PRS) map[string][]prsparser.PRS {
	output := make(map[string][]prsparser.PRS)

	for _, v := range currentVariantScoreLookup {
		if _, exists := output[v.Chromosome]; !exists {
			output[v.Chromosome] = make([]prsparser.PRS, 0)

		}

		output[v.Chromosome] = append(output[v.Chromosome], v)
	}

	return output
}
