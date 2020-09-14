package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
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

func makeOneGif(pngs []string, outName string, delay int) error {
	outGif, err := MakeOneGIFFromPaths(pngs, delay, client)
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

func makeOneGifFromZip(dicomNames []string, zipPath string, outName string, delay int) error {
	outGif, err := MakeOneGIFFromZIP(dicomNames, zipPath, delay, client)
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

func MakeOneGIFFromZIP(dicomNames []string, zipPath string, delay int, storageClient *storage.Client) (*gif.GIF, error) {
	images, err := FetchImagesFromZIPByDICOMNames(dicomNames, zipPath, false, storageClient)
	if err != nil {
		return nil, err
	}

	outGif, err := MakeOneGIF(images, delay)
	if err != nil {
		return nil, err
	}

	return outGif, nil
}

// FetchImagesFromZIPByDICOMNames fetches images from DICOM files within a Zip
// file. The Zip file may be local or on Google Storage.
func FetchImagesFromZIPByDICOMNames(dicomNames []string, zipPath string, includeOverlay bool, storageClient *storage.Client) ([]image.Image, error) {

	dicomNameMap := make(map[string]struct{})
	for _, dicomName := range dicomNames {
		dicomNameMap[dicomName] = struct{}{}
	}

	pngDats := make(map[string]image.Image)

	// Read the zip file into memory still compressed - either from a local
	// file, or from Google storage, depending on the prefix you provide.
	readerAt, zipNBytes, err := bulkprocess.MaybeOpenFromGoogleStorage(zipPath, storageClient)
	if err != nil {
		return nil, err
	}
	defer readerAt.Close()

	rc, err := zip.NewReader(readerAt, zipNBytes)
	if err != nil {
		return nil, err
	}

	// Iterate over the files within the Zip
	for _, v := range rc.File {
		if _, exists := dicomNameMap[v.Name]; !exists {
			continue
		}

		// Read the desired DICOMs into images

		dicomReader, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer dicomReader.Close()

		img, err := bulkprocess.ExtractDicomFromReader(dicomReader, includeOverlay)
		if err != nil {
			return nil, err
		}
		dicomReader.Close()

		pngDats[v.Name] = img
	}

	// Make sure that the images are in the same sort order as the original
	// input slice
	out := make([]image.Image, 0, len(dicomNames))
	for _, dicom := range dicomNames {
		out = append(out, pngDats[dicom])
	}

	return out, nil
}

// MakeOneGIFFromPaths creates an animated gif from an ordered slice of paths to
// image files - which may be local or hosted in an accessible Google Storage
// location. (The string for each png should be a fully specified path.)
func MakeOneGIFFromPaths(pngs []string, delay int, storageClient *storage.Client) (*gif.GIF, error) {
	fetches := make(chan gsData)

	for _, input := range pngs {
		go func(input string) {
			pngImage, err := bulkprocess.MaybeExtractImageFromGoogleStorage(input, storageClient)
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

	sortedPngs := make([]image.Image, 0, len(pngDats))
	for _, png := range pngs {
		if pngDats[png].image == nil {
			return nil, fmt.Errorf("One or more images could not be loaded")
		}

		sortedPngs = append(sortedPngs, pngDats[png].image)
	}

	return MakeOneGIF(sortedPngs, delay)
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
