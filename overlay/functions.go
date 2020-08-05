package overlay

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"strconv"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	_ "golang.org/x/image/bmp"
)

func OpenImageFromLocalFileOrGoogleStorage(filePath string, storageClient *storage.Client) (image.Image, error) {
	f, _, err := bulkprocess.MaybeOpenFromGoogleStorage(filePath, storageClient)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// The image decoder swallows errors, so we won't see i/o errors if they
	// happen during image decoding. To capture these, we read the full image
	// into memory here, and pass a byte reader to the image decoder.

	imgBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	imgReader := bytes.NewReader(imgBytes)

	// Extract and decode the image. Must be PNG, GIF, BMP, or JPEG formatted
	// (based on the decoders we have imported)
	img, _, err := image.Decode(imgReader)

	return img, err
}

// OpenImageFromLocalFile pulls an image with the specified suffix (derived
// from the DICOM name) from a local folder
func OpenImageFromLocalFile(filePath string) (image.Image, error) {
	return OpenImageFromLocalFileOrGoogleStorage(filePath, nil)
}

// LabeledPixelToID converts the label-encoded pixel (e.g., #010101) which is
// alpha-premultiplied into an ID in the range of 0-255
func LabeledPixelToID(c color.Color) (uint32, error) {

	// Find the color channel values for this pixel
	pr, pg, pb, a := c.RGBA()

	// Confirm that we're mapping ID 1 => #010101, etc
	if pr != pg || pg != pb || pr != pb {
		return 0, fmt.Errorf("Encoding expected to have equal values for R, G, and B. Instead, found %d, %d, %d", pr, pg, pb)
	}

	// Create the hex string representation. Since each color channel is
	// "alpha-premultiplied" (https://golang.org/pkg/image/color/#RGBA),
	// we need to divide by alpha (scaling 0-1), then multiplying by
	// 255, to get what we're actually looking for
	pixelID := uint32(math.Round(255 * float64(pr) / float64(a)))

	return pixelID, nil
}

func rgbaFromColorCode(colorCode string) (color.Color, error) {
	colorCode = strings.ReplaceAll(colorCode, "#", "")

	// Special case the background
	if len(colorCode) < 6 {
		return color.RGBA{0, 0, 0, 0}, nil
	}

	// Parse each channel
	r, err := strconv.ParseUint(colorCode[0:2], 16, 8)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseUint(colorCode[2:4], 16, 8)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseUint(colorCode[4:6], 16, 8)
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

func nrgbaFromColorCode(colorCode string) (color.Color, error) {
	colorCode = strings.ReplaceAll(colorCode, "#", "")

	// Special case the background
	if len(colorCode) < 6 {
		return color.RGBA{0, 0, 0, 0}, nil
	}

	// Parse each channel
	r, err := strconv.ParseUint(colorCode[0:2], 16, 8)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseUint(colorCode[2:4], 16, 8)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseUint(colorCode[4:6], 16, 8)
	if err != nil {
		return nil, err
	}

	return color.NRGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}, nil
}
