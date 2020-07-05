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

func main() {
	pngs := []string{
		"gs://ml4cvd/jamesp/annotation/flow/v20200702/apply/output/merged_pngs/1.3.12.2.1107.5.2.18.141243.2017042612302443114006926.dcm.png.overlay.png",
	}

	if err := run(pngs); err != nil {
		log.Fatalln(err)
	}

	log.Println("Encoded")
}

func run(pngs []string) error {

	outGif := &gif.GIF{}

	quantizer := quantize.MedianCutQuantizer{
		Aggregation:    quantize.Mode,
		Weighting:      nil,
		AddTransparent: false,
	}
	// quantizer := gogif.MedianCutQuantizer{NumColor: 256}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	// Loop this
	for _, input := range pngs {
		pngReader, err := ImportFileFromGoogleStorage(input, client)
		if err != nil {
			return err
		}

		img, err := png.Decode(pngReader)
		if err != nil {
			return err
		}

		palette := quantizer.Quantize(make([]color.Color, 0, 256), img)

		palettedImage := image.NewPaletted(img.Bounds(), palette)
		draw.Draw(palettedImage, img.Bounds(), img, image.Point{}, draw.Over)
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, 0)
	}

	// save to out.gif
	f, _ := os.OpenFile("out.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, outGif)

	return nil
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

	log.Println(bucketName, pathName)

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
