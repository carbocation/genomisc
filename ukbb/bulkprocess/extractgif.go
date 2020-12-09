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
	"runtime"

	"cloud.google.com/go/storage"
	"github.com/carbocation/go-quantize/quantize"
	"github.com/carbocation/pfx"
)

type gsData struct {
	path  string
	image image.Image
}

type orderedPaletted struct {
	key   int
	image *image.Paletted
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
	palettedImages := make(chan orderedPaletted)
	semaphore := make(chan struct{}, runtime.NumCPU())

	// This is surprisingly slow and so is worth parallelizing.
	go func() {
		for k, img := range sortedImages {
			semaphore <- struct{}{}

			go func(k int, img image.Image) {
				defer func() { <-semaphore }()

				palettedImage := image.NewPaletted(img.Bounds(), pal)
				draw.Draw(palettedImage, img.Bounds(), img, image.Point{}, draw.Over)

				palettedImages <- orderedPaletted{
					key:   k,
					image: palettedImage,
				}
			}(k, img)
		}
	}()

	// Save the outputs - in order
	sortedPalattedImages := make([]*image.Paletted, len(sortedImages))
	for range sortedImages {

		palettedImage := <-palettedImages
		sortedPalattedImages[palettedImage.key] = palettedImage.image

	}

	// Assemble into a gif
	for _, palettedImage := range sortedPalattedImages {
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

	sortedPngs, err := FetchGIFComponents(pngs, storageClient)
	if err != nil {
		return nil, err
	}

	// For now, block gif creation if all of the input images don't have the
	// same size. TODO: resize images that don't all share the same bounds.
	lastBounds := sortedPngs[0].Bounds()
	for k, pngImg := range sortedPngs {
		if x := pngImg.Bounds(); x != lastBounds {
			return nil, fmt.Errorf("Image %d (%s) had unexpected bounds (image 0 (%s) bounds: %v, image %d bounds: %v)", k, pngs[k], pngs[0], lastBounds, k, x)
		}
	}

	return MakeOneGIF(sortedPngs, delay)
}

func FetchGIFComponents(pngs []string, storageClient *storage.Client) ([]image.Image, error) {
	fetches := make(chan gsData)

	semaphore := make(chan struct{}, 4*runtime.NumCPU())

	// Download in parallel. Will fetch up to `4*runtime.NumCPU()` images
	// simultaneously.
	go func() {
		for _, input := range pngs {

			// Will block after `4*runtime.NumCPU()` simultaneous downloads are running
			semaphore <- struct{}{}

			go func(input string) {

				// Unblock once finished
				defer func() { <-semaphore }()

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
	}()

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

	// Confirm that we fetched at least one image
	if len(sortedPngs) < 1 {
		return nil, pfx.Err(fmt.Errorf("No images could be fetched from the file paths"))
	}

	return sortedPngs, nil
}

// FetchImagesFromZIP is a deprecated shortcut to FetchNamedImagesFromZIP with
// an empty list of images, which will return all images from a zip file.
func FetchImagesFromZIP(zipPath string, includeOverlay bool, storageClient *storage.Client) (map[string]image.Image, error) {
	return FetchNamedImagesFromZIP(zipPath, includeOverlay, storageClient, nil)
}

// FetchNamedImagesFromZIP fetches images from DICOM files within a Zip file.
// The Zip file may be local or on Google Storage (if path begins with gs://).
// The images are returned in a map, keyed to the DICOM name:
// map[dicom_name]image. If a non-nil slice of filenames is passed, then only
// filenames that appear in this slice will be processed and returned.
func FetchNamedImagesFromZIP(zipPath string, includeOverlay bool, storageClient *storage.Client, acceptedFiles []string) (map[string]image.Image, error) {
	pngDats := make(map[string]image.Image)

	permittedFilenames := make(map[string]struct{})
	for _, v := range acceptedFiles {
		permittedFilenames[v] = struct{}{}
	}

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

		// Don't burn cycles parsing non-permitted filenames
		if _, allowed := permittedFilenames[v.Name]; len(permittedFilenames) > 0 && !allowed {
			continue
		}

		// Read all DICOMs into images
		dicomReader, err := v.Open()
		if err != nil {
			return nil, err
		}
		// We're in a loop so try to keep track of when to close this file this
		// manually rather than relying on defer.
		// defer dicomReader.Close()

		img, err := ExtractDicomFromReader(dicomReader, int64(v.UncompressedSize64), includeOverlay)
		if err != nil {
			// If it's not a DICOM file, mention that and move on
			log.Println(err)
			dicomReader.Close()
			continue
		}

		pngDats[v.Name] = img
		dicomReader.Close()
	}

	return pngDats, nil
}
