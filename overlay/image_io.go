package overlay

import (
	"bytes"
	"image"
	"io/ioutil"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"cloud.google.com/go/storage"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	_ "golang.org/x/image/bmp"
)

// ImageFromBytes creates an image from the specified bytes. Must be PNG, GIF,
// BMP, or JPEG formatted (based on the decoders we have imported).
func ImageFromBytes(imgBytes []byte) (image.Image, error) {
	imgReader := bytes.NewReader(imgBytes)

	// Extract and decode the image.
	img, _, err := image.Decode(imgReader)

	return img, err
}

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

	return ImageFromBytes(imgBytes)
}

// OpenImageFromLocalFile pulls an image with the specified suffix (derived
// from the DICOM name) from a local folder
func OpenImageFromLocalFile(filePath string) (image.Image, error) {
	return OpenImageFromLocalFileOrGoogleStorage(filePath, nil)
}
