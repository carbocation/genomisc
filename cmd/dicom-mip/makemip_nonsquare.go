package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
)

var CanvasBackgroundColor = color.Black

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

// findCanvasAndOffsets provides information needed to offset pixels correctly
// for coronal and sagittal images.
func findCanvasAndOffsets(dicomEntries []manifestEntry, imgMap map[string]image.Image) (canvasWidthX, canvasWidthY, canvasHeight, minXPixel, minYPixel, minHeightPixel float64) {
	zMin, zMax := math.MaxFloat64, -math.MaxFloat64
	xMin, yMin := math.MaxFloat64, math.MaxFloat64
	for _, im := range dicomEntries {
		xMin = math.Min(xMin, im.ImagePositionPatientX-im.PixelWidthNativeX/2.)

		yMin = math.Min(yMin, im.ImagePositionPatientY-im.PixelWidthNativeY/2.)

		// For Z, we also need the max so we can yield the canvas height.
		zMin = math.Min(zMin, im.ImagePositionPatientZ-im.PixelWidthNativeZ/2.)
		zMax = math.Max(zMax, im.ImagePositionPatientZ+im.PixelWidthNativeZ/2.)

		// If this image defines a wider or taller canvas, we need to adjust the
		// canvas bounds.
		canvasWidthX = math.Max(canvasWidthX, float64((imgMap[im.dicom].Bounds().Dx()))*im.PixelWidthNativeX)
		canvasWidthY = math.Max(canvasWidthY, float64((imgMap[im.dicom].Bounds().Dy()))*im.PixelWidthNativeY)
	}

	return canvasWidthX, canvasWidthY, zMax - zMin, xMin, yMin, zMin
}

func canvasMakeOneCoronalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, intensityMethod, intensitySlice int) (image.Image, error) {
	// We will be using subpixel boundaries, so we need to make sure we're
	// creating a canvas big enough for all. The canvas height is always the
	// cumulative sum of mm depth for all images. For coronal images, the width
	// of our picture is the maximum X (mm per pixel) times its number of
	// columns.
	canvasWidth, canvasDepth, canvasHeight, lateralMin, depthMin, zMin := findCanvasAndOffsets(dicomEntries, imgMap)

	// If we don't have information about height/width, fall back to the old
	// coronal MIP function.
	if canvasHeight == 0. || canvasWidth == 0. {
		return makeOneCoronalMIPFromImageMap(dicomEntries, imgMap)
	}

	// Represent each image with floating-point millimeter coordinates.
	c := canvas.New(canvasWidth, canvasHeight)

	// Create a canvas context used to keep drawing state
	ctx := canvas.NewContext(c)
	// ctx.FillRule = canvas.EvenOdd

	// Set the background color to be white. Alternatively you could leave it to
	// the image type's default background color. But this approach ensure that
	// the rendering will be consistent across platforms.
	drawRectangle(ctx, 0, 0, canvasWidth, canvasHeight)
	stroke(ctx, CanvasBackgroundColor)

	seen := false

	// We need positional information. This can either be implicit (assume we
	// start at the top left corner), or explicit (in which case we need to
	// attach positional data). For now, we'll assume implicit.
	for i, dicomData := range dicomEntries {

		currentImg := imgMap[dicomData.dicom].(*image.Gray16)

		outZ := dicomData.ImagePositionPatientZ - dicomData.PixelWidthNativeZ/2. - zMin
		outNextZ := outZ + dicomData.PixelWidthNativeZ

		anteroPosteriorOffset := dicomData.PixelWidthNativeY/2. - depthMin
		resolvedAnteroPosteriorPosition := dicomData.ImagePositionPatientY - dicomData.PixelWidthNativeY/2. - depthMin

		_ = anteroPosteriorOffset
		_ = resolvedAnteroPosteriorPosition

		// Each image may have its own offsets
		// for x := 0.; x <= c.W; x += dicomData.PixelWidthNativeX {
		for x := 0; x <= currentImg.Bounds().Max.X; x++ {
			// The X position of the current pixel in the output image is a
			// function of the X position of the current pixel in the original
			// image: baseline for the image + X-column we are on * pixel width
			// + an offset against the lowestmost topleft corner of any image in
			// the stack.
			outX := (dicomData.ImagePositionPatientX - dicomData.PixelWidthNativeX/2.) +
				float64(x)*dicomData.PixelWidthNativeX -
				lateralMin

			outNextX := outX + dicomData.PixelWidthNativeX

			var maxIntensityForVector uint16
			var sumIntensityForVector float64

			// The canvasDepth defines the largest (native image) Y depth in any
			// image in any series that we are processing for this view. So,
			// with proper offsets, we can test to be sure that the particular
			// image, with the right offset, is within the range of the canvas.
			for y := 0; y <= int(math.Ceil(canvasDepth)); y++ {
				// for y := 0; y <= currentImg.Bounds().Max.Y; y++ {
				intensityHere := currentImg.Gray16At(x, y).Y

				depthHere := (dicomData.ImagePositionPatientY - dicomData.PixelWidthNativeY/2.) +
					float64(y)*dicomData.PixelWidthNativeY -
					depthMin

				// If this image is not within the plane, skip this pixel.
				// if math.Round(resolvedAnteroPosteriorPosition) < y || math.Round(resolvedAnteroPosteriorPosition) > y+1 {
				if (depthHere < 0 || depthHere > math.Ceil(canvasDepth)) && (i == x) {
					if !seen {
						seen = true
						log.Println("skipping", i, y, depthHere, canvasDepth)
					}
					continue
				}

				if intensityMethod == SliceIntensity {
					// Currently: "If the depth here is exactly the intensity
					// slice, use the value here". The problem is that there are
					// gaps, so that it's possible that none of the depth levels
					// are exactly the intensity slice.

					//
					// TODO: From the perspective of the requested slice
					// (intensitySlice), find the nearest existent slice above
					// or below it. Even better, proportionally average them.

					// if y == intensitySlice {
					if int(math.Round(depthHere)) == intensitySlice {
						maxIntensityForVector = intensityHere
					}
				} else {
					sumIntensityForVector += float64(intensityHere)
					if intensityHere > uint16(maxIntensityForVector) {
						maxIntensityForVector = intensityHere
					}
				}
			}

			// After processing each cell in the column, we can draw its pixel

			// Place the rectangle
			// ctx.DrawPath(outX+lateralOffsets[i], outZ, canvas.Rectangle(dicomData.Y*1.8, dicomData.Z*1.8))
			drawRectangle(ctx, outX-0.5, outZ-0.5, outNextX+0.5, outNextZ+0.5)

			if intensityMethod == AverageIntensity {
				stroke(ctx, color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.Y))})
			} else {
				col := color.RGBA{uint8(255. * float64(maxIntensityForVector) / 65535.), 0, 0, 255}
				switch dicomData.PixelWidthNativeZ {
				case 3:
					col = color.RGBA{0, 0, uint8(255. * float64(maxIntensityForVector) / 65535.), 255}
				case 3.5:
					col = color.RGBA{0, uint8(255. * float64(maxIntensityForVector) / 65535.), 0, 255}
				case 4:
					col = color.RGBA{uint8(255. * float64(maxIntensityForVector) / 65535.), 0, uint8(255. * float64(maxIntensityForVector) / 65535.), 255}
				}
				stroke(ctx, col) //color.Gray16{maxIntensityForVector})
			}

			outX = outNextX

		}

		continue

		dst := rasterizer.Draw(c, canvas.Resolution(1.), canvas.DefaultColorSpace)
		return dst, nil
		return nil, fmt.Errorf("testing")

		// Iterate over all pixels in each column of the original image.
		for x := 0; x <= currentImg.Bounds().Max.X; x++ {
			// The X position of the current pixel in the output image is a
			// function of the X position of the current pixel in the original
			// image: baseline for the image + X-column we are on * pixel width
			// + an offset against the lowestmost topleft corner of any image in
			// the stack.
			outX := (dicomData.ImagePositionPatientX - dicomData.PixelWidthNativeX/2.) +
				float64(x)*dicomData.PixelWidthNativeX -
				lateralMin

			outNextX := outX + dicomData.PixelWidthNativeX

			var maxIntensityForVector uint16
			var sumIntensityForVector float64
			for y := 0; y <= currentImg.Bounds().Max.Y; y++ {
				intensityHere := currentImg.Gray16At(x, y).Y
				if intensityMethod == SliceIntensity {
					if y == intensitySlice {
						maxIntensityForVector = intensityHere
					}
				} else {
					sumIntensityForVector += float64(intensityHere)
					if intensityHere > uint16(maxIntensityForVector) {
						maxIntensityForVector = intensityHere
					}
				}
			}

			// After processing each cell in the column, we can draw its pixel

			// Place the rectangle
			// ctx.DrawPath(outX+lateralOffsets[i], outZ, canvas.Rectangle(dicomData.Y*1.8, dicomData.Z*1.8))
			drawRectangle(ctx, outX-0.5, outZ-0.5, outNextX+0.5, outNextZ+0.5)

			if intensityMethod == AverageIntensity {
				stroke(ctx, color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.Y))})
			} else {
				col := color.RGBA{uint8(255. * float64(maxIntensityForVector) / 65535.), 0, 0, 255}
				switch dicomData.PixelWidthNativeZ {
				case 3:
					col = color.RGBA{0, 0, uint8(255. * float64(maxIntensityForVector) / 65535.), 255}
				case 3.5:
					col = color.RGBA{0, uint8(255. * float64(maxIntensityForVector) / 65535.), 0, 255}
				case 4:
					col = color.RGBA{uint8(255. * float64(maxIntensityForVector) / 65535.), 0, uint8(255. * float64(maxIntensityForVector) / 65535.), 255}
				}
				stroke(ctx, col) //color.Gray16{maxIntensityForVector})
			}

			outX = outNextX
		}

	}

	// canvas.Resolution defines the number of pixels per millimeter.
	dst := rasterizer.Draw(c, canvas.Resolution(1.), canvas.DefaultColorSpace)
	return dst, nil
}

