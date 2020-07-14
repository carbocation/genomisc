package prsparser

import (
	"testing"
)

func TestDefaultLayout(t *testing.T) {
	row := []string{"1:751756:C:T", "C", "1.4113e-06", "1", "751756", "C", "T"}
	parser, err := New("AVKNG2018")
	if err != nil {
		t.Error(err)
	}
	parsedRow, err := parser.ParseRow(row)
	if err != nil {
		t.Error(err)
	}
	if parsedRow.Allele1 != Allele("C") ||
		parsedRow.Allele2 != Allele("T") ||
		parsedRow.Chromosome != "1" ||
		parsedRow.EffectAllele != "C" ||
		parsedRow.Position != 751756 ||
		parsedRow.Score != 1.4113e-06 {
		t.Error("Mismatch")
	}
}

func TestLDPredLayout(t *testing.T) {
	row := []string{"chrom_1", "751756", "1:751756:C:T", "C", "T", "NA", "1.4113e-06"}
	parser, err := New("LDPRED")
	if err != nil {
		t.Error(err)
	}
	parsedRow, err := parser.ParseRow(row)
	if err != nil {
		t.Error(err)
	}
	if parsedRow.Allele1 != Allele("C") ||
		parsedRow.Allele2 != Allele("T") ||
		parsedRow.Chromosome != "1" ||
		parsedRow.EffectAllele != "T" ||
		parsedRow.Position != 751756 ||
		parsedRow.Score != 1.4113e-06 {
		t.Error("Mismatch")
	}
}

func TestSignFlip(t *testing.T) {
	row := []string{"chrom_1", "751756", "1:751756:C:T", "C", "T", "NA", "-1.4113e-06"}
	parser, err := New("LDPRED")
	if err != nil {
		t.Error(err)
	}
	parsedRow, err := parser.ParseRow(row)
	if err != nil {
		t.Error(err)
	}
	if parsedRow.Allele1 != Allele("C") ||
		parsedRow.Allele2 != Allele("T") ||
		parsedRow.Chromosome != "1" ||
		parsedRow.EffectAllele != "C" ||
		parsedRow.Position != 751756 ||
		parsedRow.Score != 1.4113e-06 {
		t.Error("Mismatch")
	}
}
