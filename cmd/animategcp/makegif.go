package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/carbocation/pfx"
	"github.com/ericpauley/go-quantize/quantize"
)

type gsData struct {
	path   string
	reader *bytes.Reader
}

func makeOneGif(pngs []string, outName string) error {
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
			pngReader, err := ImportFileFromGoogleStorage(input, client)
			if err != nil {
				log.Println(err)
			}
			dat := gsData{
				reader: pngReader,
				path:   input,
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
	for _, png := range pngs {
		sortedPngDats = append(sortedPngDats, pngDats[png])
	}

	// Loop this
	var palette color.Palette
	for i, input := range sortedPngDats {

		img, err := png.Decode(input.reader)
		if err != nil {
			return err
		}

		if i == 0 {
			palette = quantizer.Quantize(make([]color.Color, 0, 256), img)
		}

		palettedImage := image.NewPaletted(img.Bounds(), palette)
		draw.Draw(palettedImage, img.Bounds(), img, image.Point{}, draw.Over)
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, 5)
	}

	// Save file
	f, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	return gif.EncodeAll(f, outGif)
}

// ImportFileFromGoogleStorage copies a file from google storage and returns it
// as bytes.
func ImportFileFromGoogleStorage(gsFilePath string, client *storage.Client) (*bytes.Reader, error) {

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

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, pfx.Err(err)
	}

	br := bytes.NewReader(data)

	return br, pfx.Err(err)
}
