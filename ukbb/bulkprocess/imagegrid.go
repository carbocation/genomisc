package bulkprocess

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strconv"
)

// ImageGrid is a hack that converts a SAX stack (structured as all 50
// timepoints for series 1, all 50 timepoints for series 2, etc) into a grid
// that simultaneously shows each series side-by-side. It highlights the active
// series based on a seriesAlignment slice that is paired 1:1 with the
// dicomNames slice. A more elegant approach would look at the series names
// instead of making assumptions.
func ImageGrid(dicomNames []string, imageMap map[string]image.Image, series string, seriesAlignment []string, imagesInCardiacCycle, Ncols int) ([]string, map[string]image.Image, error) {
	Nseries := len(imageMap) / imagesInCardiacCycle
	Nrows := Nseries / Ncols
	if Nseries%Ncols != 0 {
		Nrows++
	}
	maxWidth := -1
	maxHeight := -1
	for imgName, img := range imageMap {
		if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
			return nil, nil, fmt.Errorf("Image %s has a height or width of 0", imgName)
		}

		if x := img.Bounds().Dx(); x > maxWidth {
			maxWidth = x
		}
		if y := img.Bounds().Dy(); y > maxHeight {
			maxHeight = y
		}
	}

	newDicomNames := make([]string, 0, imagesInCardiacCycle)
	newImageMap := make(map[string]image.Image)

	// For each part of the cardiac cycle, draw our images, for a total of
	// imagesInCardiacCycle images
	for i := 0; i < imagesInCardiacCycle; i++ {

		r := image.Rect(0, 0, Ncols*maxWidth, Nrows*maxHeight)
		thisImg := image.NewRGBA(r)
		// Set a black background
		draw.Draw(thisImg, thisImg.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

		seriesCounter := 0

		// We know that we need to draw Nseries images onto this template
	ImageGridLoop:
		for row := 0; row < Nrows; row++ {
			for col := 0; col < Ncols; col++ {

				// Account for jagged series counts
				seriesCounter++
				if seriesCounter > Nseries {
					break ImageGridLoop
				}

				imageID := i + imagesInCardiacCycle*(row*Ncols+col)
				if imageID >= len(dicomNames) || imageID >= len(imageMap) {
					break
				}

				pane, exists := imageMap[dicomNames[imageID]]
				if !exists {
					break
				}

				startX := col * maxWidth
				startY := row * maxHeight

				drawRect := image.Rect(startX, startY, startX+pane.Bounds().Dx(), startY+pane.Bounds().Dy())

				// Draw the pane into the master image for this timepoint at its designated spot
				draw.Draw(thisImg, drawRect, pane, image.ZP, draw.Src)

				// Highlight the active pane
				if len(seriesAlignment) > imageID && seriesAlignment[imageID] == series {
					borderColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}

					// Top
					innerRect := drawRect
					innerRect.Max.Y = innerRect.Min.Y + 2
					draw.Draw(thisImg, innerRect, &image.Uniform{borderColor}, image.ZP, draw.Src)

					// Bottom
					innerRect = drawRect
					innerRect.Min.Y = innerRect.Max.Y - 2
					draw.Draw(thisImg, innerRect, &image.Uniform{borderColor}, image.ZP, draw.Src)

					// Left
					innerRect = drawRect
					innerRect.Max.X = innerRect.Min.X + 2
					draw.Draw(thisImg, innerRect, &image.Uniform{borderColor}, image.ZP, draw.Src)

					// Right
					innerRect = drawRect
					innerRect.Min.X = innerRect.Max.X - 2
					draw.Draw(thisImg, innerRect, &image.Uniform{borderColor}, image.ZP, draw.Src)
				}

			}
		}

		newDicomNames = append(newDicomNames, strconv.Itoa(i))
		newImageMap[strconv.Itoa(i)] = thisImg
	}

	return newDicomNames, newImageMap, nil
}
