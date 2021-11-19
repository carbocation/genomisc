package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

type orderedPaletted struct {
	key   int
	image *image.Paletted
}

const (
	MaximumIntensity = iota
	AverageIntensity
)

func makeOneCoronalMIPFromImageMap(dicomNames []string, imgMap map[string]image.Image, outName string) error {

	// If all images are not the same size, make sure we're creating a canvas
	// big enough for all.
	greatestX := 0
	greatestY := len(dicomNames)
	for _, palettedImage := range imgMap {
		if x := palettedImage.Bounds().Max.X; x > greatestX {
			greatestX = x
		}
	}

	mipImg := image.NewGray16(image.Rect(0, 0, greatestX, greatestY))

	for mipY, dicomName := range dicomNames {
		currentImg := imgMap[dicomName].(*image.Gray16)

		for x := 0; x < currentImg.Bounds().Max.X; x++ {
			var maxIntensityForX uint16
			var sumIntensityForX float64

			// what is the brightest pixel at any y-depth for this x?
			for y := 0; y < currentImg.Bounds().Max.Y; y++ {
				intensityHere := currentImg.Gray16At(x, y).Y

				sumIntensityForX += float64(intensityHere)

				if intensityHere > uint16(maxIntensityForX) {
					maxIntensityForX = intensityHere
				}
			}

			intensity := AverageIntensity
			if intensity == AverageIntensity {
				mipImg.Set(x, mipY, color.Gray16{uint16(sumIntensityForX / float64(len(dicomNames)))})
			} else {
				mipImg.Set(x, mipY, color.Gray16{maxIntensityForX})
			}
		}
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	return png.Encode(f, mipImg)
}

func makeOneSagittalMIPFromImageMap(dicomNames []string, imgMap map[string]image.Image, outName string) error {

	// If all images are not the same size, make sure we're creating a canvas
	// big enough for all.
	greatestX := 0
	greatestY := len(dicomNames)
	for _, palettedImage := range imgMap {
		if y := palettedImage.Bounds().Max.Y; y > greatestY {
			greatestX = y
		}
	}

	mipImg := image.NewGray16(image.Rect(0, 0, greatestX, greatestY))

	for mipX, dicomName := range dicomNames {
		currentImg := imgMap[dicomName].(*image.Gray16)

		for y := 0; y < currentImg.Bounds().Max.Y; y++ {
			var maxIntensityForY uint16
			var sumIntensityForY float64

			// what is the brightest pixel at any x-depth for this y?
			for x := 0; x < currentImg.Bounds().Max.X; x++ {
				intensityHere := currentImg.Gray16At(x, y).Y

				sumIntensityForY += float64(intensityHere)

				if intensityHere > uint16(maxIntensityForY) {
					maxIntensityForY = intensityHere
				}
			}

			intensity := AverageIntensity
			if intensity == AverageIntensity {
				mipImg.Set(y, mipX, color.Gray16{uint16(sumIntensityForY / float64(len(dicomNames)))})
			} else {
				mipImg.Set(y, mipX, color.Gray16{maxIntensityForY})
			}
		}
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	return png.Encode(f, mipImg)
}
