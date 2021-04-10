package main

import (
	"sort"
)

type Permutation struct {
	Loci []Locus
}

func (p Permutation) NonMendelianGenesNearLoci(mendelian map[string]Gene, radius float64) int {
	n := 0

	return n
}

func (p Permutation) MendelianGenesNearLoci(mendelian map[string]Gene, radius float64, transcriptStartOnly, faumanMethod bool) int {
	mappedGenes := p.MendelianGeneNamesNearLoci(mendelian, radius, transcriptStartOnly, faumanMethod)

	return len(mappedGenes)
}

func (p Permutation) MendelianGeneNamesNearLoci(mendelian map[string]Gene, radius float64, transcriptStartOnly, faumanMethod bool) map[string]struct{} {
	mappedGenes := make(map[string]struct{})

	// By producing a sorted list, we will produce a stable output. Iterating
	// over a map is unstable because its sort order is random.
	sortedMendelianList := make([]string, 0, len(mendelian))
	for _, gene := range mendelian {
		sortedMendelianList = append(sortedMendelianList, gene.Symbol)
	}
	sort.StringSlice(sortedMendelianList).Sort()

locusLoop:
	for _, locus := range p.Loci {
		for _, geneSymbol := range sortedMendelianList {
			gene := mendelian[geneSymbol]
			if locus.IsGeneWithinRadius(gene, radius, transcriptStartOnly) {
				mappedGenes[gene.Symbol] = struct{}{}

				if faumanMethod {
					// Here, we will not allow more than one gene per locus.
					continue locusLoop
				}
			}
		}
	}

	return mappedGenes
}
