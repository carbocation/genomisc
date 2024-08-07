package main

import "math"

type assignment struct {
	ID1 uint32
	ID2 uint32
}

type counter struct {
	assignment
	count uint32
}

type countedImage struct {
	dicom           string
	confusionMatrix counter
}

type evalLabel struct {
	Total  int64
	Agreed int64
	Only1  int64
	Only2  int64
}

// PO: observed probability of agreement at a label
func (v evalLabel) PO(total int64) float64 {
	if v.Total == 0 {
		return 1
	}

	// return float64(e.Agreed) / float64(e.Total)
	return float64(total-v.Only1-v.Only2) / float64(total)
}

func (v evalLabel) PE(total int64) float64 {
	pR1Label := float64(v.Agreed+v.Only1) / float64(total)
	pR2Label := float64(v.Agreed+v.Only2) / float64(total)

	// Probability of chance agreement
	return (pR1Label * pR2Label) + ((1 - pR1Label) * (1 - pR2Label))
}

// Kappa is a function of the observed and expected probabilities of agreement
// at a label
func (v evalLabel) Kappa(total int64) float64 {

	if v.PE(total) == 1 {
		return 0
	}

	return (v.PO(total) - v.PE(total)) / (1 - v.PE(total))
}

func (v evalLabel) Dice() float64 {
	denom := float64((2*v.Agreed + v.Only1 + v.Only2))

	if denom == 0 {
		return 0
	}

	return float64(2*v.Agreed) / float64((2*v.Agreed + v.Only1 + v.Only2))
}

func (v evalLabel) Jaccard() float64 {

	return v.Dice() / (2 - v.Dice())
}

// Count agreement ignores whether pixels overlap and simply asks about
// agreement in terms of the total number of pixels annotated as being within a
// class. This returns a fraction scaled from 0 to 1, with the numerator being
// the smaller pixel count and the denominator being the larger pixel count.
func (v evalLabel) CountAgreement() float64 {
	p1 := float64(v.Agreed + v.Only1)
	p2 := float64(v.Agreed + v.Only2)

	denom := math.Max(p1, p2)

	if denom == 0 {
		return 0
	}

	return math.Min(p1, p2) / denom
}

func (v evalLabel) NetSum() int64 {
	return v.Total - v.Agreed - v.Only1 - v.Only2
}
