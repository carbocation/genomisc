package main

import (
	"log"
	"sync"

	"gopkg.in/guregu/null.v3"

	"github.com/brentp/vcfgo"
)

type Work struct {
	Chrom    string
	Pos      uint64
	Ref      string
	Alt      string
	SNP      string
	SampleID string
	Genotype null.Int
}

func worker(variant *vcfgo.Variant, alleleID int, work chan<- Work, concurrencyLimit <-chan struct{}, pool *sync.WaitGroup) {
	if err := variant.Header.ParseSamples(variant); err != nil {
		// if err := variant.Header.ParseSamples(variant); err != nil {
		log.Println("Sample parsing error:", err)
		// panic("hmm")
	}

	// VCF, for an alt like A,C, stores genotypes like 0/1, 0/2. The first
	// person is het for A, the second is het for C. In other words, for the Nth
	// alt allele, the genotype representing that is 1+N.
	currentAltAlleleValue := 1 + alleleID

	var altAlleles int
	var missing bool

	pool.Add(len(variant.Samples))
	for sampleLoc, sample := range variant.Samples {

		altAlleles = 0
		missing = false

		if sample == nil {
			missing = true
		} else if len(sample.GT) < 1 {
			missing = true
		} else {
			for _, gt := range sample.GT {
				if gt == 0 {
					// NTD
				} else if gt == -1 {
					missing = true
				} else if gt == currentAltAlleleValue {
					altAlleles++
				} else {
					// Now we know that you have a non-ref, non-missing genotype
					// that is also not the current alt allele. This implies
					// that you have a different alt allele. However, we will
					// treat this as a ref allele for now. Note that if you are
					// on alt allele 2 and someone is 0/1, they will count once
					// for the ref, but they will not count towards either the
					// alt OR the missing unless you do this.
				}
			}
		}

		w := Work{
			Chrom:    variant.Chrom(),
			Pos:      variant.Pos,
			Ref:      variant.Ref(),
			Alt:      variant.Alt()[alleleID],
			SNP:      variant.Id(),
			SampleID: variant.Header.SampleNames[sampleLoc],
		}

		if !missing {
			w.Genotype = null.IntFrom(int64(altAlleles))
		}

		work <- w
	}

	pool.Done()

	<-concurrencyLimit
}
