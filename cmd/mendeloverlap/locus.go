package main

import "math"

type Locus struct {
	Index      int
	Chromosome string
	Position   int
}

func (l Locus) IsGeneWithinRadius(gene Gene, radius float64, transcriptStartOnly bool) bool {
	if gene.Chromosome != l.Chromosome {
		return false
	}

	// Regardless of other options, if the SNP is physicall on a transcript, we
	// will count it:
	if l.Position >= gene.EarliestTranscriptStart && l.Position <= gene.LatestTranscriptEnd {
		return true
	}

	// If you only want to assess based on distance from the transcription start
	// site:
	if transcriptStartOnly {
		transcriptStartSite := gene.EarliestTranscriptStart
		if !gene.PlusStrand {
			transcriptStartSite = gene.LatestTranscriptEnd
		}

		if math.Abs(float64(transcriptStartSite-l.Position)) < radius*1000 {
			return true
		}

		return false
	}

	// Otherwise allow proximity to either the start or the end of the
	// transcript:
	if math.Abs(float64(gene.EarliestTranscriptStart)-float64(l.Position)) < radius*1000 ||
		math.Abs(float64(gene.LatestTranscriptEnd)-float64(l.Position)) < radius*1000 {
		return true
	}

	return false
}
