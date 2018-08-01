package prsparser

import "strings"

type Layout struct {
	Delimiter       rune
	Comment         rune
	ColEffectAllele int
	ColAllele1      int
	ColAllele2      int
	ColChromosome   int
	ColPosition     int
	ColScore        int
	Parser          *func(layout *Layout, row []string) (PRS, error)
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
		Parser:          &defaultParseRow,
	},
	"LDPRED": {
		Delimiter:       ' ',
		Comment:         '#',
		ColEffectAllele: 3,
		ColAllele1:      3,
		ColAllele2:      4,
		ColChromosome:   0,
		ColPosition:     1,
		ColScore:        6,
		Parser:          &ldpredParseRow,
	},
}

func LayoutNames() string {
	b := strings.Builder{}
	i := 0
	for m := range Layouts {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(m)
		i++
	}

	return b.String()
}
