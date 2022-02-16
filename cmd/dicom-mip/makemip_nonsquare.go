package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/fogleman/gg"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"golang.org/x/image/vector"
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
	mipImg := toGrayScale(img)
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
	return savePNG(dc.Image(), outName)
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
	return savePNG(dc.Image(), outName)
}

// vectorStrokeRasterizer causes the rasterizer path to be drawn onto the destination
// image.
func vectorStrokeRasterizer(r *vector.Rasterizer, dst draw.Image, c *image.Uniform) {
	// log.Println("Stroking rasterizer")
	r.Draw(dst, dst.Bounds(), c, image.Point{})

	r.Reset(r.Size().X, r.Size().Y)
	// log.Println("Done stroking rasterizer")
}

func vectorDrawRectangle64(r *vector.Rasterizer, x0, y0, x1, y1 float64) {
	vectorDrawRectangle(r, float32(x0), float32(y0), float32(x1), float32(y1))
}

func vectorDrawRectangle(r *vector.Rasterizer, x0, y0, x1, y1 float32) {
	// log.Println("Drawing rectangle:", x0, y0, x1, y1)
	// r.DrawOp = draw.Src
	r.MoveTo(x0, y0)
	r.LineTo(x1, y0)
	r.LineTo(x1, y1)
	r.LineTo(x0, y1)
	// r.LineTo(x0, y0) // may not be needed
	r.ClosePath()
	// log.Println("Done drawing rectangle")
}

func vectorMakeOneSagittalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {

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

	// dst := draw.Image(image.NewGray16(image.Rect(0, 0, int(math.Floor(canvasWidth)), int(math.Floor(canvasHeight)))))
	dst := draw.Image(image.NewRGBA(image.Rect(0, 0, int(math.Floor(canvasWidth)), int(math.Floor(canvasHeight)))))
	r := vector.NewRasterizer(dst.Bounds().Dx(), dst.Bounds().Dy())
	ui := image.NewUniform(color.White)

	// We need positional information. This can either be implicit (assume we
	// start at the top left corner), or explicit (in which case we need to
	// attach positional data). For now, we'll assume implicit.
	outZ := 0.
	for i, dicomData := range dicomEntries {
		// log.Println(dicomData.dicom)
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
			vectorDrawRectangle64(r, outX, outZ, outX+dicomData.Y, outZ+dicomData.Z)

			intensity := AverageIntensity
			if intensity == AverageIntensity {
				ui.C = color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.X))}
				vectorStrokeRasterizer(r, dst, ui)
			} else {
				ui.C = color.Gray16{maxIntensityForVector}
				vectorStrokeRasterizer(r, dst, ui)
			}

			outX += dicomData.Y
		}

		outZ += dicomData.Z

		if i > 10 {
			break
		}
	}

	// Save image to PNG
	return savePNG(dst, outName)
}

func stroke(c *canvas.Context, col color.Color) {
	c.SetFillColor(col)
	// c.SetStrokeColor(col)
	// c.SetStrokeWidth(0)
	c.Fill()
	// c.ResetView()
}

func drawRectangle(c *canvas.Context, x0, y0, x1, y1 float64) {
	c.MoveTo(x0, y0)
	c.LineTo(x1, y0)
	c.LineTo(x1, y1)
	c.LineTo(x0, y1)
	c.Close()
}

func canvasMakeOneCoronalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {

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

	// Represent each image with floating-point millimeter coordinates.
	c := canvas.New(canvasWidth, canvasHeight)

	// Create a canvas context used to keep drawing state
	ctx := canvas.NewContext(c)

	// By default, canvas puts (0, 0) at the bottom-left corner and the forward
	// direction in Y is upward. These commands change the coordinate system to
	// have (0, 0) at the top left corner, which is similar to most other
	// coordinate systems. ReflectY() inverts up and down so that the forward
	// direction for Y is downward. But (0,0) is still the bottom left corner,
	// so Translate() is then used to shift the (0, 0) point to the top left
	// corner. See
	// https://github.com/tdewolff/canvas/issues/72#issuecomment-772280609
	ctx.ReflectY()
	ctx.Translate(0, -c.H)
	// ctx.SetCoordSystem(canvas.CartesianIII)
	// ctx.SetView(canvas.Identity.ReflectY())

	// Set the background color to be white. Alternatively you could leave it to
	// the image type's default background color. But this approach ensure that
	// the rendering will be consistent across platforms.
	ctx.DrawPath(0, 0, canvas.Rectangle(canvasWidth*2, canvasHeight*2))
	// drawRectangle(ctx, 0, 0, canvasWidth, canvasHeight)
	stroke(ctx, color.White)

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
			ctx.DrawPath(outX, outZ, canvas.Rectangle(dicomData.Y*1.8, dicomData.Z*1.8))
			// drawRectangle(ctx, outX, outZ, outX+dicomData.Y, outZ+dicomData.Z)

			intensity := AverageIntensity
			if intensity == AverageIntensity {
				stroke(ctx, color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.X))})
			} else {
				stroke(ctx, color.Gray16{maxIntensityForVector})
			}

			outX += dicomData.X
		}

		outZ += dicomData.Z
	}

	// Save image to PNG. canvas.Resolution defines the number of pixels per
	// millimeter.
	dst := rasterizer.Draw(c, canvas.Resolution(1.0), canvas.DefaultColorSpace)
	// dst := rasterizer.Draw(c, canvas.Resolution(2.3), canvas.DefaultColorSpace)
	// return savePNG(toGrayScale(dst), outName)
	return savePNG(dst, outName)
}

func canvasMakeOneSagittalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, outName string) error {

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

	// Represent each image with floating-point millimeter coordinates.
	c := canvas.New(canvasWidth, canvasHeight)

	// Create a canvas context used to keep drawing state
	ctx := canvas.NewContext(c)

	// By default, canvas puts (0, 0) at the bottom-left corner and the forward
	// direction in Y is upward. These commands change the coordinate system to
	// have (0, 0) at the top left corner, which is similar to most other
	// coordinate systems. ReflectY() inverts up and down so that the forward
	// direction for Y is downward. But (0,0) is still the bottom left corner,
	// so Translate() is then used to shift the (0, 0) point to the top left
	// corner. See
	// https://github.com/tdewolff/canvas/issues/72#issuecomment-772280609
	ctx.ReflectY()
	ctx.Translate(0, -c.H)
	// ctx.SetCoordSystem(canvas.CartesianIII)
	// ctx.SetView(canvas.Identity.ReflectY())

	// Set the background color to be white. Alternatively you could leave it to
	// the image type's default background color. But this approach ensure that
	// the rendering will be consistent across platforms.
	ctx.DrawPath(0, 0, canvas.Rectangle(canvasWidth*2, canvasHeight*2))
	// drawRectangle(ctx, 0, 0, canvasWidth, canvasHeight)
	stroke(ctx, color.White)

	// We need positional information. This can either be implicit (assume we
	// start at the top left corner), or explicit (in which case we need to
	// attach positional data). For now, we'll assume implicit.
	outZ := 0.
	for _, dicomData := range dicomEntries {
		// log.Println(dicomData.dicom)
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
			ctx.DrawPath(outX, outZ, canvas.Rectangle(dicomData.Y*1.8, dicomData.Z*1.8))
			// drawRectangle(ctx, outX, outZ, outX+dicomData.Y, outZ+dicomData.Z)

			intensity := AverageIntensity
			if intensity == AverageIntensity {
				stroke(ctx, color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.X))})
			} else {
				stroke(ctx, color.Gray16{maxIntensityForVector})
			}

			outX += dicomData.Y
		}

		outZ += dicomData.Z
	}

	// Save image to PNG. canvas.Resolution defines the number of pixels per
	// millimeter.
	dst := rasterizer.Draw(c, canvas.Resolution(1.0), canvas.DefaultColorSpace)
	// dst := rasterizer.Draw(c, canvas.Resolution(2.3), canvas.DefaultColorSpace)
	// return savePNG(toGrayScale(dst), outName)
	return savePNG(dst, outName)
}
