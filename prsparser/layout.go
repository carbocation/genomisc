package prsparser

type Layout struct {
	Delimiter       rune
	Comment         rune
	ColEffectAllele int
	ColAllele1      int
	ColAllele2      int
	ColChromosome   int
	ColPosition     int
	ColScore        int
}

var Layouts = map[string]Layout{
	"AVKNG2018": {
		Delimiter:       '\t',
		Comment:         '#',
		ColEffectAllele: 1,
		ColAllele1:      5,
		ColAllele2:      6,
		ColChromosome:   3,
		ColPosition:     4,
		ColScore:        2,
	},
}