func canvasMakeOneSagittalMIPFromImageMapNonsquare(dicomEntries []manifestEntry, imgMap map[string]image.Image, intensityMethod, intensitySlice int) (image.Image, error) {
	// We will be using subpixel boundaries, so we need to make sure we're
	// creating a canvas big enough for all. The canvas height is always the
	// cumulative sum of mm depth for all images. For coronal images, the width
	// of our picture is the maximum X (mm per pixel) times its number of
	// columns.
	_, canvasWidth, canvasHeight, _, lateralMin, zMin := findCanvasAndOffsets(dicomEntries, imgMap)

	// If we don't have information about height/width, fall back to the old
	// coronal MIP function.
	if canvasHeight == 0. || canvasWidth == 0. {
		return makeOneSagittalMIPFromImageMap(dicomEntries, imgMap)
	}

	// Represent each image with floating-point millimeter coordinates.
	c := canvas.New(canvasWidth, canvasHeight)

	// Create a canvas context used to keep drawing state
	ctx := canvas.NewContext(c)
	// ctx.FillRule = canvas.EvenOdd

	// Set the background color to be white. Alternatively you could leave it to
	// the image type's default background color. But this approach ensure that
	// the rendering will be consistent across platforms.
	drawRectangle(ctx, 0, 0, canvasWidth, canvasHeight)
	stroke(ctx, CanvasBackgroundColor)

	// We need positional information. This can either be implicit (assume we
	// start at the top left corner), or explicit (in which case we need to
	// attach positional data). For now, we'll assume implicit.
	for _, dicomData := range dicomEntries {

		currentImg := imgMap[dicomData.dicom].(*image.Gray16)

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
				if intensityMethod == SliceIntensity {
					if x == intensitySlice {
						maxIntensityForVector = intensityHere
					}
				} else {
					sumIntensityForVector += float64(intensityHere)
					if intensityHere > uint16(maxIntensityForVector) {
						maxIntensityForVector = intensityHere
					}
				}
			}

			// After processing each cell in the column, we can draw its pixel

			// Place the rectangle
			drawRectangle(ctx, outX-0.5, outZ-0.5, outNextX+0.5, outNextZ+0.5)

			if intensityMethod == AverageIntensity {
				stroke(ctx, color.Gray16{uint16(sumIntensityForVector / float64(currentImg.Bounds().Max.X))})
			} else {
				col := color.RGBA{uint8(255. * float64(maxIntensityForVector) / 65535.), 0, 0, 255}
				switch dicomData.PixelWidthNativeZ {
				case 3:
					col = color.RGBA{0, 0, uint8(255. * float64(maxIntensityForVector) / 65535.), 255}
				case 3.5:
					col = color.RGBA{0, uint8(255. * float64(maxIntensityForVector) / 65535.), 0, 255}
				case 4:
					col = color.RGBA{uint8(255. * float64(maxIntensityForVector) / 65535.), 0, uint8(255. * float64(maxIntensityForVector) / 65535.), 255}
				}
				stroke(ctx, col) //color.Gray16{maxIntensityForVector})
			}

			outX = outNextX
		}

	}

	// canvas.Resolution defines the number of pixels per millimeter.
	dst := rasterizer.Draw(c, canvas.Resolution(1.), canvas.DefaultColorSpace)
	return dst, nil
}
