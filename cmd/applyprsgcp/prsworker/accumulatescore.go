package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/carbocation/bgen"
	"github.com/carbocation/genomisc/applyprsgcp/prsworker"
	"github.com/carbocation/genomisc/prsparser"
)

func AccumulateScore(bgenPath string, sites []bgen.VariantIndex) ([]Sample, error) {
	// At least one site is valid
	log.Println("Site 1:", sites[0])

	jobs := make(chan bgen.VariantIndex)
	results := make(chan []Sample)
	errors := make(chan error)
	final := make(chan []Sample)

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
	go func(results, final chan []Sample) {

		var accumulator []Sample

		processed := 0
		for i := 0; i < len(sites); i++ {
			res := <-results

			if len(res) == 0 {
				// We expect this when we skip a variant  because it's
				// multiallelic, etc
				continue
			}

			if processed == 0 {
				accumulator = make([]Sample, 0, len(res))
				var v Sample
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

func Worker(bgPath string, jobs <-chan bgen.VariantIndex, results chan<- []Sample, errors chan<- error) {
	b, err := bgen.Open(bgPath)
	if err != nil {
		panic(err.Error())
	}
	defer b.Close()

	for variant := range jobs {
		prs := LookupPRS(variant.Chromosome, variant.Position)
		var res []Sample

		// Because of i/o errors on GCP, may need to loop this
		for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {
			res, err = ProcessOneVariant(b, variant, prs)

			if err != nil && loadAttempts == maxLoadAttempts {

				// Ongoing failure at maxLoadAttempts is a terminal error
				log.Fatalln("Worker:", err)

			} else if err != nil && strings.Contains(err.Error(), "input/output error") {

				// If we had an error, often due to an unreliable underlying
				// filesystem, wait for a substantial amount of time before
				// retrying.
				log.Println("Worker: Sleeping 5s to recover from", err.Error(), "attempt", loadAttempts)
				time.Sleep(5 * time.Second)

				// Since the file handle is no longer valid, we need to close it
				// and reopen
				b.Close()
				_, b, err = prsworker.OpenBGIAndBGEN(bgPath, bgPath+".bgi")
				if err != nil {
					log.Fatalln("Worker:", err)
				}
				defer b.Close()

				continue

			} else if err != nil {

				// Not simply an i/o error. Send the error over the channel like
				// usual
				break

			}

			// If loading the data was error-free, no additional attempts are
			// required.
			break
		}

		if err != nil {
			errors <- err
			results <- nil
			continue
		}
		results <- res
	}
}

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
