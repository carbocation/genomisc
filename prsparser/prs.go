package prsparser

type PRS struct {
	EffectAllele Allele
	Allele1      Allele
	Allele2      Allele
	Chromosome   string
	Position     int
	Score        float64
	SNP          string
}

// UseSNP is a heuristic that suggests whether the SNP field should be used for
// a lookup instead of the Position.
func (p PRS) UseSNP() bool {
	if p.Position == 0 && len(p.SNP) > 0 {
		return true
	}

	return false
}

// AllowSNP is a heuristic that suggests whether the SNP field should be used
// included in a lookup *even when* the Position is known
func (p PRS) AllowSNP() bool {
	if p.Position != 0 && len(p.SNP) > 0 {
		return true
	}

	return false
}
