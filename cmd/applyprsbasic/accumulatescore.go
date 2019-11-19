package main

import (
	"fmt"
	"strings"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/prsparser"
)

func ProcessOneVariant(b *bgen.BGEN, vi bgen.VariantIndex, prs *prsparser.PRS) ([]Sample, error) {
	nonNilErr := ErrorInfo{Message: "", Chromosome: vi.Chromosome, Position: vi.Position}

	if prs == nil || b == nil {
		nonNilErr.Message = "prs was nil"
		return nil, nonNilErr
	}
	if b == nil {
		nonNilErr.Message = "b was nil"
		return nil, nonNilErr
	}

	if vi.NAlleles > 2 {
		nonNilErr.Message = fmt.Sprintf("%s:%d is multiallelic, skipping", vi.Chromosome, vi.Position)
		return nil, nonNilErr
	}

	// Check whether there is allelic match (we assume same strand) between PRS
	// and the genetic data. Do this in case-insensitive fashion.
	if (strings.EqualFold(string(prs.Allele1), string(vi.Allele1)) && strings.EqualFold(string(prs.Allele2), string(vi.Allele2))) ||
		(strings.EqualFold(string(prs.Allele1), string(vi.Allele2)) && strings.EqualFold(string(prs.Allele2), string(vi.Allele1))) {
		// Must be the case
	} else {
		nonNilErr.Message = fmt.Sprintf("At %s:%d, PRS Alleles were %s,%s but variant alleles were %s,%s", vi.Chromosome, vi.Position, prs.Allele1, prs.Allele2, vi.Allele1, vi.Allele2)
		return nil, nonNilErr
	}

	vr := b.NewVariantReader()

	variant := vr.ReadAt(int64(vi.FileStartPosition))
	if err := vr.Error(); err != nil {
		nonNilErr.Message = err.Error()
		return nil, nonNilErr
	}

	results := make([]Sample, variant.NSamples, variant.NSamples)

	for i := 0; i < len(results); i++ {
		results[i].SumScore = ComputeScore(variant.SampleProbabilities[i], variant, prs)
	}

	if len(results) < 1 {
		panic("Results empty")
	}

	return results, nil
}

func ComputeScore(sampleProb bgen.SampleProbability, v *bgen.Variant, prs *prsparser.PRS) float64 {
	if sampleProb.Ploidy != 2 || len(sampleProb.Probabilities) != 3 {
		return 0.0
	}

	if strings.EqualFold(string(prs.EffectAllele), string(v.Alleles[0])) {
		return prs.Score * (2.0*sampleProb.Probabilities[0] + sampleProb.Probabilities[1])
	}

	if strings.EqualFold(string(prs.EffectAllele), string(v.Alleles[1])) {
		return prs.Score * (sampleProb.Probabilities[1] + 2.0*sampleProb.Probabilities[2])
	}

	return 0.0
}
