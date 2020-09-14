package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"log"
	"os"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/go-quantize/quantize"
)

type gsData struct {
	path  string
	image image.Image
}

func makeOneGif(pngs []string, outName string, delay int) error {
	outGif, err := MakeOneGIF(pngs, delay)
	if err != nil {
		return err
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	return gif.EncodeAll(f, outGif)
}

// MakeOneGIF creates an animated gif from a (sorted) slice of paths - which may
// be local or Google Storage. The delay (between frames) is in hundredths of a
// second. A delay of 2 seems to be the smallest allowed delay.
func MakeOneGIF(pngs []string, delay int) (*gif.GIF, error) {
	outGif := &gif.GIF{}

	quantizer := quantize.MedianCutQuantizer{
		Aggregation:    quantize.Mean,
		Weighting:      nil,
		AddTransparent: false,
	}

	fetches := make(chan gsData)

	for _, input := range pngs {
		go func(input string) {
			pngImage, err := bulkprocess.MaybeExtractImageFromGoogleStorage(input, client)
			if err != nil {
				log.Println(err, "when operating on", input)
			}
			dat := gsData{
				image: pngImage,
				path:  input,
			}

			fetches <- dat

		}(input)
	}

	pngDats := make(map[string]gsData)
	for range pngs {
		dat := <-fetches
		pngDats[dat.path] = dat
	}

	sortedPngDats := make([]gsData, 0, len(pngs))
	sortedPngs := make([]image.Image, 0, len(sortedPngDats))
	for _, png := range pngs {
		if pngDats[png].image == nil {
			return nil, fmt.Errorf("One or more images could not be loaded")
		}

		sortedPngDats = append(sortedPngDats, pngDats[png])
		sortedPngs = append(sortedPngs, pngDats[png].image)
	}
	pal := quantizer.QuantizeMultiple(make([]color.Color, 0, 256), sortedPngs)

	// Convert each image to a frame in our animated gif
	for _, input := range sortedPngDats {

		img := input.image

		palettedImage := image.NewPaletted(img.Bounds(), pal)
		draw.Draw(palettedImage, img.Bounds(), img, image.Point{}, draw.Over)
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, delay)
	}

	return outGif, nil
}
