package overlay

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/carbocation/pfx"

	"github.com/tj/go-rle"
)

// A Label tracks the segmentation ID with the human-identifiable Label and
// human-interpretable color (in RGB hex, e.g., #FF0000 for red).
type Label struct {
	Label     string
	ID        uint   `json:"id"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order,omitempty"`
}

// LabelMap ([string label name]Label) keeps track of the relationship between
// human-visible colors and the segmentation ID (used for deep learning) of that
// label.
type LabelMap map[string]Label

// EncodeImageToImageSegment consumes a multi-color human-visible image into an
// image where each pixel has the same R, G, and B value mapped to the integer
// ID of the Label. For example, if background is transparent (ID 0) and the
// left atrium is red (ID 1), it will produce an image that is all black, with
// values #000000 for the background and #010101 for the left atrium.
func (l LabelMap) EncodeImageToImageSegment(bmpImage image.Image) (image.Image, error) {

	// Because the HTML5 canvas can be imprecise, we need to map nearby colors
	colorPalette := color.Palette{}
	for _, v := range l.Sorted() {

		// Create this label's color.Color based on its hex representation, and
		// add it to our permitted color palette
		col, err := nrgbaFromColorCode(v.Color)
		if err != nil {
			continue
		}
		colorPalette = append(colorPalette, col)

		// Double check that our palette colors map back to themselves
		r, g, b, _ := col.RGBA()
		cl := fmt.Sprintf("#%02x%02x%02x", uint8(r), uint8(g), uint8(b))

		// Background is special cased to be allowed to mismatch
		if cl != "#000000" && cl != v.Color {
			return nil, fmt.Errorf("Label ID %d (color %s) mapped to %s instead of its own color", v.ID, v.Color, cl)
		}
	}

	// Map from the color code to the Label, with all of its attached
	// information
	colorLabels := make(map[color.Color]Label)

	outputImage := image.NewRGBA(bmpImage.Bounds())

	for y := 0; y < bmpImage.Bounds().Max.Y; y++ {
		for x := 0; x < bmpImage.Bounds().Max.X; x++ {

			// Find the color at this pixel
			rawC := bmpImage.At(x, y)

			// If we haven't yet mapped this point's color to a Label
			// identifier, do so now:
			if _, exists := colorLabels[rawC]; !exists {

				// The exact color chosen by HTML5 may or may not exist, so we
				// will choose the nearest one from our palette:
				c := colorPalette[colorPalette.Index(rawC)]

				// Create the hex string representation
				r, g, b, _ := c.RGBA()
				cl := fmt.Sprintf("#%02x%02x%02x", uint8(r), uint8(g), uint8(b))

				// Look up the hex string representation and map it
				for _, v := range l.Sorted() {
					if v.Color == cl {
						colorLabels[rawC] = v
						break
					} else if cl == "#000000" && v.ID == 0 {
						// Background, special case
						colorLabels[rawC] = v
						break
					}
				}
			}

			// Make sure that all labels are known
			lab, exists := colorLabels[rawC]
			if !exists {
				return nil, fmt.Errorf("Saw color %v but could not find a neighbor for this color in the label map palette", rawC)
			}

			// If
			// var opacity uint8 = 255
			// if lab.ID == 0 {
			// 	opacity = 0
			// }
			outputImage.Set(x, y, color.RGBA{
				R: uint8(lab.ID),
				G: uint8(lab.ID),
				B: uint8(lab.ID),
				A: 255,
			})
		}
	}

	return outputImage, nil
}

