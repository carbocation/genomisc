package hwecgo

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "hwe.h"
import "C"

// Exact computes an exact Hardy-Weinberg equilibrium P-value. This is only
// intended for biallelic sites. AA is the number of homozygote major
// individuals; aa is the number of homozygote minors; Aa is the number of
// heterozygotes. This variant is a wrapper around Chris Chang's C code. Please
// see the enclosed license as well as the README.md for additional details and
// links to his repo.
func Exact(AA, Aa, aa int64) float64 {
	return float64(C.SNPHWE2(C.int(Aa), C.int(AA), C.int(aa), C.uint(0)))
}
