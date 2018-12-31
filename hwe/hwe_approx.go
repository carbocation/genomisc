package hwe

import (
	"math"

	"github.com/tokenme/probab/dst"
)

func Approximate(AA, Aa, aa float64) (p float64) {
	defer func() { recover() }()

	p = 1.0 - dst.ChiSquareCDF(1)(chiSquare(AA, Aa, aa))

	return
}

// ChiSquare returns a chi square value (1 degree of freedom). This value is the
// difference between observed and expected genotypes based on the observed
// allele frequency distribution. It can help understand if an allele is at
// Hardy-Weinberg equilibrium, but is more often used to determine whether a
// genotyped SNP is likely to be erroneous.
func chiSquare(AA, Aa, aa float64) float64 {

	A := AA*2 + Aa
	a := aa*2 + Aa

	// Exit early if the site is not biallelic for this population. Otherwise
	// you'll get NaN. Chi square of 0 seems appropriate, since the expectation
	// is of course that the full population will be homozygous major. P=1.0 at
	// chi square 0.
	if A == 0 || a == 0 {
		return 0.0
	}

	// Observed N (sample count) may be smaller than the number of samples
	N := float64(AA + Aa + aa)

	AlleleCount := float64(A + a)

	// Allele frequencies depend on number of observed alleles of each type, not
	// the number of samples with those alleles.
	MajorAlleleFrequency := float64(A) / AlleleCount
	MinorAlleleFrequency := float64(a) / AlleleCount

	// Genotype count expectations if alleles were distributed into genotypes
	// according to the Hardy-Weinberg rules. This is the null hypothesis that
	// we are testing. These are the expected homozygous major, heterozygous,
	// and homozygous minor allele frequencies.
	eAA := MajorAlleleFrequency * MajorAlleleFrequency * N
	eAa := 2.0 * MajorAlleleFrequency * MinorAlleleFrequency * N
	eaa := MinorAlleleFrequency * MinorAlleleFrequency * N

	// ChiSquare
	return math.Pow(eAA-float64(AA), 2)/eAA +
		math.Pow(eAa-float64(Aa), 2)/eAa +
		math.Pow(eaa-float64(aa), 2)/eaa
}
