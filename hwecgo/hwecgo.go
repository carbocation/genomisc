package hwecgo

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "hwe.h"
import "C"

func Exact(AA, Aa, aa int64) float64 {
	return float64(C.SNPHWE2(C.int(Aa), C.int(AA), C.int(aa), C.uint(0)))
}
