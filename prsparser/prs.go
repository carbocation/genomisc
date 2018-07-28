package prsparser

type PRS struct {
	EffectAllele Allele
	Allele1      Allele
	Allele2      Allele
	Chromosome   string
	Position     int
	Score        float64
}
