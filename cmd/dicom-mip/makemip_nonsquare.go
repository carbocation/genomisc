package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/fogleman/gg"
)

func toGrayScale(img image.Image) image.Image {
	gray := image.NewGray16(img.Bounds())
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			gray.Set(x, y, color.Gray16Model.Convert(img.At(x, y)))
		}
	}
	return gray
}

func savePNG(dc *gg.Context, outName string) error {
	mipImg := toGrayScale(dc.Image())
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, mipImg)
}

func makeOneCoronalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {

	// log.Println(rationalApproximation(dicomData[0].X/dicomData[0].Z, 20))

	// We will be using subpixel boundaries, so we need to make sure we're
	// creating a canvas big enough for all. The canvas height is always the
	// cumulative sum of mm depth for all images. For coronal images, the width
	// of our picture is the maximum X (mm per pixel) times its number of
	// columns.
	canvasHeight := 0.
	canvasWidth := 0.
	for _, im := range dicomEntries {
		canvasHeight += im.Z
		canvasWidth = math.Max(canvasWidth, float64(imgMap[im.dicom].Bounds().Dx())*im.X)
	}

	// If we don't have information about height/width, fall back to the old
	// coronal MIP function.
	if canvasHeight == 0. || canvasWidth == 0. {
		return makeOneCoronalMIPFromImageMap(dicomEntries, imgMap, outName)
	}

	// dc represents the drawing canvas.
	dc := gg.NewContextForImage(image.NewGray16(image.Rect(0, 0, int(math.Floor(canvasWidth)), int(math.Floor(canvasHeight)))))

	// We need positional information. This can either be implicit (assume we
	// start at the top left corner), or explicit (in which case we need to
	// attach positional data). For now, we'll assume implicit.
	outZ := 0.
	for _, dicomData := range dicomEntries {
		currentImg := imgMap[dicomData.dicom].(*image.Gray16)

		// Iterate over all pixels in each column of the original image.
		outX := 0.
		for x := 0; x < currentImg.Bounds().Max.X; x++ {
			var maxIntensityForVector uint16
			var sumIntensityForVector float64
			for y := 0; y < currentImg.Bounds().Max.Y; y++ {
				intensityHere := currentImg.Gray16At(x, y).Y
				sumIntensityForVector += float64(intensityHere)
				if intensityHere > uint16(maxIntensityForVector) {
					maxIntensityForVector = intensityHere
				}
			}

			// After processing each cell in the column, we can draw its pixel

			// Place the rectangle
			dc.DrawRectangle(outX, outZ, outX+dicomData.X, outZ+dicomData.Z)
			intensity := AverageIntensity
			if intensity == AverageIntensity {
				dc.SetColor(color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.Y))})
			} else {
				dc.SetColor(color.Gray16{maxIntensityForVector})
			}
			dc.Fill()

			outX += dicomData.X
		}

		outZ += dicomData.Z
	}

	// Save image to PNG
	return savePNG(dc, outName)
}

func makeOneSagittalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {

	// log.Println(rationalApproximation(dicomData[0].Y/dicomData[0].Z, 20))

	// We will be using subpixel boundaries, so we need to make sure we're
	// creating a canvas big enough for all. The canvas height is always the
	// cumulative sum of mm depth for all images. For sagittal images, the width
	// of our picture is the maximum Y (mm per pixel) times its number of
	// columns.
	canvasHeight := 0.
	canvasWidth := 0.
	for _, im := range dicomEntries {
		canvasHeight += im.Z
		canvasWidth = math.Max(canvasWidth, float64(imgMap[im.dicom].Bounds().Dy())*im.Y)
	}

	// If we don't have information about height/width, fall back to the old
	// coronal MIP function.
	if canvasHeight == 0. || canvasWidth == 0. {
		return makeOneSagittalMIPFromImageMap(dicomEntries, imgMap, outName)
	}

	// dc represents the drawing canvas.
	dc := gg.NewContextForImage(image.NewGray16(image.Rect(0, 0, int(math.Floor(canvasWidth)), int(math.Floor(canvasHeight)))))

	// We need positional information. This can either be implicit (assume we
	// start at the top left corner), or explicit (in which case we need to
	// attach positional data). For now, we'll assume implicit.
	outZ := 0.
	for _, dicomData := range dicomEntries {
		currentImg := imgMap[dicomData.dicom].(*image.Gray16)

		// Iterate over all pixels in each row of the original image.
		outX := 0.
		for y := 0; y < currentImg.Bounds().Max.Y; y++ {
			var maxIntensityForVector uint16
			var sumIntensityForVector float64
			for x := 0; x < currentImg.Bounds().Max.X; x++ {
				intensityHere := currentImg.Gray16At(x, y).Y
				sumIntensityForVector += float64(intensityHere)
				if intensityHere > uint16(maxIntensityForVector) {
					maxIntensityForVector = intensityHere
				}
			}

			// After processing each cell in the vector, we can draw its pixel

			// Place the rectangle
			dc.DrawRectangle(outX, outZ, outX+dicomData.Y, outZ+dicomData.Z)
			intensity := AverageIntensity
			if intensity == AverageIntensity {
				dc.SetColor(color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.X))})
			} else {
				dc.SetColor(color.Gray16{maxIntensityForVector})
			}
			dc.Fill()

			outX += dicomData.Y
		}

		outZ += dicomData.Z
	}

	// Save image to PNG
	return savePNG(dc, outName)
}
