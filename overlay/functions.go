package overlay

import (
	"image/color"
	"strconv"
	"strings"
)

func rgbaFromColorCode(colorCode string) (color.Color, error) {
	colorCode = strings.ReplaceAll(colorCode, "#", "")

	// Special case the background
	if len(colorCode) < 6 {
		return color.RGBA{0, 0, 0, 0}, nil
	}

	// Parse each channel
	r, err := strconv.ParseInt(colorCode[0:2], 16, 16)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseInt(colorCode[2:4], 16, 16)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseInt(colorCode[4:6], 16, 16)
	if err != nil {
		return nil, err
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}, nil
}
