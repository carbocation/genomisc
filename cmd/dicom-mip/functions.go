package main

import (
	"image"
	"image/color"
	"math"
)

func rescaleMaxBright(img image.Image) image.Image {
	// Find the max value in the image
	var max uint16
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			g := color.Gray16Model.Convert(img.At(x, y))
			if g.(color.Gray16).Y > max {
				max = g.(color.Gray16).Y
			}
		}
	}

	// Rescale the image to the max value
	rescaled := image.NewGray16(img.Bounds())
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			g := color.Gray16Model.Convert(img.At(x, y))
			rescaled.Set(x, y, color.Gray16{Y: uint16(float64(g.(color.Gray16).Y) * float64(math.MaxUint16-1) / float64(max))})
		}
	}

	return rescaled
}
