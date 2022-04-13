package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"sync/atomic"

	"github.com/brentp/irelate/interfaces"
	"github.com/carbocation/bix"
	"github.com/carbocation/genomisc"
	"github.com/carbocation/genomisc/prsparser"
	"github.com/carbocation/vcfgo"
)

type TabixLocus struct {
	chrom string
	start int
	end   int
}

func (tl TabixLocus) Chrom() string {
	return tl.chrom
}

func (tl TabixLocus) Start() uint32 {
	return uint32(tl.start)
}

func (tl TabixLocus) End() uint32 {
	return uint32(tl.end)
}

func scoreVCF(whichChunk int, chromosome string, chromosomalSites []prsparser.PRS, vcfTemplatePath, vcfiTemplatePath string) ([]Sample, []PRSFact, error) {
	vcfFile := vcfTemplatePath
	if strings.Contains(vcfTemplatePath, "%") {
		vcfFile = fmt.Sprintf(vcfTemplatePath, chromosome)
	}
	log.Println("Chunk", whichChunk, "File", vcfFile)

	out, err := VCFInitializeSampleList(vcfFile)
	if err != nil {
		return nil, nil, err
	}

	tabixLocus := TabixLocus{
		chrom: chromosome,
		start: math.MaxInt,
		end:   -1,
	}
	for _, prsSite := range chromosomalSites {
		if prsSite.Position <= (tabixLocus.start - 1) {
			tabixLocus.start = prsSite.Position - 1
		}
		if prsSite.Position >= tabixLocus.end {
			tabixLocus.end = prsSite.Position
		}
	}

	// Exit early if we have no sites to iterate over
	if tabixLocus.end == -1 {
		// log.Printf("Chunk %d had no sites to iterate over (%v)\n", whichChunk, tabixLocus)
		return out, nil, nil
	}

	log.Printf("Seeking sites in %v for chunk %d\n", tabixLocus, whichChunk)

	// for _, prsForSNP := range chromosomalSites {
	prsFacts, err := ReadTabixVCF(vcfFile, []TabixLocus{tabixLocus}, &out)
	if err != nil {
		return nil, nil, err
	}
	// }

	return out, prsFacts, nil
}

func VCFInitializeSampleList(vcfFile string) ([]Sample, error) {
	fraw, err := genomisc.MaybeOpenSeekerFromGoogleStorage(vcfFile, client)
	if err != nil {
		return nil, err
	}
	defer fraw.Close()

	var f io.Reader
	f, err = gzip.NewReader(fraw)
	if err != nil {
		fraw.Seek(0, 0)
		f = fraw
	}

	buffRead := bufio.NewReaderSize(f, BufferSize)
	vcfReader, err := vcfgo.NewReader(buffRead, false)
	if err != nil {
		return nil, err
	}

	out := make([]Sample, len(vcfReader.Header.SampleNames))

	for i, sampleName := range vcfReader.Header.SampleNames {
		out[i] = Sample{
			ID:           sampleName,
			FileRow:      i,
			SumScore:     0,
			NIncremented: 0,
		}
	}

	return out, nil
}

func ReadTabixVCF(vcfFile string, loci []TabixLocus, scores *[]Sample) ([]PRSFact, error) {
	tbx, err := bix.NewGCP(vcfFile, client)
	if err != nil {
		return nil, err
	}
	defer tbx.Close()

	prsFacts := make([]PRSFact, 0)

	j := 0
	for _, locus := range loci {
		vals, err := tbx.Query(locus)
		if err != nil {
			return nil, err
		}
		// defer vals.Close()

		for i := 0; ; i++ {
			j++
			v, err := vals.Next()
			if err != nil && err != io.EOF {
				// True error
				return nil, err
			} else if err == io.EOF {
				// Finished all data.
				break
			}

			// Unwrap multiple layers to get to vcfgo.Variant{}

			v2, ok := v.(interfaces.VarWrap)
			if !ok {
				return nil, fmt.Errorf(v.Chrom(), v.Source(), v.End(), "Not valid VarWrap")
			}

			snp, ok := v2.IVariant.(*vcfgo.Variant)
			if !ok {
				return nil, fmt.Errorf(v.Chrom(), v.Source(), v.End(), "Not valid IVariant")
			}

			if err := tbx.VReader.Header.ParseSamples(snp); err != nil {
				return nil, err
			}

			prsFact, err := ProcessOneVariantVCF(snp, scores)
			if err != nil {
				return nil, err
			}
			if prsFact != (PRSFact{}) {
				prsFacts = append(prsFacts, prsFact)
			}
		}
		// vals.Close()
	}
	// log.Printf("Processed %d variants.\n", j)

	return prsFacts, nil
}

