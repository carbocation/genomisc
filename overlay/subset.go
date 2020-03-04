package overlay

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
)

const (
	WhichPointBottomRight = "br"
	WhichPointTopLeft     = "tl"
)

// SubsetAndRescaleImage takes an input image, crops based on a top left point
// and a bottom right point, dilates, and then rescales the image.
func SubsetAndRescaleImage(baseImg image.Image, topLeftX, topLeftY, bottomRightX, bottomRightY, scale, dilation int) (image.Image, error) {

	// Correct max bounds if left unset (i.e., 0)
	imgBounds := baseImg.Bounds()
	if bottomRightX == 0 {
		bottomRightX = imgBounds.Max.X
	}
	if bottomRightY == 0 {
		bottomRightY = imgBounds.Max.Y
	}

	// Dilate, if necessary
	topLeftX = DilateDimension(topLeftX, imgBounds.Max.X, dilation, WhichPointTopLeft)
	topLeftY = DilateDimension(topLeftY, imgBounds.Max.Y, dilation, WhichPointTopLeft)
	bottomRightX = DilateDimension(bottomRightX, imgBounds.Max.X, dilation, WhichPointBottomRight)
	bottomRightY = DilateDimension(bottomRightY, imgBounds.Max.Y, dilation, WhichPointBottomRight)

	subImg, ok := baseImg.(*image.RGBA)
	if !ok {
		return nil, fmt.Errorf("Type %T is not currently supported", baseImg)
	}

	// First extract the bounded region of interest
	cutImg := subImg.SubImage(image.Rect(topLeftX, topLeftY, bottomRightX, bottomRightY))

	// Now apply the scale
	width := (bottomRightX - topLeftX) * scale
	outImg := imaging.Resize(cutImg, width, 0, imaging.NearestNeighbor)

	return outImg, nil
}

// DilateDimension expands an axis by "dilationFactor" pixels (additive). It
// basically adds or subtracts pixels, while paying attention to not allow the
// lower bound of the image to go below 0.
func DilateDimension(pos, max, dilationFactor int, direction string) int {
	out := pos
	if direction == WhichPointBottomRight {
		out = out + dilationFactor
	} else {
		out = out - dilationFactor
	}

	if out < 0 {
		out = 0
	}
	if out > max {
		out = max
	}

	return out
}
