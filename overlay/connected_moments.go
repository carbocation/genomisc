package overlay

import (
	"fmt"
	"math"
)

type MomentMethod uint8

const (
	MomentMethodConnected MomentMethod = iota
	MomentMethodLabel
)

type CentralMoments struct {
	Bounds struct {
		TopLeft     Coord
		BottomRight Coord
	}
	Area     float64
	Centroid struct {
		X, Y float64
	}
	LongAxisOrientationRadians float64
	LongAxisPixels             float64
	ShortAxisPixels            float64
	Eccentricity               float64
}

func (c *Connected) ComputeMoments(component ConnectedComponent, method MomentMethod) (CentralMoments, error) {
	if c.LabeledConnectedComponents == nil {
		return CentralMoments{}, fmt.Errorf("Please run &Connected.Count before calling &Connected.Eigen")
	}

	// Via https://en.wikipedia.org/wiki/Image_moment

	// Convention:
	// M* = raw moments
	// mu* = central moments

	// Raw Moments

	// MX0Y0 is the sum of all pixels that apply to our component (the area of
	// the component, ignoring non-component parts of our bounding box)
	var MX0Y0 float64

	// The sum of just the Y coordinate of all pixels of our component
	var MX0Y1 float64

	// The sum of just the X coordinate of all pixels of our component
	var MX1Y0 float64

	// X*Y
	var MX1Y1 float64

	// Higher order raw moments
	var MX2Y0 float64
	var MX0Y2 float64

	for yImg := component.Bounds.TopLeft.Y; yImg <= component.Bounds.BottomRight.Y; yImg++ {
		// y := yImg - component.Bounds.TopLeft.Y
		y := yImg
		for xImg := component.Bounds.TopLeft.X; xImg <= component.Bounds.BottomRight.X; xImg++ {
			// x := xImg - component.Bounds.TopLeft.X
			x := xImg

			// x and y now start at 0,0

			// Only add if this pixel belongs to our component of interest
			if method == MomentMethodLabel {
				if c.PixelLabelIDs[yImg][xImg] != component.LabelID {
					continue
				}
			} else {
				if c.PixelConnectedComponentIDs[yImg][xImg] != component.ComponentID {
					continue
				}
			}

			MX0Y0++
			MX0Y1 += float64(y)
			MX1Y0 += float64(x)
			MX1Y1 += float64(x * y)
			MX2Y0 += float64(x * x)
			MX0Y2 += float64(y * y)
		}
	}

	if MX0Y0 == 0 {
		return CentralMoments{}, fmt.Errorf("No pixels relevant to connected component %d were detected between %v and %v", component.ComponentID, component.Bounds.TopLeft, component.Bounds.BottomRight)
	}

	meanX := MX1Y0 / MX0Y0
	meanY := MX0Y1 / MX0Y0

	// First-order central moments
	muX0Y0 := MX0Y0
	// muX0Y1 := 0 // unused
	// muX1Y0 := 0 // unused
	muX1Y1 := MX1Y1 - meanX*MX0Y1
	muX2Y0 := MX2Y0 - meanX*MX1Y0
	muX0Y2 := MX0Y2 - meanY*MX0Y1

	// Second-order central moments
	muPrimeX2Y0 := muX2Y0 / muX0Y0
	muPrimeX0Y2 := muX0Y2 / muX0Y0
	muPrimeX1Y1 := muX1Y1 / muX0Y0

	// Used to construct eigenvalues
	eigenBase := muPrimeX2Y0 + muPrimeX0Y2
	eigenRoot := math.Sqrt(4*math.Pow(muPrimeX1Y1, 2.0) + math.Pow(muPrimeX2Y0-muPrimeX0Y2, 2.0))

	// Eigenvalues & eigenvector

	// See http://raphael.candelier.fr/?blog=Image%20Moments and
	// http://sibgrapi.sid.inpe.br/col/sid.inpe.br/banon/2002/10.23.11.34/doc/35.pdf
	// for controversy round the Eigenvalue constants.

	eigen1 := math.Sqrt(8 * (eigenBase - eigenRoot)) // w, minor elliptical axis
	eigen2 := math.Sqrt(8 * (eigenBase + eigenRoot)) // l, major elliptical axis

	var computedRadians float64
	if muPrimeX2Y0 == muPrimeX0Y2 {
		// No eccentricity by first order moments. Would lead to division by
		// zero, so instead just arbitrarily choose one of the axes.
		computedRadians = 0
	} else {
		computedRadians = 0.5 * math.Atan(2*muPrimeX1Y1/(muPrimeX2Y0-muPrimeX0Y2))
	}

	m := CentralMoments{
		Bounds: component.Bounds,
		Area:   MX0Y0,
		Centroid: struct{ X, Y float64 }{
			X: meanX,
			Y: meanY,
		},
		LongAxisOrientationRadians: computedRadians,
		LongAxisPixels:             eigen2,
		ShortAxisPixels:            eigen1,
		Eccentricity:               math.Sqrt(1 - eigen1/eigen2),
	}

	return m, nil
}

func clampPositiveRadians(radians float64) float64 {
	windowed := math.Mod(radians, 2*math.Pi)
	if windowed < 0 {
		return windowed + 2*math.Pi
	}

	return windowed
}
