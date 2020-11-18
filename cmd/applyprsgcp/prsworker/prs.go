package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
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

func FirstLastPos(currentVariantScoreLookup map[ChrPos]prsparser.PRS) (firstPos, lastPos uint32) {
	var first uint32 = math.MaxUint32
	var last uint32 = 0

	for v := range currentVariantScoreLookup {
		if v.Position < first {
			first = v.Position
		}

		if v.Position > last {
			last = v.Position
		}
	}

	return first, last
}

// LoadPRSInRange is ***not*** safe for concurrent access from multiple goroutines
func LoadPRSInRange(prsPath, layout, chromosome string, firstLine, lastLine int, alwaysIncrement bool) error {
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
	for i := 0; i <= lastLine; i++ {
		row, err := reader.Read()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if i < firstLine {
			// Don't process anything before our first desired line
			continue
		}

		val, err := parser.ParseRow(row)
		if err != nil {
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

		if p.EffectAllele != p.Allele1 && p.EffectAllele != p.Allele2 {
			return fmt.Errorf("Effect Allele (%v) is neither equal to Allele 1 (%v) nor Allele 2 (%v)", p.EffectAllele, p.Allele1, p.Allele2)
		}

		// Ensure that all scores will be positive. If the effect size is
		// negative, swap the effect and alt alleles and the effect sign.
		if alwaysIncrement && p.Score < 0 {
			p.Score *= -1
			if p.EffectAllele == p.Allele1 {
				p.EffectAllele = p.Allele2
			} else {
				p.EffectAllele = p.Allele1
			}
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
