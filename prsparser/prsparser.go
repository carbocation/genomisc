package prsparser

import (
	"encoding/csv"
	"fmt"
	"strconv"
)

type PRSParser struct {
	CSVReaderSettings *csv.Reader
	Layout            Layout
}

func New(layout string) (*PRSParser, error) {
	l, exists := Layouts[layout]
	if !exists {
		return nil, fmt.Errorf("Layout %s is not found. Valid layout names include: %s", layout, LayoutNames())
	}

	return NewWithLayout(l)
}

func NewWithLayout(layout Layout) (*PRSParser, error) {
	n := &PRSParser{}
	n.Layout = layout
	n.CSVReaderSettings = &csv.Reader{}
	n.CSVReaderSettings.Comma = layout.Delimiter
	n.CSVReaderSettings.Comment = layout.Comment

	return n, nil
}

func (prsp *PRSParser) ParseRow(row []string) (PRS, error) {
	p := PRS{}
	p.EffectAllele = Allele(row[prsp.Layout.ColEffectAllele])
	p.Allele1 = Allele(row[prsp.Layout.ColAllele1])
	p.Allele2 = Allele(row[prsp.Layout.ColAllele2])
	p.Chromosome = row[prsp.Layout.ColChromosome]

	if pos, err := strconv.Atoi(row[prsp.Layout.ColPosition]); err != nil {
		return p, err
	} else {
		p.Position = pos
	}

	if score, err := strconv.ParseFloat(row[prsp.Layout.ColScore], 64); err != nil {
		return p, err
	} else {
		p.Score = score
	}

	return p, nil
}
