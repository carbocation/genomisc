package main

import (
	"log"
	"strings"
	"sync"

	"github.com/carbocation/vcfgo"
	"gopkg.in/guregu/null.v3"
)

type Work struct {
	Chrom        string
	Pos          uint64
	Ref          string
	Alt          string
	SNP          string
	SampleID     string
	Genotype     null.Int
	SampleFields []null.String
}

func worker(variant *vcfgo.Variant, alleleID int, work chan<- Work, concurrencyLimit <-chan struct{}, pool *sync.WaitGroup, sampleFields []string) {
	if err := variant.Header.ParseSamples(variant); err != nil {
		if variant == nil {
			log.Printf("Sample parsing error: %v\n", err)
		} else {
			log.Printf("Sample parsing error for variant %+v: %v\n", *variant, err)
		}
	}

	// VCF, for an alt like A,C, stores genotypes like 0/1, 0/2. The first
	// person is het for A, the second is het for C. In other words, for the
	// 0-based Nth alt allele, the genotype representing that is 1+N.
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
					// This sample has a non-ref, non-missing genotype that is
					// also not the current alt allele. This implies that it has
					// two different alt alleles. This person has no "ref"
					// allele and the two observed alleles are two different
					// non-refs at this site. This does happen. Currently, we
					// treat the other alt as a ref allele, which is
					// *inaccurate*. Note that if you are on alt allele 2 and
					// someone is 0/1, they will count once for the ref, but
					// they will not count towards either the alt OR the missing
					// unless you do this.
				}
			}
		}

		if missing {
			// We will record missing variants, unless we have specified which
			// variant types to keep, and that specification didn't include
			// missing variants
			if keepAlt && !keepMissing {
				// Missing, and we aren't keeping everything, and we aren't
				// keeping missing: Ignore. Otherwise, we'll keep it.
				pool.Done()
				continue
			}
		} else {
			// The variant isn't missing. We will keep the variant unless (1)
			// the variant is reference and we specified that we will keep only
			// a subset of variants, or (2) the variant is not ref, and we
			// specified that we will *only* want missing variants.

			if altAlleles < 1 {
				// 1: Variant is reference
				if keepAlt || keepMissing {
					// No alt alleles found, it's not missing: it's reference. If we
					// ask for alts or missing (or both) but find a reference, we
					// don't print it. If we want everything including reference,
					// don't pass the alt / missing flags.
					pool.Done()
					continue
				}
			} else {
				// 2: Variant is non-reference
				if keepMissing && !keepAlt {
					pool.Done()
					continue
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

		// Add the sample fields in order, if requested
		for _, requestedField := range sampleFields {
			if requestedField == "FILTER" {
				// Field is FILTER?
				w.SampleFields = append(w.SampleFields, null.NewString(variant.Filter, true))
			} else if field, ok := sample.Fields[requestedField]; ok {
				// Field is a sample-level field?
				w.SampleFields = append(w.SampleFields, null.NewString(field, true))
			} else if _, err := variant.Info().Get(strings.TrimPrefix(requestedField, "INFO_")); err == nil {
				// Field is at the variant-level (within the INFO field)?
				// Note that you can force a look at the INFO field by prefixing with INFO_

				// TODO: This does not correctly handle multiallelics (which yield an array, from which you should pick the alleleID'th allele)

				// Need to ensure that the INFO_ prefix is removed, if present
				reqField := strings.TrimPrefix(requestedField, "INFO_")

				w.SampleFields = append(w.SampleFields, null.NewString(string(vcfgo.NewInfoByte(variant.Info().Bytes(), nil).SGet(reqField)), true))
			} else {
				w.SampleFields = append(w.SampleFields, null.NewString("", false))
			}
		}

		work <- w
	}

	pool.Done()

	<-concurrencyLimit
}
