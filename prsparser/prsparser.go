package prsparser

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
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

	return n, nil
}

func (prsp *PRSParser) ParseRow(row []string) (PRS, error) {
	if prsp.Layout.Parser == nil {
		return defaultParseRow(prsp.Layout, row)
	}

	return (*prsp.Layout.Parser)(prsp.Layout, row)
}

var defaultParseRow = func(layout *Layout, row []string) (PRS, error) {
	p := PRS{}
	p.EffectAllele = Allele(row[layout.ColEffectAllele])
	p.Allele1 = Allele(row[layout.ColAllele1])
	p.Allele2 = Allele(row[layout.ColAllele2])
	p.Chromosome = row[layout.ColChromosome]

	if pos, err := strconv.Atoi(row[layout.ColPosition]); err != nil {
		return p, err
	} else {
		p.Position = pos
	}

	if score, err := strconv.ParseFloat(row[layout.ColScore], 64); err != nil {
		return p, err
	} else {
		p.Score = score
	}

	return p, nil
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
