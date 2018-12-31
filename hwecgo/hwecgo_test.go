package hwecgo

import (
	"math"
	"testing"
)

type expectations struct {
	AA int64
	Aa int64
	aa int64

	P float64
}

var examples = []expectations{
	{50000, 300000, 50000, 0},
	{50000, 300, 50000, 0},
	{50001, 301, 50001, 0},
	{50002, 302, 50002, 0},
	{50000, 0, 1, 0.000009999900001},
	{500, 0, 500, 1.319669097657e-301},
	{83, 13, 4, 0.010293},
	{50, 57, 14, 0.8422797565708},
	{2, 1, 3, 0.15151515151515},
	{500, 2, 0, 1},
	{500, 0, 4, 1.033376916931e-10},
	{500, 0, 2, 0.000002988038880362},
	{500, 1, 2, 0.0000148807309415},
	{500, 4, 2, 0.0002050449518921},
	{500, 2, 2, 0.00004443531076574},
}

func TestHWECgo(t *testing.T) {
	for _, v := range examples {
		if p, expected := Exact(v.AA, v.Aa, v.aa), v.P; math.Abs(p-expected) > 1e-6 {
			t.Fatalf("\nExact: error with input: %+v\nP: %.12f\nExpected: %.12f\nDiff: %.12f\n", v, p, expected, p-expected)
		}
	}
}
