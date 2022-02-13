// Via https://gist.github.com/shawnsmithdev/09932ba3ac00e74fd6ae
package main

// rationalApproximation finds an integer numerator and denominator that
// approximate x, given the input condition maxDenominator.
func rationalApproximation(x float64, maxDenominator uint) (numerator, denominator uint) {
	a, b, c, d := uint(0), uint(1), uint(1), uint(1)
	for b <= maxDenominator && d <= maxDenominator {
		mediant := float64(a+c) / float64(b+d)
		if x == mediant {
			if b+d <= maxDenominator {
				return a + c, b + d
			} else if d > b {
				return c, d
			} else {
				return a, b
			}
		} else if x > mediant {
			a, b = a+c, b+d
		} else {
			c, d = a+c, b+d
		}
	}

	if b > maxDenominator {
		return c, d
	} else {
		return a, b
	}
}
