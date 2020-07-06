package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/go-quantize/quantize"
	"github.com/carbocation/pfx"
)

type gsData struct {
	path  string
	image image.Image
}

func makeOneGif(pngs []string, outName string, delay int) error {
	outGif := &gif.GIF{}

	quantizer := quantize.MedianCutQuantizer{
		Aggregation:    quantize.Mean,
		Weighting:      nil,
		AddTransparent: false,
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	fetches := make(chan gsData)

	for _, input := range pngs {
		go func(input string) {
			pngImage, err := ImportPNGFromGoogleStorage(input, client)
			if err != nil {
				log.Println(err)
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

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	return gif.EncodeAll(f, outGif)
}

// ImportPNGFromGoogleStorage copies a file from google storage and returns it
// as bytes.
func ImportPNGFromGoogleStorage(gsFilePath string, client *storage.Client) (image.Image, error) {

	// Detect the bucket and the path to the actual file
	pathParts := strings.SplitN(strings.TrimPrefix(gsFilePath, "gs://"), "/", 2)
	if len(pathParts) != 2 {
		return nil, fmt.Errorf("Tried to split your google storage path into 2 parts, but got %d: %v", len(pathParts), pathParts)
	}
	bucketName := pathParts[0]
	pathName := pathParts[1]

	// Open the bucket with default credentials
	bkt := client.Bucket(bucketName)
	handle := bkt.Object(pathName)

	rc, err := handle.NewReader(context.Background())
	if err != nil {
		return nil, pfx.Err(err)
	}
	defer rc.Close()

	img, err := png.Decode(rc)
	if err != nil {
		return nil, err
	}

	return img, pfx.Err(err)
}