func ProcessOneVariantVCF(b *vcfgo.Variant, scores *[]Sample) (PRSFact, error) {
	nonNilErr := ErrorInfo{Message: "", Chromosome: b.Chromosome, Position: uint32(b.Pos)}

	// Lookup this SNP in our PRS map
	prs := LookupPRS(b.Chromosome, uint32(b.Pos), "")
	if prs == nil {
		// This position doesn't have a PRS weight. Not surprising if scanning
		// through a genomic chunk based on a tabix fetch.

		// log.Printf("Skipping %v which has no weight", b)
		return PRSFact{}, nil
	}

	prsFact := PRSFact{*prs, "", "", 0, 0}

	// log.Println("Scoring these:", b.Chromosome, b.Pos, b.Ref(), b.Alt(), prs)
	incremented := 0

	if b == nil {
		nonNilErr.Message = "Variant contents are nil"
		return PRSFact{}, nonNilErr
	} else if scores == nil {
		nonNilErr.Message = "Sample list is nil"
		return PRSFact{}, nonNilErr
	} else if b.Samples == nil {
		nonNilErr.Message = "VCF Sample genotypes are nil"
		return PRSFact{}, nonNilErr
	}

	// Make sure that both the effect and non-effect PRS alleles are observed
	// among the ref and (one or more alt) alleles at this position.
	prsAllele1Numeric := -1
	prsAllele2Numeric := -1
	siteAlleles := append([]string{b.Ref()}, b.Alt()...)
	for alleleNumeric, chrPosAlleleValue := range siteAlleles {
		if prsAllele1Numeric < 0 && strings.EqualFold(chrPosAlleleValue, string(prs.Allele1)) {
			prsAllele1Numeric = alleleNumeric
		}
		if prsAllele2Numeric < 0 && strings.EqualFold(chrPosAlleleValue, string(prs.Allele2)) {
			prsAllele2Numeric = alleleNumeric
		}
	}
	if prsAllele1Numeric < 0 || prsAllele2Numeric < 0 {
		log.Printf("None of the possible ref/alt pairs for %s:%d:%s:%v matched %v", b.Chromosome, b.Pos, b.Ref(), b.Alt(), prs)
		return PRSFact{}, nil
	}

	// Assign the effect and non-effect alleles to the 0 (ref) to 1...NAllele
	// codes used in the genotype.
	effectAlleleNumeric := prsAllele1Numeric
	nonEffectAlleleNumeric := prsAllele2Numeric
	if prs.EffectAllele != prs.Allele1 {
		effectAlleleNumeric, nonEffectAlleleNumeric = nonEffectAlleleNumeric, effectAlleleNumeric
	}

	results := *scores
	for i := range results {
		// Strongly assumes diploid
		scoreAdd, incrementAdd := ComputeScoreVCF(b.Samples[i], effectAlleleNumeric, nonEffectAlleleNumeric, *prs)
		results[i].SumScore += scoreAdd
		results[i].NIncremented += incrementAdd
		incremented += incrementAdd
		prsFact.SiteEA = siteAlleles[effectAlleleNumeric]
		prsFact.SiteNEA = siteAlleles[nonEffectAlleleNumeric]
		prsFact.Scorable += 1
		prsFact.Scored += incrementAdd
	}

	// if incremented == 0 {
	// 	log.Printf("Warning: for %v, no increments\n", b)
	// }

	atomic.AddUint64(&nSitesProcessed, 1)

	return prsFact, nil
}

func ComputeScoreVCF(sampleProb *vcfgo.SampleGenotype, effectAlleleNumeric, nonEffectAlleleNumeric int, prs prsparser.PRS) (float64, int) {
	matchedEffectAlleles := 0.0
	nIncremented := 0

	// Pass 1: are each of the alleles possessed by this person at this position
	// one of the alleles allowed by the PRS?
	for _, gtInteger := range sampleProb.GT {
		// Don't score the site if one of the genotypes for this person is based
		// on an allele that is not part of the PRS.
		if gtInteger != effectAlleleNumeric && gtInteger != nonEffectAlleleNumeric {
			return matchedEffectAlleles, nIncremented
		}
	}

	// This site is valid, so we'll count an increment (even if no effect
	// alleles were present)
	nIncremented += 1

	// Pass 2: all alleles are valid for this score at this position, so just
	// sum up the effect allele count.
	for _, gtInteger := range sampleProb.GT {
		if gtInteger == effectAlleleNumeric {
			matchedEffectAlleles += 1.0
		}
	}

	return prs.Score * float64(matchedEffectAlleles), nIncremented
}