// DecodeImageFromImageSegment consumes an ID-encoded image (where each pixel is
// #010101 for ID 1, #020202 for ID 2 etc), and transforms it into a
// human-visible value based on the colors for those IDs assigned in the config
// file. It special-cases ID 0 (the background) to be transparent.
func (l LabelMap) DecodeImageFromImageSegment(bmpImage image.Image) (image.Image, error) {
	// Map from the color code to the Label, with all of its attached
	// information
	colorLabels := make(map[uint32]Label)

	outputImage := image.NewRGBA(bmpImage.Bounds())

	for y := 0; y < bmpImage.Bounds().Max.Y; y++ {
		for x := 0; x < bmpImage.Bounds().Max.X; x++ {

			// Find the color at this pixel
			c := bmpImage.At(x, y)

			// Find the color channel values for this pixel
			pr, pg, pb, a := c.RGBA()

			// Confirm that we're mapping ID 1 => #010101, etc
			if pr != pg || pg != pb || pr != pb {
				return nil, fmt.Errorf("Encoding expected to have equal values for R, G, and B. Instead, found %d, %d, %d", pr, pg, pb)
			}

			// Create the hex string representation. Since each color channel is
			// "alpha-premultiplied" (https://golang.org/pkg/image/color/#RGBA),
			// we need to divide by alpha (scaling 0-1), then multiplying by
			// 255, to get what we're actually looking for
			pixelID := uint32(math.Round(255 * float64(pr) / float64(a)))

			// If we haven't yet mapped this point's color to a Label
			// identifier, do so now:
			if _, exists := colorLabels[pixelID]; !exists {

				// Look up the ID
				for _, v := range l.Sorted() {
					if uint32(v.ID) == pixelID {
						colorLabels[pixelID] = v
						break
					} else if pixelID == 0 {
						// Background, special case
						colorLabels[pixelID] = v
						break
					}
				}
			}

			// Make sure that all labels are known
			lab, exists := colorLabels[pixelID]
			if !exists {
				return nil, pfx.Err(fmt.Errorf("Saw color %v (ID %v) but could not find this color in the label map %+v", c, pixelID, l))
			}

			// For human vision, we want the background to be special-cased to
			// transparent, and the pixels otherwise to be fully opaque.
			var opacity uint8 = 255
			if lab.ID == 0 {
				opacity = 0
			}
			humanColor, err := rgbaFromColorCode(lab.Color)
			if err != nil {
				return nil, err
			}

			hr, hg, hb, _ := humanColor.RGBA()

			outputImage.Set(x, y, color.RGBA{
				R: uint8(hr),
				G: uint8(hg),
				B: uint8(hb),
				A: opacity,
			})
		}
	}

	return outputImage, nil
}

func (l LabelMap) EncodeImageToRLE(bmpImage image.Image) ([]byte, error) {
	// Map from the color code to the Label, with all of its attached
	// information
	colorLabels := make(map[color.Color]Label)
	pixelLabels := make([]int64, 0)

	for y := 0; y < bmpImage.Bounds().Max.Y; y++ {
		for x := 0; x < bmpImage.Bounds().Max.X; x++ {

			// Find the color at this pixel
			c := bmpImage.At(x, y)

			// If we haven't yet mapped this point's color to a Label
			// identifier, do so now:
			if _, exists := colorLabels[c]; !exists {

				// Create the hex string representation
				r, g, b, _ := c.RGBA()
				cl := fmt.Sprintf("#%02x%02x%02x", uint8(r), uint8(g), uint8(b))

				// Look up the hex string representation and map it
				for _, v := range l.Sorted() {
					if v.Color == cl {
						colorLabels[c] = v
						break
					} else if cl == "#000000" && v.ID == 0 {
						// Background, special case
						colorLabels[c] = v
						break
					}
				}
			}

			// Make sure that all labels are known
			lab, exists := colorLabels[c]
			if !exists {
				return nil, fmt.Errorf("Saw color %v but could not find this color in the label map", c)
			}
			pixelLabels = append(pixelLabels, int64(lab.ID))
		}
	}

	encoded := rle.EncodeInt64(pixelLabels)

	return encoded, nil
}

func (l LabelMap) DecodeImageFromRLE(rleBytes []byte, maxX, maxY int) (image.Image, error) {
	slc, err := rle.DecodeInt64(rleBytes)
	if err != nil {
		return nil, err
	}

	// Know which label maps to each integer
	labelColors := make(map[int64]Label)
	for _, v := range l {
		labelColors[int64(v.ID)] = v
	}

	// Know which color maps to each entry
	colorCodes := make([]string, 0, len(slc))
	for _, v := range slc {
		colorCodes = append(colorCodes, labelColors[v].Color)
	}

	// Paint each pixel
	img := image.NewRGBA(image.Rect(0, 0, maxX, maxY))

	for i, label := range slc {
		colorCode := labelColors[label].Color

		colHere, err := rgbaFromColorCode(colorCode)
		if err != nil {
			return nil, err
		}

		// img.Set(i/maxY, i%maxY, colHere)
		img.Set(i%maxX, i/maxX, colHere)
	}

	return img, nil
}

// Valid ensures that the LabelMap is valid by testing that it is bijective. If
// not, it's invalid. (Previously checked to make sure that IDs started at 0 and
// had no gaps, but that requirement has since been relaxed.)
func (l LabelMap) Valid() bool {
	inverse := make(map[uint]string)
	for k, v := range l {
		inverse[v.ID] = k
	}

	// Bijective?
	if !(len(l) == len(inverse)) {
		return false
	}

	// No longer checking that it starts with 0 and has consecutive integers
	// for i := 0; i < len(inverse); i++ {
	// 	if _, exists := inverse[uint(i)]; !exists {
	// 		return false
	// 	}
	// }

	return true
}

func (l LabelMap) Sorted() []Label {
	out := make([]Label, 0, len(l))

	for k, v := range l {
		v.Label = k
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool {
		// If SortOrder is defined and different, use it:
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}

		// If SortOrder is not defined, or is the same for two values, drop down
		// to the ID field for sorting
		return out[i].ID < out[j].ID

	})

	return out
}
