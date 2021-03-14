package prsparser

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/carbocation/pfx"
)

type PRSParser struct {
	CSVReaderSettings *csv.Reader
	Layout            *Layout
}

func New(layout string) (*PRSParser, error) {
	l, exists := Layouts[layout]
	if !exists {
		return nil, fmt.Errorf("Layout %s is not found. Valid layout names include: %s", layout, LayoutNames())
	}

	return NewWithLayout(&l)
}

func NewWithLayout(layout *Layout) (*PRSParser, error) {
	n := &PRSParser{}
	n.Layout = layout
	n.CSVReaderSettings = &csv.Reader{}
	n.CSVReaderSettings.Comma = layout.Delimiter
	n.CSVReaderSettings.Comment = layout.Comment
	n.CSVReaderSettings.TrimLeadingSpace = true

	return n, nil
}

func (prsp *PRSParser) ParseRow(row []string) (PRS, error) {
	if prsp.Layout.Parser == nil {
		return defaultParseRow(prsp.Layout, row)
	}

	return (*prsp.Layout.Parser)(prsp.Layout, row)
}

func DefaultParseRow(layout *Layout, row []string) (PRS, error) {
	p := PRS{}
	p.EffectAllele = Allele(row[layout.ColEffectAllele])
	p.Allele1 = Allele(row[layout.ColAllele1])
	p.Allele2 = Allele(row[layout.ColAllele2])
	p.Chromosome = row[layout.ColChromosome]

	// If position is set, prefer that over SNP
	if layout.ColPosition >= 0 {
		if pos, err := strconv.Atoi(row[layout.ColPosition]); err != nil {
			return p, pfx.Err(fmt.Errorf("error at ColPosition (%d): %v", layout.ColPosition, err))
		} else {
			p.Position = pos
		}
	} else if layout.ColSNP >= 0 {
		p.SNP = row[layout.ColSNP]
	} else {
		return p, pfx.Err(fmt.Errorf("either a column for genomic position or for SNP ID must be set"))
	}

	if score, err := strconv.ParseFloat(row[layout.ColScore], 64); err != nil {
		return p, pfx.Err(fmt.Errorf("error at ColScore (%d): %v", layout.ColScore, err))
	} else {
		p.Score = score
	}

	if p.EffectAllele != p.Allele1 && p.EffectAllele != p.Allele2 {
		return p, pfx.Err(fmt.Errorf("effect allele %v is neither allele1 (%s) nor allele2 (%s)", p.EffectAllele, p.Allele1, p.Allele2))
	}

	// Align all alleles such that the effect allele is risk-increasing. This
	// requires flipping the effect and non-effect allele.
	if p.Score < 0 {
		p.Score = -1 * p.Score

		if p.EffectAllele == p.Allele1 {
			p.EffectAllele = p.Allele2
		} else {
			p.EffectAllele = p.Allele1
		}
	}

	return p, nil
}

var defaultParseRow = func(layout *Layout, row []string) (PRS, error) {
	return DefaultParseRow(layout, row)
}

var ldpredParseRow = func(layout *Layout, row []string) (PRS, error) {
	// Parse as usual...
	p, err := defaultParseRow(layout, row)
	if err != nil {
		return p, err
	}

	// ... but remove the chrom_ prefix from the chomosome column
	p.Chromosome = strings.TrimPrefix(row[layout.ColChromosome], "chrom_")

	return p, nil
}
