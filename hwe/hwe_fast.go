package hwe

import "github.com/BenLubar/memoize"

var memoizedExact = memoize.Memoize(Exact)
var memoizedApproximate = memoize.Memoize(Approximate)

// Fast uses the Chi Square approximation. If the P value based on a 1
// dimensional Chi Square test is found to be significant based on your Cutoff,
// then the exact P value is calculated and returned.
func Fast(AA, Aa, aa, Cutoff float64) (p float64) {
	// p = Approximate(AA, Aa, aa)
	p = memoizedApproximate.(func(float64, float64, float64) float64)(AA, Aa, aa)

	if p < Cutoff {
		return memoizedExact.(func(int64, int64, int64) float64)(int64(AA), int64(Aa), int64(aa))
	}

	return p
}
