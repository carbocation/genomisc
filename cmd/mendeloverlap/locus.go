package main

import "math"

type Locus struct {
	Index      int
	Chromosome string
	Position   int
}

func (l Locus) IsGeneWithinRadius(gene Gene, radius float64) bool {
	if gene.Chromosome != l.Chromosome {
		return false
	}

	if math.Abs(float64(gene.EarliestTranscriptStart)-float64(l.Position)) < radius*1000 ||
		math.Abs(float64(gene.LatestTranscriptEnd)-float64(l.Position)) < radius*1000 {
		return true
	}

	return false
}
