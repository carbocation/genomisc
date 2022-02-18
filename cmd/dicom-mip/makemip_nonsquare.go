package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
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

func savePNG(img image.Image, outName string) error {
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// return png.Encode(f, imaging.Resize(img, img.Bounds().Max.X/2, img.Bounds().Max.Y/2, imaging.Lanczos))
	return png.Encode(f, img)
}

func stroke(c *canvas.Context, col color.Color) {
	c.SetFillColor(col)
	c.SetStrokeColor(col)
	c.SetStrokeWidth(0.0)
	c.FillStroke()
}

func drawRectangle(c *canvas.Context, x0, y0, x1, y1 float64) {
	c.MoveTo(x0, y0)
	c.LineTo(x1, y0)
	c.LineTo(x1, y1)
	c.LineTo(x0, y1)
	c.Close()
}

func canvasMakeOneCoronalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {
	// We will be using subpixel boundaries, so we need to make sure we're
	// creating a canvas big enough for all. The canvas height is always the
	// cumulative sum of mm depth for all images. For coronal images, the width
	// of our picture is the maximum X (mm per pixel) times its number of
	// columns.
	zMin, zMax := math.MaxFloat64, -math.MaxFloat64
	canvasWidth := 0.
	referenceX := 0.
	lateralMin := math.MaxFloat64
	relativeLateralOffsets := make([]float64, len(dicomEntries))
	for i, im := range dicomEntries {
		// We need to know the lateral offset of each image, otherwise there
		// will be bits that jut out unless all images were aligned to the same
		// ImagePositionPatient. Note that ImagePositionPatient is the *center*
		// of the voxel. So it is additionally adjusted by 1/2 of the voxel
		// thickness in the direction of interest so that we get the leftmost
		// edge of the voxel from the perspective of the transformed image.
		//
		// TODO: don't assume that the first image is the reference.
		if i == 0 {
			referenceX = im.ImagePositionPatientX - im.PixelWidthNativeX/2.
			log.Println("")
			log.Println("IPPX, WidthX:", im.ImagePositionPatientX, im.PixelWidthNativeX/2., referenceX)
		}

		// The offset accounts not only for the reference, but also for the
		// width of the pixel of this slice. So, subsequent uses of the X
		// position only need to account for the lateralOffsets and don't need
		// to redundantly account for these two values.
		relativeLateralOffsets[i] = (im.ImagePositionPatientX - im.PixelWidthNativeX/2.) - referenceX
		if relativeLateralOffsets[i] != 0. {
			log.Println("Lateral offset:", relativeLateralOffsets[i])
		}

		// Knowing the minimum and maximum pixels in the depth dimension of all
		// images is useful for determining the canvas height.
		if im.ImagePositionPatientZ-im.PixelWidthNativeZ/2. < zMin {
			zMin = im.ImagePositionPatientZ - im.PixelWidthNativeZ/2.
		}
		if im.ImagePositionPatientZ+im.PixelWidthNativeZ/2. > zMax {
			zMax = im.ImagePositionPatientZ + im.PixelWidthNativeZ/2.
		}

		lateralMin = math.Min(lateralMin, im.ImagePositionPatientX-im.PixelWidthNativeX/2.)
		canvasWidth = math.Max(canvasWidth, float64((imgMap[im.dicom].Bounds().Dx()))*im.PixelWidthNativeX)
	}

	canvasHeight := zMax - zMin

	log.Println(canvasWidth, canvasHeight)

	// If we don't have information about height/width, fall back to the old
	// coronal MIP function.
	if canvasHeight == 0. || canvasWidth == 0. {
		return makeOneCoronalMIPFromImageMap(dicomEntries, imgMap, outName)
	}

	// Represent each image with floating-point millimeter coordinates.
	c := canvas.New(canvasWidth, canvasHeight)

	// Create a canvas context used to keep drawing state
	ctx := canvas.NewContext(c)
	// ctx.FillRule = canvas.EvenOdd

	// By default, canvas puts (0, 0) at the bottom-left corner and the forward
	// direction in Y is upward. These commands change the coordinate system to
	// have (0, 0) at the top left corner, which is similar to most other
	// coordinate systems. ReflectY() inverts up and down so that the forward
	// direction for Y is downward. But (0,0) is still the bottom left corner,
	// so Translate() is then used to shift the (0, 0) point to the top left
	// corner. See
	// https://github.com/tdewolff/canvas/issues/72#issuecomment-772280609
	//
	// ctx.ReflectY()
	// ctx.Translate(0, -c.H)

	// Set the background color to be white. Alternatively you could leave it to
	// the image type's default background color. But this approach ensure that
	// the rendering will be consistent across platforms.
	// ctx.DrawPath(0, 0, canvas.Rectangle(canvasWidth*2, canvasHeight*2))
	drawRectangle(ctx, 0, 0, canvasWidth, canvasHeight)
	stroke(ctx, color.White)

	// We need positional information. This can either be implicit (assume we
	// start at the top left corner), or explicit (in which case we need to
	// attach positional data). For now, we'll assume implicit.
	for i, dicomData := range dicomEntries {

		currentImg := imgMap[dicomData.dicom].(*image.Gray16)

		// TODO: double check the math here.
		outZ := dicomData.ImagePositionPatientZ - dicomData.PixelWidthNativeZ/2. - zMin
		outNextZ := outZ + dicomData.PixelWidthNativeZ

		// Iterate over all pixels in each column of the original image.
		for x := 0; x <= currentImg.Bounds().Max.X; x++ {
			outX := float64(x)*dicomData.PixelWidthNativeX + dicomData.ImagePositionPatientX - dicomData.PixelWidthNativeX/2. + relativeLateralOffsets[i] - lateralMin
			outNextX := outX + dicomData.PixelWidthNativeX

			var maxIntensityForVector uint16
			var sumIntensityForVector float64
			for y := 0; y <= currentImg.Bounds().Max.Y; y++ {
				intensityHere := currentImg.Gray16At(x, y).Y
				sumIntensityForVector += float64(intensityHere)
				if intensityHere > uint16(maxIntensityForVector) {
					maxIntensityForVector = intensityHere
				}
			}

			if i == 0 && x < 10 {
				log.Println("First row, first couple columns:", outX, outNextX)
			}

			// After processing each cell in the column, we can draw its pixel

			// Place the rectangle
			// ctx.DrawPath(outX+lateralOffsets[i], outZ, canvas.Rectangle(dicomData.Y*1.8, dicomData.Z*1.8))
			drawRectangle(ctx, outX-0.5, outZ-0.5, outNextX+0.5, outNextZ+0.5)

			intensity := AverageIntensity
			if intensity == AverageIntensity {
				stroke(ctx, color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.Y))})
			} else {
				stroke(ctx, color.Gray16{maxIntensityForVector})
			}

			// outX += dicomData.PixelWidthNativeX
			outX = outNextX
		}

	}

	// Save image to PNG. canvas.Resolution defines the number of pixels per
	// millimeter.
	dst := rasterizer.Draw(c, canvas.Resolution(1.), canvas.DefaultColorSpace)
	// dst := rasterizer.Draw(c, canvas.Resolution(2.3), canvas.DefaultColorSpace)
	// return savePNG(toGrayScale(dst), outName)
	return savePNG(dst, outName)
}

func canvasMakeOneSagittalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {

	// We will be using subpixel boundaries, so we need to make sure we're
	// creating a canvas big enough for all. The canvas height is always the
	// cumulative sum of mm depth for all images. For coronal images, the width
	// of our picture is the maximum X (mm per pixel) times its number of
	// columns.
	zMin, zMax := math.MaxFloat64, -math.MaxFloat64
	canvasWidth := 0.
	referenceY := 0.
	lateralMin := math.MaxFloat64
	relativeLateralOffsets := make([]float64, len(dicomEntries))
	for i, im := range dicomEntries {
		// We need to know the lateral offset of each image, otherwise there
		// will be bits that jut out unless all images were aligned to the same
		// ImagePositionPatient. Note that ImagePositionPatient is the *center*
		// of the voxel. So it is additionally adjusted by 1/2 of the voxel
		// thickness in the direction of interest so that we get the leftmost
		// edge of the voxel from the perspective of the transformed image.
		//
		// TODO: don't assume that the first image is the reference.
		if i == 0 {
			referenceY = im.ImagePositionPatientY - im.PixelWidthNativeY/2.
			log.Println("")
			log.Println("Sag. IPPX, WidthX:", im.ImagePositionPatientY, im.PixelWidthNativeY/2., referenceY)
		}

		// The offset accounts not only for the reference, but also for the
		// width of the pixel of this slice. So, subsequent uses of the X
		// position only need to account for the lateralOffsets and don't need
		// to redundantly account for these two values.
		relativeLateralOffsets[i] = (im.ImagePositionPatientY - im.PixelWidthNativeY/2.) - referenceY
		if relativeLateralOffsets[i] != 0. {
			log.Println("Sag. Lateral offset:", relativeLateralOffsets[i])
		}

		// Knowing the minimum and maximum pixels in the depth dimension of all
		// images is useful for determining the canvas height.
		if im.ImagePositionPatientZ-im.PixelWidthNativeZ/2. < zMin {
			zMin = im.ImagePositionPatientZ - im.PixelWidthNativeZ/2.
		}
		if im.ImagePositionPatientZ+im.PixelWidthNativeZ/2. > zMax {
			zMax = im.ImagePositionPatientZ + im.PixelWidthNativeZ/2.
		}

		lateralMin = math.Min(lateralMin, im.ImagePositionPatientY-im.PixelWidthNativeY/2.)
		canvasWidth = math.Max(canvasWidth, float64((imgMap[im.dicom].Bounds().Dy()))*im.PixelWidthNativeY)
	}

	canvasHeight := zMax - zMin

	log.Println(canvasWidth, canvasHeight)

	// If we don't have information about height/width, fall back to the old
	// coronal MIP function.
	if canvasHeight == 0. || canvasWidth == 0. {
		return makeOneSagittalMIPFromImageMap(dicomEntries, imgMap, outName)
	}

	// Represent each image with floating-point millimeter coordinates.
	c := canvas.New(canvasWidth, canvasHeight)

	// Create a canvas context used to keep drawing state
	ctx := canvas.NewContext(c)
	// ctx.FillRule = canvas.EvenOdd

	// By default, canvas puts (0, 0) at the bottom-left corner and the forward
	// direction in Y is upward. These commands change the coordinate system to
	// have (0, 0) at the top left corner, which is similar to most other
	// coordinate systems. ReflectY() inverts up and down so that the forward
	// direction for Y is downward. But (0,0) is still the bottom left corner,
	// so Translate() is then used to shift the (0, 0) point to the top left
	// corner. See
	// https://github.com/tdewolff/canvas/issues/72#issuecomment-772280609
	//
	// ctx.ReflectX()
	// ctx.Translate(-c.W, 0)

	// Set the background color to be white. Alternatively you could leave it to
	// the image type's default background color. But this approach ensure that
	// the rendering will be consistent across platforms.
	// ctx.DrawPath(0, 0, canvas.Rectangle(canvasWidth*2, canvasHeight*2))
	drawRectangle(ctx, 0, 0, canvasWidth, canvasHeight)
	stroke(ctx, color.White)

	// We need positional information. This can either be implicit (assume we
	// start at the top left corner), or explicit (in which case we need to
	// attach positional data). For now, we'll assume implicit.
	for i, dicomData := range dicomEntries {

		currentImg := imgMap[dicomData.dicom].(*image.Gray16)

		// TODO: double check the math here.
		outZ := dicomData.ImagePositionPatientZ - dicomData.PixelWidthNativeZ/2. - zMin
		outNextZ := outZ + dicomData.PixelWidthNativeZ

		// Iterate over all pixels in each column of the original image.
		for y := 0; y <= currentImg.Bounds().Max.Y; y++ {
			// The X position of the current pixel in the output image is a
			// function of the Y position of the current pixel in the original
			// image: baseline for the image + Y-column we are on * pixel width
			// + an offset against the lowestmost topleft corner of any image in
			// the stack.
			// outX := relativeLateralOffsets[i] + float64(y)*dicomData.PixelWidthNativeY
			outX := (dicomData.ImagePositionPatientY - dicomData.PixelWidthNativeY/2.) +
				float64(y)*dicomData.PixelWidthNativeY -
				lateralMin

			outNextX := outX + dicomData.PixelWidthNativeY

			var maxIntensityForVector uint16
			var sumIntensityForVector float64
			for x := 0; x <= currentImg.Bounds().Max.X; x++ {
				intensityHere := currentImg.Gray16At(x, y).Y
				sumIntensityForVector += float64(intensityHere)
				if intensityHere > uint16(maxIntensityForVector) {
					maxIntensityForVector = intensityHere
				}
			}

			if i == 0 && y < 10 {
				log.Println("Sag. First row, first couple columns:", outX, outNextX)
			}

			// After processing each cell in the column, we can draw its pixel

			// Place the rectangle
			drawRectangle(ctx, outX-0.5, outZ-0.5, outNextX+0.5, outNextZ+0.5)

			intensity := MaximumIntensity
			if intensity == AverageIntensity {
				stroke(ctx, color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.X))})
			} else {
				stroke(ctx, color.Gray16{maxIntensityForVector})
			}

			// outX += dicomData.PixelWidthNativeX
			outX = outNextX
		}

	}

	// Save image to PNG. canvas.Resolution defines the number of pixels per
	// millimeter.
	dst := rasterizer.Draw(c, canvas.Resolution(1.), canvas.DefaultColorSpace)
	// dst := rasterizer.Draw(c, canvas.Resolution(2.3), canvas.DefaultColorSpace)
	// return savePNG(toGrayScale(dst), outName)
	return savePNG(dst, outName)
}
