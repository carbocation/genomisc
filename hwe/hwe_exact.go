package hwe

import (
	"math"
	"math/big"

	"github.com/BenLubar/memoize"
)

var memoizedExactFor = memoize.Memoize(exactFor)
var memoizedFactorial = memoize.Memoize(factorial)

// Exact computes an exact Hardy-Weinberg equilibrium P-value, based on the
// Abecasis paper, itself based on RA Fisher's method. Exact is safe to call
// from concurrent goroutines. The resources used to create this were
// http://courses.washington.edu/b516/lectures_2009/HWE_Lecture.pdf slides 21-22
// and https://www.cog-genomics.org/software/stats for sanity checks.
func Exact(AA, Aa, aa int64) (p float64) {
	// Enforce AA common, aa rare
	if aa > AA {
		AA, aa = aa, AA
	}

	// baseP is the probability of the exact base configuration.
	// baseP := exactFor(AA, Aa, aa)
	baseP := memoizedExactFor.(func(int64, int64, int64) float64)(AA, Aa, aa)

	// The P value is the sum of all probabilities at this exact configuration
	// *or more extreme*. See
	// http://courses.washington.edu/b516/lectures_2009/HWE_Lecture.pdf slides
	// 21-22.
	sumP := baseP

	// Start with the most extreme heterozygote situation
	// Aa += 2 * aa
	// AA, aa = AA-aa, 0
	origAA, origAa, origaa := AA, Aa, aa

	// Left tail: Start with the exact number of hets and increase until we're
	// at an extreme.
	for i := 0; ; i, Aa, AA, aa = i+1, Aa+2, AA-1, aa-1 {
		if aa < 0 {
			break
		}

		if i == 0 {
			continue
		}

		// newest := exactFor(AA, Aa, aa)
		newest := memoizedExactFor.(func(int64, int64, int64) float64)(AA, Aa, aa)

		if newest > baseP {
			continue
		}

		if newest <= math.SmallestNonzeroFloat64 {
			break
		}

		sumP += newest
	}

	// Right tail: Start with the exact number of hets and decrease until we're
	// at an extreme.
	AA, Aa, aa = origAA, origAa, origaa
	for i := 0; ; i, Aa, AA, aa = i+1, Aa-2, AA+1, aa+1 {
		if Aa < 0 {
			break
		}

		if i == 0 {
			continue
		}

		// newest := exactFor(AA, Aa, aa)
		newest := memoizedExactFor.(func(int64, int64, int64) float64)(AA, Aa, aa)

		if newest > baseP {
			continue
		}

		if newest <= math.SmallestNonzeroFloat64 {
			break
		}

		sumP += newest
	}

	return sumP
}

// exactFor yields the probability of observing exactly Aa heterozygotes in a
// sample of AA+Aa+aa individuals with Aa+2*aa minor alleles.
func exactFor(AA, Aa, aa int64) (p float64) {
	A := AA*2 + Aa
	a := aa*2 + Aa
	N := AA + Aa + aa

	// big.Int where needed:
	nAa := big.NewInt(Aa)

	// Generate intermediates
	var denom, nexp big.Int

	// Numerator
	nexp.Exp(big.NewInt(2), nAa, nil)
	nexp.Mul(&nexp, memoizedFactorial.(func(int64, int64) *big.Int)(1, A))
	nexp.Mul(&nexp, memoizedFactorial.(func(int64, int64) *big.Int)(1, a))

	// Denominator
	denom.Add(&denom, memoizedFactorial.(func(int64, int64) *big.Int)(N+1, 2*N))
	denom.Mul(&denom, memoizedFactorial.(func(int64, int64) *big.Int)(1, AA))
	denom.Mul(&denom, memoizedFactorial.(func(int64, int64) *big.Int)(1, Aa))
	denom.Mul(&denom, memoizedFactorial.(func(int64, int64) *big.Int)(1, aa))

	// Results
	var ratNum, ratDenom big.Rat
	ratNum.SetInt(&nexp)
	ratDenom.SetInt(&denom)
	final, _ := new(big.Rat).Quo(&ratNum, &ratDenom).Float64()

	return final
}

func factorial(a, b int64) *big.Int {
	return big.NewInt(1).MulRange(a, b)
}
