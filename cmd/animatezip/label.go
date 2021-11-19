package main

import (
	"image"

	"github.com/fogleman/gg"
)

// addLabelGG writes text on top of an image.
func addLabelGG(img image.Image, label string) image.Image {
	ctx := gg.NewContextForImage(img)
	ctx.SetRGB(1, 1, 1)

	if err := ctx.LoadFontFace("/Library/Fonts/Arial Unicode.ttf", 8); err != nil {
		panic(err)
	}

	ctx.DrawString(label, 0, 8)

	return ctx.Image()
}
