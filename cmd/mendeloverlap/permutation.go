package main

type Permutation struct {
	Loci []Locus
}

func (p Permutation) NonMendelianGenesNearLoci(mendelian map[string]Gene, radius float64) int {
	n := 0

	return n
}

func (p Permutation) MendelianGenesNearLoci(mendelian map[string]Gene, radius float64) int {
	n := 0

	for _, locus := range p.Loci {
		for _, gene := range mendelian {
			if locus.IsGeneWithinRadius(gene, radius) {
				n++
			}
		}
	}

	return n
}
