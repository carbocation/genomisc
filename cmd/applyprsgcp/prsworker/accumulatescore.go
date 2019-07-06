package main

import (
	"broad/ghgwas/cmd/applyprs"
	"fmt"
	"log"
	"runtime"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/prsparser"
)

func AccumulateScore(bgenPath string, sites []bgen.VariantIndex) ([]applyprs.Sample, error) {
	// At least one site is valid
	log.Println("Site 1:", sites[0])

	jobs := make(chan bgen.VariantIndex)
	results := make(chan []applyprs.Sample)
	errors := make(chan error)
	final := make(chan []applyprs.Sample)

	// Launch workers
	concurrency := 2 * runtime.NumCPU()
	for w := 0; w < concurrency; w++ {
		go Worker(bgenPath, jobs, results, errors)
	}

	// Launch error notifier that tells the coordinator about errors.
	// Not all errors are guaranteed to be returned
	go func(errors <-chan error) {
		for {
			select {
			case err := <-errors:
				log.Println(err)
			}
		}
	}(errors)

	// Launch reducer that accumulates results
	go func(results, final chan []applyprs.Sample) {

		var accumulator []applyprs.Sample

		processed := 0
		for i := 0; i < len(sites); i++ {
			res := <-results

			if len(res) == 0 {
				// We expect this when we skip a variant  because it's
				// multiallelic, etc
				continue
			}

			if processed == 0 {
				accumulator = make([]applyprs.Sample, 0, len(res))
				var v applyprs.Sample
				for j := range res {
					v = res[j]
					v.NIncremented = 1
					accumulator = append(accumulator, v)
				}
			} else {
				if len(res) != len(accumulator) {
					log.Fatalf("Result size %d differed from accumulator size %d", len(res), len(accumulator))
				}
				for j, v := range res {
					accumulator[j].SumScore += v.SumScore
					accumulator[j].NIncremented++
				}
			}

			if processed > 0 && processed%100 == 0 {
				log.Println("Finished", processed, "sites")
			}

			processed++
		}

		final <- accumulator
	}(results, final)

	// Run all jobs
	for _, site := range sites {
		jobs <- site
	}
	close(jobs)

	// Retrieve the accumulated result
	endResult := <-final
	return endResult, nil

}

func Worker(bgPath string, jobs <-chan bgen.VariantIndex, results chan<- []applyprs.Sample, errors chan<- error) {
	b, err := bgen.Open(bgPath)
	if err != nil {
		panic(err.Error())
	}
	defer b.Close()

	for variant := range jobs {
		prs := LookupPRS(variant.Chromosome, variant.Position)
		res, err := ProcessOneVariant(b, variant, prs)
		if err != nil {
			errors <- err
			results <- nil
			continue
		}
		results <- res
	}
}

func ProcessOneVariant(b *bgen.BGEN, vi bgen.VariantIndex, prs *prsparser.PRS) ([]applyprs.Sample, error) {
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

	if (prs.Allele1 == vi.Allele1 && prs.Allele2 == vi.Allele2) ||
		(prs.Allele1 == vi.Allele2 && prs.Allele2 == vi.Allele1) {
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

	results := make([]applyprs.Sample, variant.NSamples, variant.NSamples)

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

	if prs.EffectAllele == v.Alleles[0] {
		return prs.Score * (2.0*sampleProb.Probabilities[0] + sampleProb.Probabilities[1])
	}

	if prs.EffectAllele == v.Alleles[1] {
		return prs.Score * (sampleProb.Probabilities[1] + 2.0*sampleProb.Probabilities[2])
	}

	return 0.0
}
