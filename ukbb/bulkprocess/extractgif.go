package bulkprocess

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io/ioutil"
	"log"

	"cloud.google.com/go/storage"
	"github.com/carbocation/go-quantize/quantize"
)

type gsData struct {
	path  string
	image image.Image
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

// MakeOneGIFFromMap creates an animated gif from an ordered list of image names
// along with a map of the respective images.
func MakeOneGIFFromMap(dicomNames []string, imgMap map[string]image.Image, delay int) (*gif.GIF, error) {
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

// MakeOneGIFFromPaths creates an animated gif from an ordered slice of paths to
// image files - which may be local or hosted in an accessible Google Storage
// location. (The string for each png should be a fully specified path.)
func MakeOneGIFFromPaths(pngs []string, delay int, storageClient *storage.Client) (*gif.GIF, error) {
	fetches := make(chan gsData)

	// Download the images in parallel
	for _, input := range pngs {
		go func(input string) {
			pngImage, err := MaybeExtractImageFromGoogleStorage(input, storageClient)
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

// FetchImagesFromZIP fetches images from DICOM files within a Zip file. The Zip
// file may be local or on Google Storage (if path begins with gs://). The
// images are returned in a map, keyed to the DICOM name: map[dicom_name]image
func FetchImagesFromZIP(zipPath string, includeOverlay bool, storageClient *storage.Client) (map[string]image.Image, error) {
	pngDats := make(map[string]image.Image)

	// Read the zip file into memory still compressed - either from a local
	// file, or from Google storage, depending on the prefix you provide.
	readerAt, zipNBytes, err := MaybeOpenFromGoogleStorage(zipPath, storageClient)
	if err != nil {
		return nil, err
	}
	defer readerAt.Close()

	// Since we read all images, just do it all at once for more efficient
	// transfer over the wire.
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

		// Read all DICOMs into images
		dicomReader, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer dicomReader.Close()

		img, err := ExtractDicomFromReader(dicomReader, includeOverlay)
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
