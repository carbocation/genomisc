package main

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/carbocation/go-quantize/quantize"
)

type gsData struct {
	path  string
	image image.Image
}

func makeOneGifFromImageMap(dicomNames []string, imgMap map[string]image.Image, outName string, delay int) error {

	outGif, err := MakeOneGIFFromMap(dicomNames, imgMap, delay, client)
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

func MakeOneGIFFromMap(dicomNames []string, imgMap map[string]image.Image, delay int, storageClient *storage.Client) (*gif.GIF, error) {
	images := make([]image.Image, 0, len(dicomNames))
	for _, dicomName := range dicomNames {
		images = append(images, imgMap[dicomName])
	}

	outGif, err := MakeOneGIF(images, delay)
	if err != nil {
		return nil, err
	}

	return outGif, nil
}

// FetchImagesFromZIP fetches images from DICOM files within a Zip file. The Zip
// file may be local or on Google Storage. They are returned as
// map[dicom_name]image
func FetchImagesFromZIP(zipPath string, includeOverlay bool, storageClient *storage.Client) (map[string]image.Image, error) {
	pngDats := make(map[string]image.Image)

	// Read the zip file into memory still compressed - either from a local
	// file, or from Google storage, depending on the prefix you provide.
	readerAt, zipNBytes, err := bulkprocess.MaybeOpenFromGoogleStorage(zipPath, storageClient)
	if err != nil {
		return nil, err
	}
	defer readerAt.Close()

	// Since we read all images, just do it all at once
	allBytes, err := ioutil.ReadAll(readerAt)
	if err != nil {
		return nil, err
	}

	zipBytes := bytes.NewReader(allBytes)

	rc, err := zip.NewReader(zipBytes, zipNBytes)
	if err != nil {
		return nil, err
	}

	// Iterate over the files within the Zip
	for _, v := range rc.File {
		// Read the desired DICOMs into images

		dicomReader, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer dicomReader.Close()

		img, err := bulkprocess.ExtractDicomFromReader(dicomReader, includeOverlay)
		if err != nil {
			// If it's not a DICOM file, mention that and move on
			log.Println(err)
			continue
		}
		dicomReader.Close()

		pngDats[v.Name] = img
	}

	return pngDats, nil
}

// MakeOneGIF creates an animated gif from an ordered slice of images. The delay
// between frames is in hundredths of a second. The color quantizer is built
// from *all* input images, and the quantized palette is shared across all of
// the output frames.
func MakeOneGIF(sortedImages []image.Image, delay int) (*gif.GIF, error) {
	outGif := &gif.GIF{}

	quantizer := quantize.MedianCutQuantizer{
		Aggregation:    quantize.Mean,
		Weighting:      nil,
		AddTransparent: false,
	}

	pal := quantizer.QuantizeMultiple(make([]color.Color, 0, 256), sortedImages)

	// Convert each image to a frame in our animated gif
	for _, img := range sortedImages {
		palettedImage := image.NewPaletted(img.Bounds(), pal)
		draw.Draw(palettedImage, img.Bounds(), img, image.Point{}, draw.Over)
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, delay)
	}

	return outGif, nil
}
